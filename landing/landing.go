package landing

import (
	"context"
	"errors"
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
	Done    Verdict = "done"
	NotDone Verdict = "not_done"
)

type Class string

const (
	Landed    Class = "landed"
	NotLanded Class = "not_landed"
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

	r := Report{
		Trunk:      name,
		Branch:     "main",
		Verdict:    Done,
		Bars:       make([]BarResult, 0, len(q.Bars)),
		Unfinished: []string{},
	}

	for _, b := range q.Bars {
		res := BarResult{ID: b.ID, Status: b.Status, Commit: b.Commit, Class: NotLanded}

		if sha, ok := landedAt(ctx, run, q.Dir, b.Commit, trunk); ok {
			res.Class = Landed
			res.Landing = sha
			r.Landing = sha
		} else {
			r.Verdict = NotDone
		}

		if b.Status != statusDone {
			r.Unfinished = append(r.Unfinished, b.ID)
		}

		r.Bars = append(r.Bars, res)
	}

	if r.Verdict != Done {
		r.Landing = ""
	}

	return r, nil
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

func landedAt(ctx context.Context, run git.Runner, dir, commit, trunk string) (string, bool) {
	if !fullSHA.MatchString(commit) {
		return "", false
	}

	sha, ok := git.ResolveCommit(ctx, run, dir, commit)
	if !ok || !git.IsAncestor(ctx, run, dir, sha, trunk) {
		return "", false
	}

	return sha, true
}
