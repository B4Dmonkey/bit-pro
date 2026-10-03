package migrate

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var ErrUnknownFiles = errors.New(".bit/ holds files migrate doesn't know")

var feedbackName = regexp.MustCompile(`^.+-\d+-\d+\.md$`)

func checkKnown(src string) error {
	unknown, err := unknownFiles(src)
	if err != nil {
		return err
	}

	if len(unknown) > 0 {
		return fmt.Errorf("%w:\n  %s", ErrUnknownFiles, strings.Join(unknown, "\n  "))
	}

	return nil
}

func unknownFiles(src string) ([]string, error) {
	var unknown []string

	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.Type().IsRegular() {
			return nil
		}

		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}

		rel = filepath.ToSlash(rel)
		if !known(rel) {
			unknown = append(unknown, rel)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", src, err)
	}

	slices.Sort(unknown)

	return unknown, nil
}

func known(rel string) bool {
	parts := strings.Split(rel, "/")
	name := parts[len(parts)-1]
	md := strings.HasSuffix(name, ".md") && !strings.HasPrefix(name, ".")

	switch path.Dir(rel) {
	case ".":
		return rel == "config.toml"
	case "tasks", "completed", "archive/tasks":
		return md
	case "feedback":
		return md && feedbackName.MatchString(name)
	case "retro":
		return md && strings.HasSuffix(name, "-proposals.md")
	}

	return len(parts) == 3 && parts[0] == "research" && md
}
