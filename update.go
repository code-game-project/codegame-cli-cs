package main

import (
	"fmt"

	"github.com/code-game-project/go-utils/cgfile"
	"github.com/code-game-project/go-utils/modules"
)

func Update() error {
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
		return updateClient(data.LibraryVersion, config)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func updateClient(libraryVersion string, config *cgfile.CodeGameFileData) error {
	return nil
}

func updateClientTemplate(projectName, gameName, displayName, description string, eventNames, commandNames []string) error {
	return execClientTemplate(projectName, gameName, displayName, description, eventNames, commandNames, false)
}
