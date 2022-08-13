package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Bananenpro/cli"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "USAGE: %s <command> [...]\n", os.Args[0])
		os.Exit(1)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	projectName := filepath.Base(workingDir)

	switch os.Args[1] {
	case "new":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "USAGE: %s new client\n", os.Args[0])
			os.Exit(1)
		}
		switch os.Args[2] {
		case "client":
			err = CreateNewClient(projectName)
		default:
			fmt.Fprintln(os.Stderr, "Unknown project type:", os.Args[2])
			os.Exit(1)
		}
	case "update":
		err = Update(projectName)
	case "run":
		err = Run()
	case "build":
		err = Build()
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", os.Args[1])
		os.Exit(1)
	}
	if err != nil {
		if err != cli.ErrCanceled {
			cli.Error(err.Error())
		}
		os.Exit(1)
	}
}
