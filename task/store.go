package task

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/pathologize"
)

const (
	tasksSubdir    = "tasks"
	configFileName = "config.toml"
	dirMode        = 0o755
	fileMode       = 0o644
)

type Store struct {
	root string
}

func New(root string) *Store {
	return &Store{root: root}
}

func (s *Store) tasksDir() string {
	return filepath.Join(s.root, tasksSubdir)
}

func (s *Store) Path(id string) string {
	return pathologize.Join(s.tasksDir(), id+".md")
}

func (s *Store) Load(id string) (*Task, error) {
	data, err := os.ReadFile(s.Path(id))
	if err != nil {
		return nil, fmt.Errorf("loading task %s: %w", id, err)
	}
	return Parse(data)
}

func (s *Store) Save(t *Task) error {
	data, err := t.Bytes()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.tasksDir(), dirMode); err != nil {
		return fmt.Errorf("creating %s: %w", s.tasksDir(), err)
	}
	if err := os.WriteFile(s.Path(t.ID), data, fileMode); err != nil {
		return fmt.Errorf("writing task %s: %w", t.ID, err)
	}
	return nil
}

func (s *Store) Delete(id string) error {
	if err := os.Remove(s.Path(id)); err != nil {
		return fmt.Errorf("deleting task %s: %w", id, err)
	}
	return nil
}

func (s *Store) List() ([]*Task, error) {
	matches, err := filepath.Glob(filepath.Join(s.tasksDir(), "*.md"))
	if err != nil {
		return nil, fmt.Errorf("scanning %s for tasks: %w", s.tasksDir(), err)
	}

	tasks := make([]*Task, 0, len(matches))
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		t, err := Parse(data)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}
		tasks = append(tasks, t)
	}
	slices.SortFunc(tasks, func(a, b *Task) int { return compareIDs(a.ID, b.ID) })
	return tasks, nil
}

// compareIDs orders IDs by (track, bar): tracks descending, newest first
// (BIT-10 before BIT-9), and a track's own bars ascending directly beneath
// it (BIT-2.1 before BIT-2.13). A track's own bar number is 0, so it always
// heads its group. An unparseable suffix can only come from a hand-edited
// file and sorts last.
func compareIDs(a, b string) int {
	at, ab, aOK := idParts(a)
	bt, bb, bOK := idParts(b)
	switch {
	case aOK && bOK:
		if at != bt {
			return bt - at
		}
		return ab - bb
	case aOK:
		return -1
	case bOK:
		return 1
	default:
		return 0
	}
}

func idParts(id string) (track, bar int, ok bool) {
	suffix := id[strings.LastIndex(id, "-")+1:]
	trackStr, barStr, hasBar := strings.Cut(suffix, ".")
	track, err := strconv.Atoi(trackStr)
	if err != nil {
		return 0, 0, false
	}
	if !hasBar {
		return track, 0, true
	}
	bar, err = strconv.Atoi(barStr)
	if err != nil {
		return 0, 0, false
	}
	return track, bar, true
}

func (s *Store) NextChildID(parent string) (string, error) {
	if _, err := os.Stat(s.Path(parent)); err != nil {
		return "", fmt.Errorf("parent %s does not exist: %w", parent, err)
	}

	matches, err := filepath.Glob(filepath.Join(s.tasksDir(), parent+".*.md"))
	if err != nil {
		return "", fmt.Errorf("scanning %s for existing child IDs: %w", s.tasksDir(), err)
	}

	re := regexp.MustCompile(`^` + regexp.QuoteMeta(parent) + `\.(\d+)\.md$`)
	highest := 0
	for _, m := range matches {
		sub := re.FindStringSubmatch(filepath.Base(m))
		if sub == nil {
			continue
		}
		n, err := strconv.Atoi(sub[1])
		if err != nil {
			continue
		}
		highest = max(highest, n)
	}
	return fmt.Sprintf("%s.%d", parent, highest+1), nil
}

func (s *Store) NextID(prefix string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(s.tasksDir(), prefix+"-*.md"))
	if err != nil {
		return "", fmt.Errorf("scanning %s for existing task IDs: %w", s.tasksDir(), err)
	}

	re := regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `-(\d+)\.md$`)
	highest := 0
	for _, m := range matches {
		sub := re.FindStringSubmatch(filepath.Base(m))
		if sub == nil {
			continue
		}
		n, err := strconv.Atoi(sub[1])
		if err != nil {
			continue
		}
		highest = max(highest, n)
	}
	return fmt.Sprintf("%s-%d", prefix, highest+1), nil
}
