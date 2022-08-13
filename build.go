package main

import (
	"fmt"

	"github.com/code-game-project/go-utils/cgfile"
	"github.com/code-game-project/go-utils/modules"
)

func Build() error {
	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}

	data, err := modules.ReadCommandConfig[modules.BuildData]()
	if err != nil {
		return err
	}

	switch config.Type {
	case "client":
		return buildClient(config.Game, data.Output, config.URL)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func buildClient(gameName, output, url string) error {
	return nil
}
