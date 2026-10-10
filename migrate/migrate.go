package migrate

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/pathologize"
	"gopkg.in/yaml.v3"

	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/git"
	"github.com/B4Dmonkey/bit-pro/project"
	"github.com/B4Dmonkey/bit-pro/store"
	"github.com/B4Dmonkey/bit-pro/task"
)

type Options struct {
	Dir string
	Git git.Runner
	Now func() time.Time
}

type Result struct {
	Code, Path string
	Already    bool
	Tracked    bool
	Renumbered []Renumber
}

type Renumber struct {
	From, To, Kept string
}

var ErrCodeTaken = errors.New("code belongs to another project")

type config struct {
	Prefix string `toml:"prefix"`
}

func Run(ctx context.Context, q *orm.Queries, opts Options) (Result, error) {
	src, path, err := source(opts.Dir)
	if err != nil {
		return Result{}, err
	}

	h := git.ReadHead(ctx, opts.Git, opts.Dir)
	head := task.Commit{SHA: h.SHA, Branch: h.Branch, At: opts.Now()}

	ps, res, err := already(ctx, q, path)
	if err != nil || res.Already {
		return res, err
	}

	code, err := readSource(src, ps)
	if err != nil {
		return Result{}, err
	}

	ren, renumbered, err := renumber(src, code)
	if err != nil {
		return Result{}, err
	}

	if err := land(src, code, h, head, ren); err != nil {
		return Result{}, err
	}

	if err := q.CreateProject(ctx, orm.CreateProjectParams{Path: path, Code: code}); err != nil {
		return Result{}, fmt.Errorf("registering %s: %w", path, err)
	}

	return Result{
		Code:       code,
		Path:       path,
		Tracked:    git.Tracks(ctx, opts.Git, filepath.Dir(src), ".bit"),
		Renumbered: renumbered,
	}, nil
}

func land(src, code string, h git.Head, head task.Commit, ren renames) error {
	data, err := store.Dir()
	if err != nil {
		return err
	}

	stage, err := os.MkdirTemp(data, ".migrate-")
	if err != nil {
		return fmt.Errorf("creating a staging dir in %s: %w", data, err)
	}
	defer os.RemoveAll(stage)

	staged := filepath.Join(stage, code)
	if err := os.MkdirAll(staged, dirMode); err != nil {
		return fmt.Errorf("creating %s: %w", staged, err)
	}

	s := task.NewProject(staged, code).WithDataRoot(stage)

	if err := copyAll(src, s, h, head, ren); err != nil {
		return err
	}

	if problems := verify(src, s, code, h, ren); len(problems) > 0 {
		return fmt.Errorf("%w:\n  %s", ErrVerify, strings.Join(problems, "\n  "))
	}

	root, err := store.ProjectDir(code)
	if err != nil {
		return err
	}

	return commit(stage, data, staged, root)
}

func already(ctx context.Context, q *orm.Queries, path string) ([]project.Project, Result, error) {
	ps, err := project.Load(ctx, q)
	if err != nil {
		return nil, Result{}, err
	}

	p, ok := project.ByPath(ps, path)
	if !ok {
		return ps, Result{}, nil
	}

	if p.Removed {
		return nil, Result{}, fmt.Errorf("%s: %w", path, project.ErrRemoved)
	}

	return ps, Result{Code: p.Code, Path: p.Path, Already: true}, nil
}

func readSource(src string, ps []project.Project) (string, error) {
	cfgPath := filepath.Join(src, "config.toml")

	var cfg config
	if _, err := toml.DecodeFile(cfgPath, &cfg); err != nil {
		return "", fmt.Errorf("reading %s config: %w", src, err)
	}

	code, err := project.ValidateCode(cfg.Prefix)
	if err != nil {
		return "", fmt.Errorf("%s: %w", cfgPath, err)
	}

	if err := claimed(ps, code); err != nil {
		return "", err
	}

	return code, checkKnown(src)
}

