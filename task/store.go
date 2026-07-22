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
	archiveSubdir  = "archive"
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

func (s *Store) archiveDir() string {
	return filepath.Join(s.root, archiveSubdir)
}

func (s *Store) archivePath(id string) string {
	return pathologize.Join(s.archiveDir(), id+".md")
}

func (s *Store) relocate(id string) error {
	if err := os.MkdirAll(s.archiveDir(), dirMode); err != nil {
		return fmt.Errorf("creating %s: %w", s.archiveDir(), err)
	}
	if err := os.Rename(s.Path(id), s.archivePath(id)); err != nil {
		return fmt.Errorf("relocating task %s: %w", id, err)
	}
	return nil
}

func (s *Store) Relocate(id string, force bool) error {
	return s.relocate(id)
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

func (s *Store) Move(id, anchor string, before bool) error {
	if id == anchor {
		return fmt.Errorf("cannot move %s relative to itself", id)
	}
	parent, ok := barParent(id)
	if !ok {
		return fmt.Errorf("%s is not a bar", id)
	}
	aParent, ok := barParent(anchor)
	if !ok || aParent != parent {
		return fmt.Errorf("anchor %s is not a sibling of %s", anchor, id)
	}
	if _, err := s.Load(id); err != nil {
		return err
	}
	if _, err := s.Load(anchor); err != nil {
		return err
	}
	return s.insertInOrder(parent, id, anchor, before)
}

// InsertAfter places a newly created bar id into parent's Order immediately
// after anchor. A parent that has never been reordered has its order
// materialized from the current ID sequence first, so the position is well
// defined even for a legacy track. The anchor must already belong to that
// order — an unknown or cross-track anchor is refused rather than silently
// appended, which keeps the list ⇄ files bijection intact.
func (s *Store) InsertAfter(parent, id, anchor string) error {
	return s.insertInOrder(parent, id, anchor, false)
}

// insertInOrder is the shared splice: it materializes a legacy parent's order,
// drops id if it is already present, then places it before or after anchor. Both
// Move (relocating an existing bar) and InsertAfter (positioning a new one) route
// through here so ordering has a single definition.
func (s *Store) insertInOrder(parent, id, anchor string, before bool) error {
	track, err := s.Load(parent)
	if err != nil {
		return err
	}
	order := track.Order
	if len(order) == 0 {
		order, err = s.materializeOrder(parent)
		if err != nil {
			return err
		}
	}
	order = slices.DeleteFunc(order, func(x string) bool { return x == id })
	at := slices.Index(order, anchor)
	if at == -1 {
		return fmt.Errorf("anchor %s is not in %s's order", anchor, parent)
	}
	if !before {
		at++
	}
	track.Order = slices.Insert(order, at, id)
	return s.Save(track)
}

// AppendToOrder adds a newly created child to its parent track's explicit
// Order. A parent that has never been reordered has an empty Order, and List
// still places the new child last by ID, so it is left untouched — writing an
// Order there would be premature. Only a present, reordered manifest is
// extended, keeping the list ⇄ files bijection intact.
func (s *Store) AppendToOrder(parent, id string) error {
	track, err := s.Load(parent)
	if err != nil {
		return err
	}
	if len(track.Order) == 0 {
		return nil
	}
	track.Order = append(track.Order, id)
	return s.Save(track)
}

func (s *Store) materializeOrder(parent string) ([]string, error) {
	tasks, err := s.List()
	if err != nil {
		return nil, err
	}
	var order []string
	for _, t := range tasks {
		if p, ok := barParent(t.ID); ok && p == parent {
			order = append(order, t.ID)
		}
	}
	return order, nil
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
	pos := orderPositions(tasks)
	slices.SortFunc(tasks, func(a, b *Task) int {
		if cmp, ok := compareByOrder(pos, a.ID, b.ID); ok {
			return cmp
		}
		return compareIDs(a.ID, b.ID)
	})
	return tasks, nil
}

// orderPositions indexes each track's explicit Order into a track ID → bar ID →
// position map, so List's comparator can look up a bar's rank in O(1).
func orderPositions(tasks []*Task) map[string]map[string]int {
	pos := make(map[string]map[string]int)
	for _, t := range tasks {
		if len(t.Order) == 0 {
			continue
		}
		idx := make(map[string]int, len(t.Order))
		for i, id := range t.Order {
			idx[id] = i
		}
		pos[t.ID] = idx
	}
	return pos
}

// compareByOrder ranks two bars by their track's explicit Order, but only when
// both are children of the same track and both appear in its order list. Every
// other pairing (tracks, cross-track, unlisted bars) falls through to compareIDs.
func compareByOrder(pos map[string]map[string]int, a, b string) (int, bool) {
	pa, aOK := barParent(a)
	pb, bOK := barParent(b)
	if !aOK || !bOK || pa != pb {
		return 0, false
	}
	idx, ok := pos[pa]
	if !ok {
		return 0, false
	}
	ia, aOK := idx[a]
	ib, bOK := idx[b]
	if !aOK || !bOK {
		return 0, false
	}
	return ia - ib, true
}

func barParent(id string) (string, bool) {
	i := strings.LastIndex(id, ".")
	if i == -1 {
		return "", false
	}
	return id[:i], true
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
