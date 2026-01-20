package main

import (
	"os"
	"path/filepath"
)

// getTemplatePath returns the full path to a template file
func getTemplatePath(filename string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "prom", "templates", filename), nil
}
