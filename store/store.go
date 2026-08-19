package store

import (
	"fmt"
	"os"
	"path/filepath"
)

const dirMode = 0o755

func Dir() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("locating the home directory: %w", err)
		}

		base = filepath.Join(home, ".local", "share")
	}

	dir := filepath.Join(filepath.Clean(base), "bit-pro")
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return "", fmt.Errorf("creating %s: %w", dir, err)
	}

	return dir, nil
}
