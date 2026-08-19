package store

import (
	"fmt"
	"os"
	"path/filepath"
)

const dirMode = 0o755

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locating the home directory: %w", err)
	}

	dir := filepath.Join(home, ".local", "share", "bit-pro")
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return "", fmt.Errorf("creating %s: %w", dir, err)
	}

	return dir, nil
}
