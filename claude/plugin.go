package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func InstalledVersion(home, projectRoot string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(home, ".claude", "plugins", "installed_plugins.json"))
	if err != nil {
		return "", false
	}

	var record struct {
		Plugins map[string][]struct {
			ProjectPath string `json:"projectPath"`
			Version     string `json:"version"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(data, &record); err != nil {
		return "", false
	}

	want := filepath.Clean(projectRoot)
	for _, install := range record.Plugins[pluginKey] {
		if install.ProjectPath != "" && filepath.Clean(install.ProjectPath) == want {
			return install.Version, true
		}
	}

	return "", false
}

func LatestVersion(home string) (string, bool) {
	path := filepath.Join(home, ".claude", "plugins", "marketplaces", marketplaceName,
		"bit", ".claude-plugin", "plugin.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	var manifest struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", false
	}

	if manifest.Version == "" {
		return "", false
	}

	return manifest.Version, true
}
