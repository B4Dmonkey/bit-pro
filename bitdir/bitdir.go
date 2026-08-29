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

func Root() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	if dir, ok := worktreeCut(wd); ok {
		return filepath.Dir(dir)
	}

	return wd
}

func Canonical(wd string) string {
	if dir, ok := worktreeCut(wd); ok {
		return dir
	}

	return defaultDir
}

func ForRoot(root string) string {
	if dir, ok := worktreeCut(root); ok {
		return dir
	}

	return filepath.Join(root, defaultDir)
}

func worktreeCut(path string) (string, bool) {
	sep := string(filepath.Separator)
	segments := strings.Split(path, sep)

	for i := 0; i+1 < len(segments); i++ {
		if segments[i] == claudeDir && segments[i+1] == worktreesDir {
			return filepath.Join(strings.Join(segments[:i], sep), defaultDir), true
		}
	}

	return "", false
}
