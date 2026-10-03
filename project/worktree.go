package project

import (
	"path/filepath"
	"strings"
)

const (
	claudeDir    = ".claude"
	worktreesDir = "worktrees"
)

func MainCheckout(dir string) (string, bool) {
	sep := string(filepath.Separator)
	segments := strings.Split(dir, sep)

	for i := 0; i+1 < len(segments); i++ {
		if segments[i] == claudeDir && segments[i+1] == worktreesDir {
			return strings.Join(segments[:i], sep), true
		}
	}

	return "", false
}
