package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Bananenpro/cli"
	"github.com/code-game-project/go-utils/cgfile"
	"github.com/code-game-project/go-utils/exec"
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
	data.OS = strings.ReplaceAll(data.OS, "macos", "osx")
	data.OS = strings.ReplaceAll(data.OS, "windows", "win")
	data.Arch = strings.ReplaceAll(data.Arch, "arm32", "arm")

	switch config.Type {
	case "client":
		return buildClient(config.Game, data.Output, config.URL, data.OS, data.Arch)
	default:
		return fmt.Errorf("Unknown project type: %s", config.Type)
	}
}

func buildClient(gameName, output, url, os, arch string) (err error) {
	cli.BeginLoading("Building...")
	gameDir := toPascal(gameName)
	err = replaceInFile(filepath.Join(gameDir, "Game.cs"), "throw new InvalidOperationException(\"The CG_GAME_URL environment variable must be set.\")", "return \""+url+"\"")
	if err != nil {
		return err
	}
	defer func() {
		err2 := replaceInFile(filepath.Join(gameDir, "Game.cs"), "return \""+url+"\"", "throw new InvalidOperationException(\"The CG_GAME_URL environment variable must be set.\")")
		if err == nil && err2 != nil {
			err = err2
		}
	}()

	args := []string{"publish", "--nologo", "--configuration", "Release", "--self-contained"}
	if os == "current" {
		os = getOS()
	}
	if arch == "current" {
		arch = getArch()
	}
	args = append(args, "--runtime", os+"-"+arch)
	if output != "" {
		args = append(args, "--output", output)
	}
	_, err = exec.Execute(true, "dotnet", args...)
	if err != nil {
		return err
	}

	cli.FinishLoading()

	return nil
}

func replaceInFile(filename, old, new string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Failed to replace '%s' with '%s' in '%s': %s", old, new, filename, err)
	}
	content = []byte(strings.ReplaceAll(string(content), old, new))
	err = os.WriteFile(filename, content, 0o644)
	if err != nil {
		return fmt.Errorf("Failed to replace '%s' with '%s' in '%s': %s", old, new, filename, err)
	}
	return nil
}

func getOS() string {
	os := runtime.GOOS
	switch os {
	case "linux":
		return "linux"
	case "darwin":
		return "osx"
	case "windows":
		return "win"
	}
	return ""
}

func getArch() string {
	os := runtime.GOARCH
	switch os {
	case "amd64":
		return "x64"
	case "386":
		return "x86"
	case "arm64":
		return "arm64"
	case "arm":
		return "arm"
	}
	return ""
}
