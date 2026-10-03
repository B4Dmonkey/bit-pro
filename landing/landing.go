package landing

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/B4Dmonkey/bit-pro/git"
)

type Bar struct {
	ID     string
	Status string
	Commit string
}

type Query struct {
	Dir    string
	Bars   []Bar
	Commit string
	PR     int
}

type AmbiguousPRError struct {
	PR   int
	SHAs []string
}

func (e *AmbiguousPRError) Error() string {
	return fmt.Sprintf("PR #%d matches more than one commit on trunk:\n%s", e.PR, strings.Join(e.SHAs, "\n"))
}

type Verdict string

const (
	Done     Verdict = "done"
	NotDone  Verdict = "not_done"
	Partly   Verdict = "partly"
	CantTell Verdict = "cant_tell"
	NoGit    Verdict = "no_git"
)

type Class string

const (
	Landed       Class = "landed"
	Pushed       Class = "pushed"
	Local        Class = "local"
	Unresolvable Class = "unresolvable"
	NoHash       Class = "no_hash"
	Squash       Class = "squash"
)

type BarResult struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Commit  string `json:"commit"`
	Class   Class  `json:"class"`
	Landing string `json:"landing"`
	Repoint bool   `json:"repoint"`
}

type Report struct {
	Trunk      string      `json:"trunk"`
	Branch     string      `json:"branch"`
	Shallow    bool        `json:"shallow"`
	Verdict    Verdict     `json:"verdict"`
	Landing    string      `json:"landing"`
	Bars       []BarResult `json:"bars"`
	Unfinished []string    `json:"unfinished"`
}

const (
	statusDone = "done"
	mainBranch = "main"
)

var (
	fullSHA   = regexp.MustCompile(`^[0-9a-f]{40}$`)
	answerSHA = regexp.MustCompile(`^[0-9a-fA-F]{4,40}$`)

	ErrNoTrunk    = errors.New("no trunk: neither origin/main nor main exists")
	ErrBadAnswer  = errors.New("not a commit")
	ErrNotOnTrunk = errors.New("not on trunk")
)

func Check(ctx context.Context, run git.Runner, q Query) (Report, error) {
	if !git.IsRepo(ctx, run, q.Dir) {
		return noGit(q), nil
	}

	return check(ctx, run, q)
}

func check(ctx context.Context, run git.Runner, q Query) (Report, error) {
	name, trunk, err := resolveTrunk(ctx, run, q.Dir)
	if err != nil {
		return Report{}, err
	}

	if git.IsShallow(ctx, run, q.Dir) {
		return shallow(name, q), nil
	}

	chain, pos, answer, err := readTrunk(ctx, run, q, name, trunk)
	if err != nil {
		return Report{}, err
	}

	newest := len(chain)

	r := Report{
		Trunk:      name,
		Branch:     mainBranch,
		Bars:       make([]BarResult, 0, len(q.Bars)),
		Unfinished: []string{},
	}

	counts := map[Class]int{}
	sq := &squashes{run: run, dir: q.Dir, trunk: trunk, pos: pos}

	for _, b := range q.Bars {
		res, at, err := place(ctx, run, q.Dir, b, trunk, pos, answer, sq)
		if err != nil {
			return Report{}, err
		}

		if at < 0 {
			counts[res.Class]++
		} else {
			counts[Landed]++
			res.Landing = chain[at]

			if at < newest {
				newest = at
				r.Landing = chain[at]
			}
		}

		if b.Status != statusDone {
			r.Unfinished = append(r.Unfinished, b.ID)
		}

		r.Bars = append(r.Bars, res)
	}

	r.Verdict = verdict(counts, len(r.Unfinished))
	if r.Verdict != Done && r.Verdict != Partly {
		r.Landing = ""
	}

	return r, nil
}

func readTrunk(
	ctx context.Context, run git.Runner, q Query, name, trunk string,
) ([]string, map[string]int, int, error) {
	chain, err := git.FirstParents(ctx, run, q.Dir, trunk)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("reading trunk %s: %w", name, err)
	}

	pos := positions(chain)

	commit := q.Commit
	if commit == "" && q.PR > 0 {
		commit, err = prCommit(ctx, run, q.Dir, q.PR, name, trunk)
		if err != nil {
			return nil, nil, 0, err
		}
	}

	answer, err := placeAnswer(ctx, run, q.Dir, commit, name, trunk, pos)

	return chain, pos, answer, err
}

func place(
	ctx context.Context, run git.Runner, dir string, b Bar, trunk string, pos map[string]int, answer int, sq *squashes,
) (BarResult, int, error) {
	res := BarResult{ID: b.ID, Status: b.Status, Commit: b.Commit}

	class, at, err := classify(ctx, run, dir, b.Commit, trunk, pos)
	if err != nil {
		return BarResult{}, 0, err
	}

	res.Class = class

	if class == Pushed || class == Local {
		if at, ok, err := sq.match(ctx, b.Commit); err != nil || ok {
			if ok {
				res.Class, res.Repoint = Squash, true
			}

			return res, at, err
		}
	}

	switch {
	case class == Landed:
		return res, at, nil
	case answer >= 0:
		res.Repoint = true

		return res, answer, nil
	default:
		return res, -1, nil
	}
}

type squashes struct {
	run    git.Runner
	dir    string
	trunk  string
	pos    map[string]int
	list   []git.Squash
	lines  []map[string]int
	loaded bool
}

