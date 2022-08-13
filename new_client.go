package main

import (
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/cggenevents"
	"github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/modules"
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
		installLibArgs = append(installLibArgs, "--version")
		installLibArgs = append(installLibArgs, data.LibraryVersion)
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
