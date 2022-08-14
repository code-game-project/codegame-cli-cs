package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/code-game-project/go-utils/cgfile"
	"github.com/code-game-project/go-utils/external"
	"github.com/code-game-project/go-utils/modules"
)

func Run() error {
	config, err := cgfile.LoadCodeGameFile("")
	if err != nil {
		return err
	}

	data, err := modules.ReadCommandConfig[modules.RunData]()
	if err != nil {
		return err
	}

	url := external.TrimURL(config.URL)

	switch config.Type {
	case "client":
		return runClient(url, data.Args)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func runClient(url string, args []string) error {
	cmdArgs := []string{"run", "--no-self-contained", "--"}
	cmdArgs = append(cmdArgs, args...)

	env := []string{"CG_GAME_URL=" + url}
	env = append(env, os.Environ()...)

	if _, err := exec.LookPath("dotnet"); err != nil {
		return fmt.Errorf("'dotnet' ist not installed!")
	}

	cmd := exec.Command("dotnet", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run 'CG_GAME_URL=%s dotnet %s'", url, strings.Join(cmdArgs, " "))
	}
	return nil
}
