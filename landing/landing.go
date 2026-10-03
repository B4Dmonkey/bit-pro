package landing

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/B4Dmonkey/bit-pro/git"
)

type Bar struct {
	ID     string
	Status string
	Commit string
}

type Query struct {
	Dir  string
	Bars []Bar
}

type Verdict string

const (
	Done     Verdict = "done"
	NotDone  Verdict = "not_done"
	Partly   Verdict = "partly"
	CantTell Verdict = "cant_tell"
)

type Class string

const (
	Landed       Class = "landed"
	Pushed       Class = "pushed"
	Local        Class = "local"
	Unresolvable Class = "unresolvable"
	NoHash       Class = "no_hash"
)

type BarResult struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Commit  string `json:"commit"`
	Class   Class  `json:"class"`
	Landing string `json:"landing"`
}

type Report struct {
	Trunk      string      `json:"trunk"`
	Branch     string      `json:"branch"`
	Verdict    Verdict     `json:"verdict"`
	Landing    string      `json:"landing"`
	Bars       []BarResult `json:"bars"`
	Unfinished []string    `json:"unfinished"`
}

const statusDone = "done"

var (
	fullSHA = regexp.MustCompile(`^[0-9a-f]{40}$`)

	ErrNoTrunk = errors.New("no trunk: neither origin/main nor main exists")
)

func Check(ctx context.Context, run git.Runner, q Query) (Report, error) {
	name, trunk, err := resolveTrunk(ctx, run, q.Dir)
	if err != nil {
		return Report{}, err
	}

	chain, err := git.FirstParents(ctx, run, q.Dir, trunk)
	if err != nil {
		return Report{}, fmt.Errorf("reading trunk %s: %w", name, err)
	}

	pos := make(map[string]int, len(chain))
	for i, sha := range chain {
		pos[sha] = i
	}

	newest := len(chain)

	r := Report{
		Trunk:      name,
		Branch:     "main",
		Bars:       make([]BarResult, 0, len(q.Bars)),
		Unfinished: []string{},
	}

	counts := map[Class]int{}

	for _, b := range q.Bars {
		res := BarResult{ID: b.ID, Status: b.Status, Commit: b.Commit}

		class, at, err := classify(ctx, run, q.Dir, b.Commit, trunk, pos)
		if err != nil {
			return Report{}, err
		}

		res.Class = class
		counts[class]++

		if class == Landed {
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
		return "main", sha, nil
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
