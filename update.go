package main

import (
	"fmt"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/cgfile"
	"github.com/code-game-project/go-utils/cggenevents"
	"github.com/code-game-project/go-utils/exec"
	"github.com/code-game-project/go-utils/modules"
	"github.com/code-game-project/go-utils/server"
)

func Update(projectName string) error {
	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}

	data, err := modules.ReadCommandConfig[modules.UpdateData]()
	if err != nil {
		return err
	}
	switch config.Type {
	case "client":
		return updateClient(projectName, data.LibraryVersion, config)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func updateClient(projectName, libraryVersion string, config *cgfile.CodeGameFileData) error {
	api, err := server.NewAPI(config.URL)
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

	err = updateClientTemplate(projectName, config.Game, info.DisplayName, info.Description, eventNames, commandNames)
	if err != nil {
		return err
	}

	cli.BeginLoading("Updating csharp-client...")
	installLibArgs := []string{"add", "package", "CodeGame.Client"}
	if libraryVersion != "latest" {
		libraryVersion, err = nugetVersion("CodeGame.Client", libraryVersion)
		if err != nil {
			return err
		}
		installLibArgs = append(installLibArgs, "--version", libraryVersion)
	}
	_, err = exec.Execute(true, "dotnet", installLibArgs...)
	if err != nil {
		return err
	}
	cli.FinishLoading()

	cli.BeginLoading("Updating dependencies...")
	_, err = exec.Execute(true, "dotnet", "add", "package", "System.CommandLine", "--version", "2.0.0-beta4.22272.1")
	if err != nil {
		return err
	}
	cli.FinishLoading()
	return nil
}

func updateClientTemplate(projectName, gameName, displayName, description string, eventNames, commandNames []string) error {
	return execClientTemplate(projectName, gameName, displayName, description, eventNames, commandNames, true)
}
