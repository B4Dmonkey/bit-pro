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
	tasksSubdir     = "tasks"
	completedSubdir = "completed"
	archiveSubdir   = "archive"
	configFileName  = "config.toml"
	dirMode         = 0o755
	fileMode        = 0o644
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
	return pathologize.Join(s.tasksDir(), normalizeID(id)+".md")
}

func (s *Store) archiveTasksDir() string {
	return filepath.Join(s.root, archiveSubdir, tasksSubdir)
}

func (s *Store) archivePath(id string) string {
	return pathologize.Join(s.archiveTasksDir(), normalizeID(id)+".md")
}

func (s *Store) completedDir() string {
	return filepath.Join(s.root, completedSubdir)
}

func (s *Store) completedPath(id string) string {
	return pathologize.Join(s.completedDir(), normalizeID(id)+".md")
}

func (s *Store) relocateInto(dir, id string) error {
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}
	if err := os.Rename(s.Path(id), pathologize.Join(dir, normalizeID(id)+".md")); err != nil {
		return fmt.Errorf("relocating task %s: %w", id, err)
	}
	return nil
}

func normalizeID(id string) string {
	return strings.ToUpper(id)
}

func (s *Store) Children(parent string) ([]*Task, error) {
	return s.children(parent)
}

func (s *Store) children(parent string) ([]*Task, error) {
	parent = normalizeID(parent)
	tasks, err := s.List()
	if err != nil {
		return nil, err
	}
	var kids []*Task
	for _, t := range tasks {
		if p, ok := barParent(t.ID); ok && p == parent {
			kids = append(kids, t)
		}
	}
	return kids, nil
}

type UnfinishedBarsError struct {
	Bars []string
}

func (e *UnfinishedBarsError) Error() string {
	return "cannot relocate: unfinished bars " + strings.Join(e.Bars, ", ")
}

func (s *Store) Relocate(id string, force bool) error {
	return s.relocateTree(s.archiveTasksDir(), id, force)
}

func (s *Store) Complete(id string) error {
	return s.relocateTree(s.completedDir(), id, false)
}

func (s *Store) relocateTree(dir, id string, force bool) error {
	kids, err := s.children(id)
	if err != nil {
		return err
	}
	if !force {
		var unfinished []string
		for _, kid := range kids {
			if kid.Status != "done" {
				unfinished = append(unfinished, kid.ID)
			}
		}
		if len(unfinished) > 0 {
			return &UnfinishedBarsError{Bars: unfinished}
		}
	}
	for _, kid := range kids {
		if err := s.relocateInto(dir, kid.ID); err != nil {
			return err
		}
	}
	if err := s.relocateInto(dir, id); err != nil {
		return err
	}
	if parent, ok := barParent(id); ok {
		return s.removeFromOrder(parent, id)
	}
	return nil
}

func (s *Store) removeFromOrder(parent, id string) error {
	track, err := s.Load(parent)
	if err != nil {
		return err
	}
	if len(track.Order) == 0 {
		return nil
	}
	track.Order = slices.DeleteFunc(track.Order, func(x string) bool { return x == id })
	return s.Save(track)
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

func (s *Store) Move(id, anchor string, before bool) error {
	id, anchor = normalizeID(id), normalizeID(anchor)
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
	return s.insertInOrder(normalizeID(parent), normalizeID(id), normalizeID(anchor), false)
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
	parent, id = normalizeID(parent), normalizeID(id)
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
	parent = normalizeID(parent)
	if _, err := os.Stat(s.Path(parent)); err != nil {
		return "", fmt.Errorf("parent %s does not exist: %w", parent, err)
	}

	glob := parent + ".*.md"
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(parent) + `\.(\d+)\.md$`)
	highest, err := s.highestReserved(glob, re, "child IDs")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%d", parent, highest+1), nil
}

func (s *Store) NextID(prefix string) (string, error) {
	glob := prefix + "-*.md"
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `-(\d+)\.md$`)
	highest, err := s.highestReserved(glob, re, "task IDs")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d", prefix, highest+1), nil
}

func (s *Store) highestReserved(glob string, re *regexp.Regexp, what string) (int, error) {
	highest := 0
	for _, dir := range []string{s.tasksDir(), s.completedDir(), s.archiveTasksDir()} {
		n, err := highestSuffix(dir, glob, re)
		if err != nil {
			return 0, fmt.Errorf("scanning %s for existing %s: %w", dir, what, err)
		}
		highest = max(highest, n)
	}
	return highest, nil
}

func highestSuffix(dir, glob string, re *regexp.Regexp) (int, error) {
	matches, err := filepath.Glob(filepath.Join(dir, glob))
	if err != nil {
		return 0, err
	}
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
	return highest, nil
}
