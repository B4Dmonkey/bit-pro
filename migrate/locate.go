package migrate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/B4Dmonkey/bit-pro/project"
)

var ErrNoBitDir = errors.New("no .bit/ here or above")

func source(dir string) (src, path string, err error) {
	holder, err := locate(dir)
	if err != nil {
		return "", "", err
	}

	path, err = project.CanonicalPath(holder)
	if err != nil {
		return "", "", err
	}

	return filepath.Join(holder, ".bit"), path, nil
}

func locate(dir string) (string, error) {
	if root, ok := project.MainCheckout(dir); ok {
		if hasBitDir(root) {
			return root, nil
		}

		return "", fmt.Errorf("%s: %w", root, ErrNoBitDir)
	}

	for d := dir; ; {
		if hasBitDir(d) {
			return d, nil
		}

		parent := filepath.Dir(d)
		if parent == d {
			return "", fmt.Errorf("%s: %w", dir, ErrNoBitDir)
		}

		d = parent
	}
}

func hasBitDir(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".bit"))

	return err == nil && info.IsDir()
}