func claimed(ps []project.Project, code string) error {
	p, ok := project.ByCode(ps, code)
	if !ok {
		return nil
	}

	if p.Removed {
		return fmt.Errorf("code %s: %w", code, project.ErrCodeRemoved)
	}

	return fmt.Errorf("code %s is already used by %s: %w", code, p.Path, ErrCodeTaken)
}

const dirMode = 0o755

var places = []struct {
	dir   string
	place task.Place
}{
	{"tasks", task.Active},
	{"completed", task.Completed},
	{filepath.Join("archive", "tasks"), task.Archived},
}

type rename struct {
	id    string
	order []string
}

type renames map[task.Place]map[string]rename

func (r renames) set(p task.Place, stem string, to rename) {
	if r[p] == nil {
		r[p] = map[string]rename{}
	}

	r[p][stem] = to
}

func (r rename) apply(t *task.Task) {
	t.ID = r.id
	if r.order != nil {
		t.Order = r.order
	}
}

var keepRank = map[task.Place]int{task.Completed: 0, task.Archived: 1, task.Active: 2}

type sourceFile struct {
	place  task.Place
	stem   string
	title  string
	ids    rawIDs
	traces bool
}

type rawIDs struct {
	ID    string   `yaml:"id"`
	Order []string `yaml:"order"`
}

func readRawIDs(raw []byte) (rawIDs, error) {
	head, _ := splitFrontmatter(raw)
	content := head[len("---\n") : len(head)-len("---\n")]

	var ids rawIDs
	if err := yaml.Unmarshal(content, &ids); err != nil {
		return rawIDs{}, err
	}

	return ids, nil
}

func hasLowercase(ids ...string) bool {
	for _, id := range ids {
		if id != strings.ToUpper(id) {
			return true
		}
	}

	return false
}

func renumber(src, code string) (renames, []Renumber, error) {
	tracks, bars, highest, err := sourceFiles(src, code)
	if err != nil {
		return nil, nil, err
	}

	ren := renames{}

	var out []Renumber

	ids := slices.SortedFunc(maps.Keys(tracks), func(a, b string) int {
		return cmp.Or(trackNumber(code, a)-trackNumber(code, b), strings.Compare(a, b))
	})

	for _, id := range ids {
		group := tracks[id]
		if len(group) < 2 {
			continue
		}

		slices.SortStableFunc(group, keepFirst)
		kept := group[0]

		for _, moved := range group[1:] {
			highest++
			to := fmt.Sprintf("%s-%d", code, highest)
			ren.set(moved.place, moved.stem, rename{id: to, order: moveBars(ren, bars, moved, kept, to)})
			out = append(out, Renumber{From: id, To: to, Kept: kept.title})
		}
	}

	return ren, out, nil
}

func moveBars(ren renames, bars []sourceFile, moved, kept sourceFile, to string) []string {
	var order []string

	for _, entry := range moved.ids.Order {
		if slices.Contains(kept.ids.Order, entry) {
			order = append(order, task.NormalizeID(entry))

			continue
		}

		_, suffix, _ := strings.Cut(entry, ".")
		newID := to + "." + suffix
		order = append(order, newID)

		for _, b := range bars {
			if b.ids.ID == entry {
				ren.set(b.place, b.stem, rename{id: newID})
			}
		}
	}

	return order
}

func keepFirst(a, b sourceFile) int {
	if a.traces != b.traces {
		if a.traces {
			return -1
		}

		return 1
	}

	return keepRank[a.place] - keepRank[b.place]
}

func sourceFiles(src, code string) (tracks map[string][]sourceFile, bars []sourceFile, highest int, err error) {
	tracks = map[string][]sourceFile{}

	for _, pl := range places {
		dir := filepath.Join(src, pl.dir)

		stems, err := sourceStems(dir, ".md")
		if err != nil {
			return nil, nil, 0, fmt.Errorf("listing %s: %w", dir, err)
		}

		for _, stem := range stems {
			f, id, err := readSourceFile(filepath.Join(dir, stem+".md"), pl.place, stem)
			if err != nil {
				return nil, nil, 0, err
			}

			if strings.Contains(id, ".") {
				bars = append(bars, f)

				continue
			}

			highest = max(highest, trackNumber(code, id))
			tracks[id] = append(tracks[id], f)
		}
	}

	return tracks, bars, highest, nil
}

