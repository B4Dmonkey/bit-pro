package claude

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestWriteSettings_MergesIntoExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	existing := []byte(`{"model": "opus", "enabledPlugins": {"go@go-skills": true}}`)
	if err := os.WriteFile(path, existing, 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
	}

	if err := WriteSettings(path); err != nil {
		t.Fatalf("WriteSettings(%q) returned error: %v", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
	}

	var merged struct {
		Model                  string                     `json:"model"`
		ExtraKnownMarketplaces map[string]json.RawMessage `json:"extraKnownMarketplaces"`
		EnabledPlugins         map[string]bool            `json:"enabledPlugins"`
	}
	if err := json.Unmarshal(data, &merged); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if merged.Model != "opus" {
		t.Errorf("model = %q, want %q", merged.Model, "opus")
	}

	if len(merged.EnabledPlugins) != 2 {
		t.Errorf("enabledPlugins = %v, want 2 entries", merged.EnabledPlugins)
	}

	if !merged.EnabledPlugins["go@go-skills"] {
		t.Errorf("enabledPlugins = %v, want go@go-skills to be true", merged.EnabledPlugins)
	}

	if !merged.EnabledPlugins["bit@bit-pro"] {
		t.Errorf("enabledPlugins = %v, want bit@bit-pro to be true", merged.EnabledPlugins)
	}

	if _, ok := merged.ExtraKnownMarketplaces["bit-pro"]; !ok {
		t.Errorf("extraKnownMarketplaces = %v, want a bit-pro entry", merged.ExtraKnownMarketplaces)
	}
}

func TestWriteSettings_RejectsUnparseableFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	broken := []byte(`{ not json`)
	if err := os.WriteFile(path, broken, 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
	}

	err := WriteSettings(path)
	if err == nil {
		t.Fatal("WriteSettings returned nil error, want non-nil")
	}

	if !strings.Contains(err.Error(), path) {
		t.Errorf("error = %q, want it to name %q", err, path)
	}

	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", path, readErr)
	}

	if !bytes.Equal(data, broken) {
		t.Errorf("settings.json = %s, want it unchanged as %s", data, broken)
	}
}

func TestWriteSettings_IsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	if err := WriteSettings(path); err != nil {
		t.Fatalf("first WriteSettings(%q) returned error: %v", path, err)
	}

	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
	}

	if err := WriteSettings(path); err != nil {
		t.Fatalf("second WriteSettings(%q) returned error: %v", path, err)
	}

	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
	}

	if !bytes.Equal(first, second) {
		t.Errorf("second write = %s, want it identical to the first %s", second, first)
	}
}
