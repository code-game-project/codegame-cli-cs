package main

import (
	"os"
	"path/filepath"
	"text/template"
)

func ExecTemplate(templateText, path string, data any) error {
	err := os.MkdirAll(filepath.Join(filepath.Dir(path)), 0755)
	if err != nil {
		return err
	}

	tmpl, err := template.New(path).Parse(templateText)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(path))
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}
