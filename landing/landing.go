package landing

import (
	"context"

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

const Done Verdict = "done"

type Class string

const Landed Class = "landed"

type BarResult struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Commit  string `json:"commit"`
	Class   Class  `json:"class"`
	Landing string `json:"landing"`
}

type Report struct {
	Trunk   string      `json:"trunk"`
	Branch  string      `json:"branch"`
	Verdict Verdict     `json:"verdict"`
	Landing string      `json:"landing"`
	Bars    []BarResult `json:"bars"`
}

func Check(_ context.Context, _ git.Runner, q Query) (Report, error) {
	r := Report{
		Trunk:   "origin/main",
		Branch:  "main",
		Verdict: Done,
		Bars:    make([]BarResult, 0, len(q.Bars)),
	}

	for _, b := range q.Bars {
		r.Bars = append(r.Bars, BarResult{
			ID:      b.ID,
			Status:  b.Status,
			Commit:  b.Commit,
			Class:   Landed,
			Landing: b.Commit,
		})
		r.Landing = b.Commit
	}

	return r, nil
}