func readSourceFile(path string, p task.Place, stem string) (sourceFile, string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return sourceFile{}, "", fmt.Errorf("reading %s: %w", path, err)
	}

	t, err := task.Parse(raw)
	if err != nil {
		return sourceFile{}, "", fmt.Errorf("parsing %s: %w", path, err)
	}

	ids, err := readRawIDs(raw)
	if err != nil {
		return sourceFile{}, "", fmt.Errorf("parsing %s: %w", path, err)
	}

	f := sourceFile{
		place:  p,
		stem:   stem,
		title:  t.Title,
		ids:    ids,
		traces: hasLowercase(append([]string{stem, ids.ID}, ids.Order...)...),
	}

	return f, t.ID, nil
}

func trackNumber(code, id string) int {
	n, err := strconv.Atoi(strings.TrimPrefix(id, code+"-"))
	if err != nil {
		return 0
	}

	return n
}

func copyAll(src string, s *task.Store, h git.Head, head task.Commit, ren renames) error {
	for _, pl := range places {
		if err := copyTasks(filepath.Join(src, pl.dir), s, pl.place, h, ren[pl.place]); err != nil {
			return err
		}
	}

	if err := copyNotes(filepath.Join(src, "feedback"), s, head); err != nil {
		return err
	}

	if err := copyResearch(filepath.Join(src, "research"), s, head); err != nil {
		return err
	}

	return copyRetro(filepath.Join(src, "retro"), s, head)
}

func copyTasks(dir string, s *task.Store, p task.Place, h git.Head, ren map[string]rename) error {
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

		if r, ok := ren[strings.TrimSuffix(e.Name(), ".md")]; ok {
			r.apply(t)
		}

		t.Branch = h.Branch
		t.Commit = h.SHA

		if err := s.SaveTo(p, t); err != nil {
			return fmt.Errorf("copying %s: %w", path, err)
		}
	}

	return nil
}

var noteName = regexp.MustCompile(`^(.+)-(\d+)\.md$`)

func copyNotes(dir string, s *task.Store, head task.Commit) error {
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

		if _, err := s.ImportNote(m[1], seq, string(raw), head); err != nil {
			return fmt.Errorf("copying %s: %w", path, err)
		}
	}

	return nil
}

func copyResearch(dir string, s *task.Store, head task.Commit) error {
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
			if _, err := s.WriteResearch(tr.Name(), topic, string(raw), head); err != nil {
				return fmt.Errorf("copying %s: %w", path, err)
			}
		}
	}

	return nil
}

func copyRetro(dir string, s *task.Store, head task.Commit) error {
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

		if _, err := s.WriteRetro(strings.TrimSuffix(e.Name(), ".md"), string(raw), head); err != nil {
			return fmt.Errorf("copying %s: %w", path, err)
		}
	}

	return nil
}

var ErrVerify = errors.New("migrated copy doesn't match the source")

func verify(src string, s *task.Store, code string, h git.Head, ren renames) []string {
	var problems []string

	for _, pl := range places {
		problems = append(problems, verifyTasks(src, pl.dir, s, pl.place, code, h, ren[pl.place])...)
	}

	problems = append(problems, verifyNotes(src, s)...)
	problems = append(problems, verifyResearch(src, s)...)

	return append(problems, verifyRetro(src, s, code)...)
}

func sourceStems(dir, suffix string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var stems []string

	for _, e := range entries {
		if e.Type().IsRegular() && strings.HasSuffix(e.Name(), suffix) {
			stems = append(stems, strings.TrimSuffix(e.Name(), ".md"))
		}
	}

	return stems, nil
}

