package migrate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/store"
	"github.com/B4Dmonkey/bit-pro/task"
)

type Options struct {
	Dir string
}

type Result struct {
	Code, Path string
}

type config struct {
	Prefix string `toml:"prefix"`
}

func Run(ctx context.Context, q *orm.Queries, opts Options) (Result, error) {
	src := filepath.Join(opts.Dir, ".bit")

	var cfg config
	if _, err := toml.DecodeFile(filepath.Join(src, "config.toml"), &cfg); err != nil {
		return Result{}, fmt.Errorf("reading %s config: %w", src, err)
	}

	code := cfg.Prefix

	path, err := project.CanonicalPath(filepath.Dir(src))
	if err != nil {
		return Result{}, err
	}

	data, err := store.Dir()
	if err != nil {
		return Result{}, err
	}

	root, err := store.ProjectDir(code)
	if err != nil {
		return Result{}, err
	}

	s := task.NewProject(root, code).WithDataRoot(data)

	for _, pl := range []struct {
		dir   string
		place task.Place
	}{
		{"tasks", task.Active},
		{"completed", task.Completed},
		{filepath.Join("archive", "tasks"), task.Archived},
	} {
		if err := copyTasks(filepath.Join(src, pl.dir), s, pl.place); err != nil {
			return Result{}, err
		}
	}

	if err := copyNotes(filepath.Join(src, "feedback"), s); err != nil {
		return Result{}, err
	}

	if err := copyResearch(filepath.Join(src, "research"), s); err != nil {
		return Result{}, err
	}

	if err := copyRetro(filepath.Join(src, "retro"), s); err != nil {
		return Result{}, err
	}

	if err := q.CreateProject(ctx, orm.CreateProjectParams{Path: path, Code: code}); err != nil {
		return Result{}, fmt.Errorf("registering %s: %w", path, err)
	}

	return Result{Code: code, Path: path}, nil
}

func copyTasks(dir string, s *task.Store, p task.Place) error {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading %s: %w", dir, err)
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, e.Name())

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		t, err := task.Parse(raw)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		if err := s.SaveTo(p, t); err != nil {
			return fmt.Errorf("copying %s: %w", path, err)
		}
	}

	return nil
}

var noteName = regexp.MustCompile(`^(.+)-(\d+)\.md$`)

func copyNotes(dir string, s *task.Store) error {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading %s: %w", dir, err)
	}

	for _, e := range entries {
		m := noteName.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}

		seq, err := strconv.Atoi(m[2])
		if err != nil {
			return fmt.Errorf("parsing %s: %w", e.Name(), err)
		}

		path := filepath.Join(dir, e.Name())

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		if _, err := s.ImportNote(m[1], seq, string(raw), task.Commit{}); err != nil {
			return fmt.Errorf("copying %s: %w", path, err)
		}
	}

	return nil
}

func copyResearch(dir string, s *task.Store) error {
	tracks, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading %s: %w", dir, err)
	}

	for _, tr := range tracks {
		if !tr.IsDir() {
			continue
		}

		trackDir := filepath.Join(dir, tr.Name())

		topics, err := os.ReadDir(trackDir)
		if err != nil {
			return fmt.Errorf("reading %s: %w", trackDir, err)
		}

		for _, e := range topics {
			if !strings.HasSuffix(e.Name(), ".md") {
				continue
			}

			path := filepath.Join(trackDir, e.Name())

			raw, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}

			topic := strings.TrimSuffix(e.Name(), ".md")
			if _, err := s.WriteResearch(tr.Name(), topic, string(raw), task.Commit{}); err != nil {
				return fmt.Errorf("copying %s: %w", path, err)
			}
		}
	}

	return nil
}

func copyRetro(dir string, s *task.Store) error {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading %s: %w", dir, err)
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), "-proposals.md") {
			continue
		}

		path := filepath.Join(dir, e.Name())

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		if _, err := s.WriteRetro(strings.TrimSuffix(e.Name(), ".md"), string(raw), task.Commit{}); err != nil {
			return fmt.Errorf("copying %s: %w", path, err)
		}
	}

	return nil
}
