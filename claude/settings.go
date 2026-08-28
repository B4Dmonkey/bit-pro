package claude

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const pluginKey = "bit@bit-pro"

var marketplace = json.RawMessage(`{"source": {"source": "github", "repo": "B4Dmonkey/bit-pro"}}`)

func WriteSettings(path string) error {
	doc, err := load(path)
	if err != nil {
		return err
	}

	if err := merge(doc, "extraKnownMarketplaces", "bit-pro", marketplace); err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	if err := merge(doc, "enabledPlugins", pluginKey, json.RawMessage(`true`)); err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

func load(path string) (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]json.RawMessage{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	doc := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return doc, nil
}

func merge(doc map[string]json.RawMessage, section, key string, value json.RawMessage) error {
	entries := map[string]json.RawMessage{}
	if raw, ok := doc[section]; ok {
		if err := json.Unmarshal(raw, &entries); err != nil {
			return fmt.Errorf("decoding %s: %w", section, err)
		}
	}

	entries[key] = value

	merged, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", section, err)
	}

	doc[section] = merged

	return nil
}