func normalizeStems(stems []string, ren map[string]rename) (ids []string, rawOf map[string]string) {
	rawOf = make(map[string]string, len(stems))
	ids = make([]string, 0, len(stems))

	for _, stem := range stems {
		id := task.NormalizeID(stem)
		if r, ok := ren[stem]; ok {
			id = r.id
		}

		rawOf[id] = stem
		ids = append(ids, id)
	}

	return ids, rawOf
}

func compareSets(rel string, want, got []string) (problems, both []string) {
	for _, w := range want {
		if slices.Contains(got, w) {
			both = append(both, w)
		} else {
			problems = append(problems, filepath.Join(rel, w+".md")+": missing from the migrated copy")
		}
	}

	for _, g := range got {
		if !slices.Contains(want, g) {
			problems = append(problems, filepath.Join(rel, g+".md")+": in the migrated copy but not the source")
		}
	}

	return problems, both
}

func verifyTasks(
	src, rel string, s *task.Store, p task.Place, code string, h git.Head, ren map[string]rename,
) []string {
	stems, err := sourceStems(filepath.Join(src, rel), ".md")
	if err != nil {
		return []string{rel + ": " + err.Error()}
	}

	want, rawOf := normalizeStems(stems, ren)

	got, err := s.IDs(p)
	if err != nil {
		return []string{rel + ": " + err.Error()}
	}

	problems, both := compareSets(rel, want, got)

	for _, id := range both {
		name := filepath.Join(rel, rawOf[id]+".md")

		raw, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			problems = append(problems, name+": "+err.Error())

			continue
		}

		srcTask, err := task.Parse(raw)
		if err != nil {
			problems = append(problems, name+": "+err.Error())

			continue
		}

		srcTask.ID = id
		if r, ok := ren[rawOf[id]]; ok {
			r.apply(srcTask)
		}

		copied, err := s.LoadFrom(p, id)
		if err != nil {
			problems = append(problems, name+": "+err.Error())

			continue
		}

		if diff := taskDiff(srcTask, copied, code, h); len(diff) > 0 {
			problems = append(problems, name+": "+strings.Join(diff, ", ")+" differ")
		}
	}

	return problems
}

func taskDiff(want, got *task.Task, code string, h git.Head) []string {
	checks := []struct {
		field string
		same  bool
	}{
		{"id", want.ID == got.ID},
		{"title", want.Title == got.Title},
		{"status", want.Status == got.Status},
		{"approved", want.Approved == got.Approved},
		{"phase", want.Phase == got.Phase},
		{"phase_label", want.PhaseLabel == got.PhaseLabel},
		{"order", slices.Equal(want.Order, got.Order)},
		{"body", want.Body == got.Body},
		{"branch", got.Branch == h.Branch},
		{"commit", got.Commit == h.SHA},
		{"project", got.Project == code},
	}

	var diff []string

	for _, c := range checks {
		if !c.same {
			diff = append(diff, c.field)
		}
	}

	return diff
}

func verifyBodies(src, rel string, stems []string, read func(stem string) (string, error)) []string {
	var problems []string

	for _, stem := range stems {
		name := filepath.Join(rel, stem+".md")

		raw, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			problems = append(problems, name+": "+err.Error())

			continue
		}

		body, err := read(stem)
		if err != nil {
			problems = append(problems, name+": "+err.Error())

			continue
		}

		if body != string(raw) {
			problems = append(problems, name+": body differs")
		}
	}

	return problems
}

func verifyNotes(src string, s *task.Store) []string {
	const rel = "feedback"

	stems, err := sourceStems(filepath.Join(src, rel), ".md")
	if err != nil {
		return []string{rel + ": " + err.Error()}
	}

	got, err := s.ListNotes("")
	if err != nil {
		return []string{rel + ": " + err.Error()}
	}

	want, rawOf := normalizeStems(stems, nil)

	problems, both := compareSets(rel, want, got)

	raw := make([]string, 0, len(both))
	for _, id := range both {
		raw = append(raw, rawOf[id])
	}

	return append(problems, verifyBodies(src, rel, raw, s.ReadNote)...)
}