func (s *squashes) match(ctx context.Context, sha string) (int, bool, error) {
	if !s.loaded {
		list, err := git.Squashes(ctx, s.run, s.dir, s.trunk)
		if err != nil {
			return 0, false, err
		}

		s.list, s.lines, s.loaded = list, lineCounts(list), true
	}

	subject, at, ok := git.Subject(ctx, s.run, s.dir, sha)
	if !ok {
		return -1, false, nil
	}

	for i, sq := range slices.Backward(s.list) {
		if sq.Time < at || s.lines[i][subject] == 0 {
			continue
		}

		if p, on := s.pos[sq.SHA]; on {
			s.lines[i][subject]--

			return p, true, nil
		}
	}

	return -1, false, nil
}

func lineCounts(list []git.Squash) []map[string]int {
	counts := make([]map[string]int, len(list))

	for i, sq := range list {
		counts[i] = map[string]int{sq.Subject: 1}
		for _, line := range sq.Listed {
			counts[i][line]++
		}
	}

	return counts
}

func placeAnswer(
	ctx context.Context, run git.Runner, dir, commit, name, trunk string, pos map[string]int,
) (int, error) {
	if commit == "" {
		return -1, nil
	}

	if !answerSHA.MatchString(commit) {
		return 0, fmt.Errorf("%q: %w", commit, ErrBadAnswer)
	}

	notOnTrunk := fmt.Errorf("%s %w %s: fetch and retry", commit, ErrNotOnTrunk, name)

	sha, ok := git.ResolveCommit(ctx, run, dir, commit)
	if !ok || !git.IsAncestor(ctx, run, dir, sha, trunk) {
		return 0, notOnTrunk
	}

	at, err := landedAt(ctx, run, dir, sha, trunk, pos)
	if err != nil {
		return 0, err
	}

	if at < 0 {
		return 0, notOnTrunk
	}

	return at, nil
}

func prCommit(ctx context.Context, run git.Runner, dir string, pr int, name, trunk string) (string, error) {
	shas, err := git.PRCommits(ctx, run, dir, trunk, pr)
	if err != nil {
		return "", err
	}

	switch len(shas) {
	case 0:
		return "", fmt.Errorf("PR #%d isn't on %s: fetch and retry: %w", pr, name, ErrNotOnTrunk)
	case 1:
		return shas[0], nil
	default:
		return "", &AmbiguousPRError{PR: pr, SHAs: shas}
	}
}

func positions(chain []string) map[string]int {
	pos := make(map[string]int, len(chain))
	for i, sha := range chain {
		pos[sha] = i
	}

	return pos
}

func noGit(q Query) Report {
	return unclassed(Report{Verdict: NoGit}, q)
}

func shallow(trunk string, q Query) Report {
	return unclassed(Report{Trunk: trunk, Branch: mainBranch, Shallow: true, Verdict: CantTell}, q)
}

func unclassed(r Report, q Query) Report {
	r.Bars = make([]BarResult, 0, len(q.Bars))
	r.Unfinished = []string{}

	for _, b := range q.Bars {
		r.Bars = append(r.Bars, BarResult{ID: b.ID, Status: b.Status, Commit: b.Commit})

		if b.Status != statusDone {
			r.Unfinished = append(r.Unfinished, b.ID)
		}
	}

	return r
}

func verdict(counts map[Class]int, unfinished int) Verdict {
	landed := counts[Landed]
	elsewhere := counts[Pushed] + counts[Local]

	switch {
	case landed == 0 && elsewhere == 0:
		return CantTell
	case landed == 0:
		return NotDone
	case elsewhere+counts[Unresolvable]+unfinished == 0:
		return Done
	default:
		return Partly
	}
}

func classify(ctx context.Context, run git.Runner, dir, commit, trunk string, pos map[string]int) (Class, int, error) {
	if commit == "" {
		return NoHash, 0, nil
	}

	if !fullSHA.MatchString(commit) {
		return Unresolvable, 0, nil
	}

	sha, ok := git.ResolveCommit(ctx, run, dir, commit)
	if !ok {
		return Unresolvable, 0, nil
	}

	if git.IsAncestor(ctx, run, dir, sha, trunk) {
		at, err := landedAt(ctx, run, dir, sha, trunk, pos)
		if err != nil || at >= 0 {
			return Landed, at, err
		}
	}

	if git.RemoteContains(ctx, run, dir, sha) {
		return Pushed, 0, nil
	}

	return Local, 0, nil
}

func resolveTrunk(ctx context.Context, run git.Runner, dir string) (string, string, error) {
	if sha, ok := git.ResolveCommit(ctx, run, dir, "refs/remotes/origin/main"); ok {
		return "origin/main", sha, nil
	}

	if sha, ok := git.ResolveCommit(ctx, run, dir, "refs/heads/main"); ok {
		return mainBranch, sha, nil
	}

	return "", "", ErrNoTrunk
}

func landedAt(ctx context.Context, run git.Runner, dir, sha, trunk string, pos map[string]int) (int, error) {
	path, err := git.AncestryPath(ctx, run, dir, sha, trunk)
	if err != nil {
		return 0, fmt.Errorf("landing of %s: %w", sha, err)
	}

	at := -1

	for _, c := range append(path, sha) {
		if i, on := pos[c]; on && i > at {
			at = i
		}
	}

	return at, nil
}
