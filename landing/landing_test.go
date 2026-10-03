package landing_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/git"
	"github.com/B4Dmonkey/bit-pro/git/gittest"
	"github.com/B4Dmonkey/bit-pro/landing"
)

const (
	bar1 = "BIT-1.1"
	bar2 = "BIT-1.2"
	done = "done"
)

type recordingGit struct {
	results map[string]string
	calls   []string
}

func (f *recordingGit) run(_ context.Context, _ string, args ...string) (string, error) {
	call := strings.Join(args, " ")
	f.calls = append(f.calls, call)

	out, ok := f.results[call]
	if !ok {
		return "", errors.New("unexpected git " + call)
	}

	return out, nil
}

func TestCheck(t *testing.T) {
	t.Run("a bar committed but not pushed", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")
		b := r.Commit("feat(bit): two")

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: done, Commit: b},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Verdict != landing.NotDone || got.Landing != "" {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, landing.NotDone, "")
		}

		if got.Bars[0].Class != landing.Landed || got.Bars[0].Landing != a {
			t.Errorf("Bars[0] = %+v, want class %q landing %q", got.Bars[0], landing.Landed, a)
		}

		if got.Bars[1].Class != landing.NotLanded || got.Bars[1].Landing != "" {
			t.Errorf("Bars[1] = %+v, want class %q landing %q", got.Bars[1], landing.NotLanded, "")
		}
	})

	t.Run("every bar pushed", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		b := r.Commit("feat(bit): two")
		r.Git("push", "origin", "main")

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: done, Commit: b},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Verdict != landing.Done || got.Landing != b {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, landing.Done, b)
		}
	})

	t.Run("a malformed commit never reaches git", func(t *testing.T) {
		const trunk = "6a1d3459c0ffee000000000000000000000000ab"

		bad := []string{"--output=/tmp/x", "abc123"}
		fake := &recordingGit{results: map[string]string{
			"rev-parse --verify -q refs/remotes/origin/main^{commit}": trunk,
		}}

		got, err := landing.Check(t.Context(), fake.run, landing.Query{Dir: "/repo", Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: bad[0]},
			{ID: bar2, Status: done, Commit: bad[1]},
		}})
		if err != nil {
			t.Fatal(err)
		}

		for i, b := range got.Bars {
			if b.Class != landing.NotLanded {
				t.Errorf("Bars[%d].Class = %q, want %q", i, b.Class, landing.NotLanded)
			}
		}

		for _, call := range fake.calls {
			for _, v := range bad {
				if strings.Contains(call, v) {
					t.Errorf("git call %q contains %q", call, v)
				}
			}
		}
	})

	t.Run("an unfinished bar is listed", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: "todo"},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if want := []string{bar2}; !slices.Equal(got.Unfinished, want) {
			t.Errorf("Unfinished = %q, want %q", got.Unfinished, want)
		}
	})
}
