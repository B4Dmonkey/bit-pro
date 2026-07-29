package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSettings_CreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".claude", "settings.json")

	if err := WriteSettings(path); err != nil {
		t.Fatalf("WriteSettings(%q) returned error: %v", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
	}

	var settings struct {
		ExtraKnownMarketplaces map[string]struct {
			Source struct {
				Source string `json:"source"`
				Repo   string `json:"repo"`
			} `json:"source"`
		} `json:"extraKnownMarketplaces"`
		EnabledPlugins map[string]bool `json:"enabledPlugins"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	market, ok := settings.ExtraKnownMarketplaces["bit-pro"]
	if !ok {
		t.Fatalf("extraKnownMarketplaces = %v, want a bit-pro entry", settings.ExtraKnownMarketplaces)
	}
	if market.Source.Source != "github" {
		t.Errorf("bit-pro source.source = %q, want %q", market.Source.Source, "github")
	}
	if market.Source.Repo != "B4Dmonkey/bit-pro" {
		t.Errorf("bit-pro source.repo = %q, want %q", market.Source.Repo, "B4Dmonkey/bit-pro")
	}
	if !settings.EnabledPlugins["bit@bit-pro"] {
		t.Errorf("enabledPlugins = %v, want bit@bit-pro to be true", settings.EnabledPlugins)
	}
}
