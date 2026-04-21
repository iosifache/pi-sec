package paths

import (
	"os"
	"path/filepath"
)

func DataDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return filepath.Join("..", "data")
	}

	if filepath.Base(cwd) == "sync" {
		return filepath.Clean(filepath.Join(cwd, "..", "data"))
	}

	if info, err := os.Stat(filepath.Join(cwd, "sync")); err == nil && info.IsDir() {
		return filepath.Join(cwd, "data")
	}

	return filepath.Join(cwd, "data")
}

func NPMDataDir() string {
	return filepath.Join(DataDir(), "npm-data")
}

func GitHubDataDir() string {
	return filepath.Join(DataDir(), "github")
}

func AlertsFile() string {
	return filepath.Join(DataDir(), "alerts.json")
}
