package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "embed"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/cggenevents"
	"github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/external"
	"github.com/code-game-project/go-utils/modules"
	"github.com/code-game-project/go-utils/semver"
	"github.com/code-game-project/go-utils/server"
)

//go:embed templates/new/client/Program.cs.tmpl
var clientProgramTemplate string

//go:embed templates/new/client/Game.cs.tmpl
var clientGameTemplate string

//go:embed templates/new/client/Events.cs.tmpl
var clientEventsTemplate string

//go:embed templates/new/client/csproj.tmpl
var clientCSProjTemplate string

//go:embed templates/new/gitignore.tmpl
var gitignoreTemplate string

func CreateNewClient(projectName string) error {
	data, err := modules.ReadCommandConfig[modules.NewClientData]()
	if err != nil {
		return err
	}

	api, err := server.NewAPI(data.URL)
	if err != nil {
		return err
	}

	info, err := api.FetchGameInfo()
	if err != nil {
		return err
	}
	if info.DisplayName == "" {
		info.DisplayName = info.Name
	}

	cge, err := api.GetCGEFile()
	if err != nil {
		return err
	}
	cgeVersion, err := cggenevents.ParseCGEVersion(cge)
	if err != nil {
		return err
	}

	eventNames, commandNames, err := cggenevents.GetEventNames(api.BaseURL(), cgeVersion)
	if err != nil {
		return err
	}

	err = createClientTemplate(projectName, data.Name, info.DisplayName, info.Description, eventNames, commandNames)
	if err != nil {
		return err
	}

	cli.BeginLoading("Installing csharp-client...")
	installLibArgs := []string{"add", "package", "CodeGame.Client"}
	if data.LibraryVersion != "latest" {
		data.LibraryVersion, err = nugetVersion("CodeGame.Client", data.LibraryVersion)
		if err != nil {
			return err
		}
		installLibArgs = append(installLibArgs, "--version", data.LibraryVersion)
	}
	_, err = exec.Execute(true, "dotnet", installLibArgs...)
	if err != nil {
		return err
	}
	cli.FinishLoading()

	cli.BeginLoading("Installing dependencies...")
	_, err = exec.Execute(true, "dotnet", "add", "package", "System.CommandLine", "--version", "2.0.0-beta4.22272.1")
	if err != nil {
		return err
	}
	cli.FinishLoading()
	return nil
}

func createClientTemplate(projectName, gameName, displayName, description string, eventNames, commandNames []string) error {
	return execClientTemplate(projectName, gameName, displayName, description, eventNames, commandNames, false)
}

func execClientTemplate(projectName, gameName, displayName, description string, eventNames, commandNames []string, update bool) error {
	gameDir := toPascal(gameName)
	if update {
		cli.Warn("This action will ERASE and regenerate ALL files in '%s/'.\nYou will have to manually update your code to work with the new version.", gameDir)
		ok, err := cli.YesNo("Continue?", false)
		if err != nil || !ok {
			return cli.ErrCanceled
		}
		os.RemoveAll(gameDir)
	} else {
		cli.Warn("DO NOT EDIT the `%s/` directory inside of the project. ALL CHANGES WILL BE LOST when running `codegame update`.", gameDir)
	}

	type event struct {
		Name       string
		PascalName string
	}

	events := make([]event, len(eventNames))
	for i, e := range eventNames {
		events[i] = event{
			Name:       e,
			PascalName: toPascal(e),
		}
	}

	commands := make([]event, len(commandNames))
	for i, c := range commandNames {
		commands[i] = event{
			Name:       c,
			PascalName: toPascal(c),
		}
	}

	data := struct {
		ProjectName    string
		GameNamePascal string
		DisplayName    string
		Description    string
		Events         []event
		Commands       []event
	}{
		ProjectName:    projectName,
		GameNamePascal: toPascal(gameName),
		DisplayName:    displayName,
		Description:    description,
		Events:         events,
		Commands:       commands,
	}

	if !update {
		err := ExecTemplate(clientProgramTemplate, "Program.cs", data)
		if err != nil {
			return err
		}
		err = ExecTemplate(clientCSProjTemplate, projectName+".csproj", data)
		if err != nil {
			return err
		}
		err = ExecTemplate(gitignoreTemplate, ".gitignore", data)
		if err != nil {
			return err
		}
	}

	err := ExecTemplate(clientGameTemplate, filepath.Join(gameDir, "Game.cs"), data)
	if err != nil {
		return err
	}

	err = ExecTemplate(clientEventsTemplate, filepath.Join(gameDir, "Events.cs"), data)
	if err != nil {
		return err
	}

	return nil
}

func toPascal(text string) string {
	text = strings.ReplaceAll(text, "_", " ")
	text = strings.ReplaceAll(text, "-", " ")
	text = strings.Title(text)
	text = strings.ReplaceAll(text, " ", "")
	return text
}

func nugetVersion(pkg, version string) (string, error) {
	res, err := http.Get(fmt.Sprintf("https://api.nuget.org/v3/registration5-gz-semver2/%s/index.json", strings.ToLower(pkg)))
	if err != nil || res.StatusCode != http.StatusOK || !external.HasContentType(res.Header, "application/json") {
		return "", fmt.Errorf("Couldn't access version information from 'https://api.nuget.org/v3/registration5-gz-semver2/%s/index.json'.", strings.ToLower(pkg))
	}
	defer res.Body.Close()
	type response struct {
		Items []struct {
			Items []struct {
				Entry struct {
					Version string `json:"version"`
				} `json:"catalogEntry"`
			} `json:"items"`
		} `json:"items"`
	}
	var data response
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return "", fmt.Errorf("Couldn't decode nuget version data: %s", err)
	}

	versions := make([]string, 0, len(data.Items[0].Items))

	for _, item := range data.Items[0].Items {
		if strings.HasPrefix(item.Entry.Version, version) {
			versions = append(versions, item.Entry.Version)
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		a := versions[i]
		b := versions[j]

		a1, a2, a3, err := semver.ParseVersion(a)
		if err != nil {
			return false
		}
		b1, b2, b3, err := semver.ParseVersion(b)
		if err != nil {
			return false
		}

		return a1 > b1 || (a1 == b1 && a2 > b2) || (a1 == b1 && a2 == b2 && a3 > b3)
	})

	if len(versions) > 0 {
		return versions[0], nil
	}

	return "", fmt.Errorf("Couldn't fetch the correct library package version to use.")
}
