package project

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
)

var (
	ErrNotRegistered = errors.New("not a bit project; run `bp add`")
	ErrNeedsMigrate  = errors.New("found a v1 .bit/ directory; run `bp migrate`")
	ErrRemoved       = errors.New("this project was removed; run `bp add` here to revive it")
)

type Project struct {
	ID      int64
	Code    string
	Path    string
	Removed bool
}

func CanonicalPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("absolute path of %s: %w", path, err)
	}

	existing, tail := abs, ""

	for {
		resolved, err := filepath.EvalSymlinks(existing)
		if err == nil {
			return filepath.Join(resolved, tail), nil
		}

		parent := filepath.Dir(existing)
		if !errors.Is(err, fs.ErrNotExist) || parent == existing {
			return "", fmt.Errorf("resolving %s: %w", path, err)
		}

		tail = filepath.Join(filepath.Base(existing), tail)
		existing = parent
	}
}

func Resolve(projects []Project, dir string) (Project, error) {
	canonical, err := CanonicalPath(dir)
	if err != nil {
		return Project{}, err
	}

	var (
		best     Project
		bestPath string
	)

	for _, p := range projects {
		path, err := CanonicalPath(p.Path)
		if err != nil {
			return Project{}, err
		}

		if within(path, canonical) && len(path) > len(bestPath) {
			best, bestPath = p, path
		}
	}

	if best.Removed {
		return Project{}, fmt.Errorf("%s: %w", dir, ErrRemoved)
	}

	if bestPath != "" {
		return best, nil
	}

	if bitDirAbove(canonical) {
		return Project{}, fmt.Errorf("%s: %w", dir, ErrNeedsMigrate)
	}

	return Project{}, fmt.Errorf("%s: %w", dir, ErrNotRegistered)
}

func Load(ctx context.Context, q *orm.Queries) ([]Project, error) {
	rows, err := q.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	projects := make([]Project, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, Project{ID: row.ID, Code: row.Code, Path: row.Path, Removed: row.Removed != 0})
	}

	return projects, nil
}

func Find(ctx context.Context, dir string) (Project, error) {
	sqlDB, err := db.Open()
	if err != nil {
		return Project{}, err
	}
	defer sqlDB.Close()

	projects, err := Load(ctx, orm.New(sqlDB))
	if err != nil {
		return Project{}, err
	}

	return Resolve(projects, dir)
}

func within(path, dir string) bool {
	if strings.EqualFold(dir, path) {
		return true
	}

	prefix := path + string(filepath.Separator)

	return len(dir) > len(prefix) && strings.EqualFold(dir[:len(prefix)], prefix)
}

func bitDirAbove(dir string) bool {
	for {
		if info, err := os.Stat(filepath.Join(dir, ".bit")); err == nil && info.IsDir() {
			return true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}

		dir = parent
	}
}
