package claude

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	pluginKey       = "bit@bit-pro"
	marketplaceName = "bit-pro"
	userScope       = "user"
)

func InstalledVersion(home string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(home, ".claude", "plugins", "installed_plugins.json"))
	if err != nil {
		return "", false
	}

	var record struct {
		Plugins map[string][]struct {
			Scope   string `json:"scope"`
			Version string `json:"version"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(data, &record); err != nil {
		return "", false
	}

	for _, install := range record.Plugins[pluginKey] {
		if install.Scope == userScope {
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

func PluginState(home string) (installed, latest string, ok bool) {
	installed, ok = InstalledVersion(home)
	if !ok {
		return "", "", false
	}

	latest, ok = LatestVersion(home)
	if !ok {
		return "", "", false
	}

	return installed, latest, true
}

func RefreshMarketplace() {
	_ = start("claude", "plugin", "marketplace", "update", marketplaceName)
}

func start(name string, args ...string) error {
	_ = exec.Command(name, args...).Start()

	return nil
}