func verifyResearch(src string, s *task.Store) []string {
	const rel = "research"

	tracks, err := os.ReadDir(filepath.Join(src, rel))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	if err != nil {
		return []string{rel + ": " + err.Error()}
	}

	var problems []string

	for _, tr := range tracks {
		if !tr.IsDir() {
			continue
		}

		trackRel := filepath.Join(rel, tr.Name())

		want, err := sourceStems(filepath.Join(src, trackRel), ".md")
		if err != nil {
			problems = append(problems, trackRel+": "+err.Error())

			continue
		}

		got, err := s.ResearchTopics(tr.Name())
		if err != nil {
			problems = append(problems, trackRel+": "+err.Error())

			continue
		}

		missing, both := compareSets(trackRel, want, got)
		problems = append(problems, missing...)
		problems = append(problems, verifyBodies(src, trackRel, both, func(topic string) (string, error) {
			return s.ReadResearch(tr.Name(), topic)
		})...)
	}

	return problems
}

func verifyRetro(src string, s *task.Store, code string) []string {
	const rel = "retro"

	stems, err := sourceStems(filepath.Join(src, rel), "-proposals.md")
	if err != nil {
		return []string{rel + ": " + err.Error()}
	}

	listed, err := s.ListRetro()
	if err != nil {
		return []string{rel + ": " + err.Error()}
	}

	got := make([]string, 0, len(listed))
	for _, p := range listed {
		got = append(got, p.Name)
	}

	stemOf := make(map[string]string, len(stems))
	want := make([]string, 0, len(stems))

	for _, stem := range stems {
		name := task.RetroName(code, stem)
		stemOf[name] = stem
		want = append(want, name)
	}

	var problems []string

	for _, name := range want {
		if !slices.Contains(got, name) {
			problems = append(problems, filepath.Join(rel, stemOf[name]+".md")+": missing from the migrated copy as "+name)
		}
	}

	for _, name := range got {
		if !slices.Contains(want, name) {
			problems = append(problems, filepath.Join(rel, name+".md")+": in the migrated copy but not the source")
		}
	}

	var both []string

	for _, stem := range stems {
		if slices.Contains(got, task.RetroName(code, stem)) {
			both = append(both, stem)
		}
	}

	return append(problems, verifyBodies(src, rel, both, func(stem string) (string, error) {
		_, body, err := s.ReadRetro(task.RetroName(code, stem))

		return body, err
	})...)
}

type move struct {
	from, to string
}

func commit(stage, data, staged, root string) error {
	var moves []move

	for _, sub := range []string{"feedback", "retro"} {
		entries, err := os.ReadDir(filepath.Join(stage, sub))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return fmt.Errorf("reading staged %s: %w", sub, err)
		}

		for _, e := range entries {
			moves = append(moves, move{
				from: filepath.Join(stage, sub, e.Name()),
				to:   pathologize.Join(filepath.Join(data, sub), e.Name()),
			})
		}
	}

	for _, dst := range append(destinations(moves), root) {
		if _, err := os.Lstat(dst); err == nil {
			return fmt.Errorf("committing the migration: %s already exists", dst)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("checking %s: %w", dst, err)
		}
	}

	for _, m := range moves {
		if err := os.MkdirAll(filepath.Dir(m.to), dirMode); err != nil {
			return fmt.Errorf("creating %s: %w", filepath.Dir(m.to), err)
		}

		if err := os.Rename(m.from, m.to); err != nil {
			return fmt.Errorf("moving %s into place: %w", m.to, err)
		}
	}

	if err := os.Rename(staged, root); err != nil {
		return fmt.Errorf("moving %s into place: %w", root, err)
	}

	return nil
}

func destinations(moves []move) []string {
	dsts := make([]string, 0, len(moves)+1)
	for _, m := range moves {
		dsts = append(dsts, m.to)
	}

	return dsts
}
