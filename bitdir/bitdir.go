package bitdir

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	claudeDir    = ".claude"
	worktreesDir = "worktrees"
	defaultDir   = ".bit"
)

var current = defaultDir

func Current() string {
	return current
}

func Resolve() {
	current = defaultDir
	if wd, err := os.Getwd(); err == nil {
		current = Canonical(wd)
	}
}

func Canonical(wd string) string {
	sep := string(filepath.Separator)
	segments := strings.Split(wd, sep)

	for i := 0; i+1 < len(segments); i++ {
		if segments[i] == claudeDir && segments[i+1] == worktreesDir {
			return filepath.Join(strings.Join(segments[:i], sep), defaultDir)
		}
	}

	return defaultDir
}
