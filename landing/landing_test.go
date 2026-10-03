package landing_test

import (
	"context"
	"errors"
	"path/filepath"
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
	todo = "todo"
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

		if got.Verdict != landing.Partly || got.Landing != a {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, landing.Partly, a)
		}

		if got.Bars[0].Class != landing.Landed || got.Bars[0].Landing != a {
			t.Errorf("Bars[0] = %+v, want class %q landing %q", got.Bars[0], landing.Landed, a)
		}

		if got.Bars[1].Class != landing.Local || got.Bars[1].Landing != "" {
			t.Errorf("Bars[1] = %+v, want class %q landing %q", got.Bars[1], landing.Local, "")
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

		if got.Shallow {
			t.Error("Shallow = true, want false")
		}
	})

	t.Run("a shallow clone is reported and not classed", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Commit("feat(bit): two")
		c := r.Commit("feat(bit): three")
		r.Git("push", "origin", "main")

		sd := filepath.Join(t.TempDir(), "shallow")
		r.Git("clone", "--depth", "1", "file://"+r.Origin, sd)

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: sd, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: done, Commit: c},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if !got.Shallow || got.Verdict != landing.CantTell {
			t.Errorf("Shallow, Verdict = %v, %q, want true, %q", got.Shallow, got.Verdict, landing.CantTell)
		}

		if want := []landing.Class{"", ""}; !slices.Equal(classes(got), want) {
			t.Errorf("classes = %q, want %q", classes(got), want)
		}
	})

	t.Run("no origin makes local main trunk", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("remote", "remove", "origin")
		a := r.Commit("feat(bit): local")

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: a},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Trunk != "main" || got.Branch != "main" {
			t.Errorf("Trunk, Branch = %q, %q, want %q, %q", got.Trunk, got.Branch, "main", "main")
		}

		if got.Verdict != landing.Done || got.Landing != a {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, landing.Done, a)
		}
	})

	t.Run("no main at all is an error", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("remote", "remove", "origin")
		r.Git("branch", "-m", "main", "trunk")

		_, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir})
		if !errors.Is(err, landing.ErrNoTrunk) {
			t.Errorf("err = %v, want %v", err, landing.ErrNoTrunk)
		}
	})

	t.Run("a folder with no git", func(t *testing.T) {
		gittest.Isolate(t)

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: t.TempDir(), Bars: []landing.Bar{
			{ID: bar1, Status: done},
			{ID: bar2, Status: todo},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Verdict != landing.NoGit || got.Trunk != "" {
			t.Errorf("Verdict, Trunk = %q, %q, want %q, \"\"", got.Verdict, got.Trunk, landing.NoGit)
		}

		if want := []string{bar2}; !slices.Equal(got.Unfinished, want) {
			t.Errorf("Unfinished = %q, want %q", got.Unfinished, want)
		}
	})

	t.Run("a malformed commit never reaches git", func(t *testing.T) {
		const trunk = "6a1d3459c0ffee000000000000000000000000ab"

		bad := []string{"--output=/tmp/x", "abc123"}
		fake := &recordingGit{results: map[string]string{
			"rev-parse --git-dir": ".git",
			"rev-parse --verify -q refs/remotes/origin/main^{commit}": trunk,
			"rev-list --first-parent " + trunk:                        trunk,
		}}

		got, err := landing.Check(t.Context(), fake.run, landing.Query{Dir: "/repo", Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: bad[0]},
			{ID: bar2, Status: done, Commit: bad[1]},
		}})
		if err != nil {
			t.Fatal(err)
		}

		for i, b := range got.Bars {
			if b.Class != landing.Unresolvable {
				t.Errorf("Bars[%d].Class = %q, want %q", i, b.Class, landing.Unresolvable)
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

	t.Run("a bar merged through a true merge lands at the merge", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		b1 := r.Commit("feat(bit): on a branch")
		r.Git("checkout", "main")
		r.Commit("chore: trunk moves")
		r.Git("merge", "--no-ff", "-m", "Merge branch 'feat'", "feat")
		m := r.Git("rev-parse", "HEAD")
		r.Git("push", "origin", "main")

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: b1},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Bars[0].Landing != m {
			t.Errorf("Bars[0].Landing = %q, want %q", got.Bars[0].Landing, m)
		}

		if got.Verdict != landing.Done || got.Landing != m {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, landing.Done, m)
		}
	})

	t.Run("a bar from before a back merge lands at the merge", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		b1 := r.Commit("feat(bit): one")
		r.Git("checkout", "main")
		r.Commit("chore: trunk moves")
		r.Git("checkout", "feat")
		r.Git("merge", "--no-ff", "-m", "Merge branch 'main' into feat", "main")
		b2 := r.Commit("feat(bit): two")
		r.Git("checkout", "main")
		r.Git("merge", "--no-ff", "-m", "Merge branch 'feat'", "feat")
		m := r.Git("rev-parse", "HEAD")
		r.Commit("chore: trunk moves again")
		r.Git("push", "origin", "main")

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: b1},
			{ID: bar2, Status: done, Commit: b2},
		}})
		if err != nil {
			t.Fatal(err)
		}

		for i, b := range got.Bars {
			if b.Landing != m {
				t.Errorf("Bars[%d].Landing = %q, want %q", i, b.Landing, m)
			}
		}

		if got.Landing != m {
			t.Errorf("Landing = %q, want %q", got.Landing, m)
		}
	})

	t.Run("a fast forward merge lands each bar at itself", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		b1 := r.Commit("feat(bit): one")
		b2 := r.Commit("feat(bit): two")
		r.Git("checkout", "main")
		r.Git("merge", "--ff-only", "feat")
		r.Git("push", "origin", "main")

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: b1},
			{ID: bar2, Status: done, Commit: b2},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Bars[0].Landing != b1 || got.Bars[1].Landing != b2 {
			t.Errorf("Landings = %q, %q, want %q, %q", got.Bars[0].Landing, got.Bars[1].Landing, b1, b2)
		}

		if got.Landing != b2 {
			t.Errorf("Landing = %q, want %q", got.Landing, b2)
		}
	})

	t.Run("the newest landing wins whatever the bar order", func(t *testing.T) {
		r := gittest.New(t)

		c2 := r.Commit("feat(bit): two")
		c1 := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: c1},
			{ID: bar2, Status: done, Commit: c2},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Landing != c1 {
			t.Errorf("Landing = %q, want %q", got.Landing, c1)
		}
	})

	t.Run("some bars landed and one only pushed to a branch", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")
		r.Git("checkout", "-b", "feat")
		b := r.Commit("feat(bit): two")
		r.Git("push", "origin", "feat")

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: done, Commit: b},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Verdict != landing.Partly || got.Landing != a {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, landing.Partly, a)
		}

		if want := []landing.Class{landing.Landed, landing.Pushed}; !slices.Equal(classes(got), want) {
			t.Errorf("classes = %q, want %q", classes(got), want)
		}
	})

	t.Run("nothing landed and one bar local", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("push", "origin", "main")
		r.Git("checkout", "-b", "feat")
		b := r.Commit("feat(bit): two")

		assertVerdict(t, r.Dir, []landing.Bar{{ID: bar1, Status: done, Commit: b}},
			landing.NotDone, "", []landing.Class{landing.Local})
	})

	t.Run("nothing landed and one bar pushed to a branch", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("push", "origin", "main")
		r.Git("checkout", "-b", "feat")
		b := r.Commit("feat(bit): two")
		r.Git("push", "origin", "feat")

		assertVerdict(t, r.Dir, []landing.Bar{{ID: bar1, Status: done, Commit: b}},
			landing.NotDone, "", []landing.Class{landing.Pushed})
	})

	t.Run("every hash unresolvable", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("push", "origin", "main")

		assertVerdict(t, r.Dir, []landing.Bar{{ID: bar1, Status: done, Commit: "0123456789abcdef0123456789abcdef01234567"}},
			landing.CantTell, "", []landing.Class{landing.Unresolvable})
	})

	t.Run("no bar has a hash", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("push", "origin", "main")

		assertVerdict(t, r.Dir, []landing.Bar{{ID: bar1, Status: done}, {ID: bar2, Status: done}},
			landing.CantTell, "", []landing.Class{landing.NoHash, landing.NoHash})
	})

	t.Run("a bar with no hash beside landed bars", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")

		assertVerdict(t, r.Dir, []landing.Bar{{ID: bar1, Status: done, Commit: a}, {ID: bar2, Status: done}},
			landing.Done, a, []landing.Class{landing.Landed, landing.NoHash})
	})

	t.Run("an unfinished bar is listed", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: todo},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if want := []string{bar2}; !slices.Equal(got.Unfinished, want) {
			t.Errorf("Unfinished = %q, want %q", got.Unfinished, want)
		}
	})

	t.Run("an unfinished bar makes a landed track partly done", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		b := r.Commit("feat(bit): two")
		r.Git("push", "origin", "main")

		got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: r.Dir, Bars: []landing.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: "doing", Commit: b},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Verdict != landing.Partly || got.Landing != b {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, landing.Partly, b)
		}

		if want := []string{bar2}; !slices.Equal(got.Unfinished, want) {
			t.Errorf("Unfinished = %q, want %q", got.Unfinished, want)
		}
	})

	t.Run("a todo bar with no hash", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")

		assertVerdict(t, r.Dir, []landing.Bar{{ID: bar1, Status: done, Commit: a}, {ID: bar2, Status: todo}},
			landing.Partly, a, []landing.Class{landing.Landed, landing.NoHash})
	})
}

func classes(r landing.Report) []landing.Class {
	out := make([]landing.Class, 0, len(r.Bars))
	for _, b := range r.Bars {
		out = append(out, b.Class)
	}

	return out
}

func assertVerdict(
	t *testing.T, dir string, bars []landing.Bar, verdict landing.Verdict, landingSHA string, want []landing.Class,
) {
	t.Helper()

	got, err := landing.Check(t.Context(), git.ExecRunner, landing.Query{Dir: dir, Bars: bars})
	if err != nil {
		t.Fatal(err)
	}

	if got.Verdict != verdict || got.Landing != landingSHA {
		t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, verdict, landingSHA)
	}

	if !slices.Equal(classes(got), want) {
		t.Errorf("classes = %q, want %q", classes(got), want)
	}
}
