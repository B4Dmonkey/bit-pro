package mergecheck_test

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/git"
	"github.com/B4Dmonkey/bit-pro/git/gittest"
	"github.com/B4Dmonkey/bit-pro/mergecheck"
)

const (
	bar1 = "BIT-1.1"
	bar2 = "BIT-1.2"
	done = "done"
	todo = "todo"

	play  = "feat(tui): play prompt"
	queue = "feat(tui): queue view"
	tidy  = "chore(bit): tidy"
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

func TestRun(t *testing.T) {
	t.Run("a bar committed but not pushed", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")
		b := r.Commit("feat(bit): two")

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: r.Dir, Bars: []mergecheck.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: done, Commit: b},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Verdict != mergecheck.Partly || got.Landing != a {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, mergecheck.Partly, a)
		}

		if got.Bars[0].Class != mergecheck.Landed || got.Bars[0].Landing != a {
			t.Errorf("Bars[0] = %+v, want class %q landing %q", got.Bars[0], mergecheck.Landed, a)
		}

		if got.Bars[1].Class != mergecheck.Local || got.Bars[1].Landing != "" {
			t.Errorf("Bars[1] = %+v, want class %q landing %q", got.Bars[1], mergecheck.Local, "")
		}
	})

	t.Run("every bar pushed", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		b := r.Commit("feat(bit): two")
		r.Git("push", "origin", "main")

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: r.Dir, Bars: []mergecheck.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: done, Commit: b},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Verdict != mergecheck.Done || got.Landing != b {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, mergecheck.Done, b)
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

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: sd, Bars: []mergecheck.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: done, Commit: c},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if !got.Shallow || got.Verdict != mergecheck.CantTell {
			t.Errorf("Shallow, Verdict = %v, %q, want true, %q", got.Shallow, got.Verdict, mergecheck.CantTell)
		}

		if want := []mergecheck.Class{"", ""}; !slices.Equal(classes(got), want) {
			t.Errorf("classes = %q, want %q", classes(got), want)
		}
	})

	t.Run("no origin makes local main trunk", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("remote", "remove", "origin")
		a := r.Commit("feat(bit): local")

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: r.Dir, Bars: []mergecheck.Bar{
			{ID: bar1, Status: done, Commit: a},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Trunk != "main" || got.Branch != "main" {
			t.Errorf("Trunk, Branch = %q, %q, want %q, %q", got.Trunk, got.Branch, "main", "main")
		}

		if got.Verdict != mergecheck.Done || got.Landing != a {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, mergecheck.Done, a)
		}
	})

	t.Run("no main at all is an error", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("remote", "remove", "origin")
		r.Git("branch", "-m", "main", "trunk")

		_, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: r.Dir})
		if !errors.Is(err, mergecheck.ErrNoTrunk) {
			t.Errorf("err = %v, want %v", err, mergecheck.ErrNoTrunk)
		}
	})

	t.Run("a folder with no git", func(t *testing.T) {
		gittest.Isolate(t)

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: t.TempDir(), Bars: []mergecheck.Bar{
			{ID: bar1, Status: done},
			{ID: bar2, Status: todo},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Verdict != mergecheck.NoGit || got.Trunk != "" {
			t.Errorf("Verdict, Trunk = %q, %q, want %q, \"\"", got.Verdict, got.Trunk, mergecheck.NoGit)
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

		got, err := mergecheck.Run(t.Context(), fake.run, mergecheck.Query{Dir: "/repo", Bars: []mergecheck.Bar{
			{ID: bar1, Status: done, Commit: bad[0]},
			{ID: bar2, Status: done, Commit: bad[1]},
		}})
		if err != nil {
			t.Fatal(err)
		}

		for i, b := range got.Bars {
			if b.Class != mergecheck.Unresolvable {
				t.Errorf("Bars[%d].Class = %q, want %q", i, b.Class, mergecheck.Unresolvable)
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

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: r.Dir, Bars: []mergecheck.Bar{
			{ID: bar1, Status: done, Commit: b1},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Bars[0].Landing != m {
			t.Errorf("Bars[0].Landing = %q, want %q", got.Bars[0].Landing, m)
		}

		if got.Verdict != mergecheck.Done || got.Landing != m {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, mergecheck.Done, m)
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

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: r.Dir, Bars: []mergecheck.Bar{
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

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: r.Dir, Bars: []mergecheck.Bar{
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

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: r.Dir, Bars: []mergecheck.Bar{
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

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: r.Dir, Bars: []mergecheck.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: done, Commit: b},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Verdict != mergecheck.Partly || got.Landing != a {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, mergecheck.Partly, a)
		}

		if want := []mergecheck.Class{mergecheck.Landed, mergecheck.Pushed}; !slices.Equal(classes(got), want) {
			t.Errorf("classes = %q, want %q", classes(got), want)
		}
	})

	t.Run("nothing landed and one bar local", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("push", "origin", "main")
		r.Git("checkout", "-b", "feat")
		b := r.Commit("feat(bit): two")

		assertVerdict(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done, Commit: b}},
			mergecheck.NotDone, "", []mergecheck.Class{mergecheck.Local})
	})

	t.Run("nothing landed and one bar pushed to a branch", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("push", "origin", "main")
		r.Git("checkout", "-b", "feat")
		b := r.Commit("feat(bit): two")
		r.Git("push", "origin", "feat")

		assertVerdict(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done, Commit: b}},
			mergecheck.NotDone, "", []mergecheck.Class{mergecheck.Pushed})
	})

	t.Run("every hash unresolvable", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("push", "origin", "main")

		const missing = "0123456789abcdef0123456789abcdef01234567"

		assertVerdict(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done, Commit: missing}},
			mergecheck.CantTell, "", []mergecheck.Class{mergecheck.Unresolvable})
	})

	t.Run("no bar has a hash", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("push", "origin", "main")

		assertVerdict(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done}, {ID: bar2, Status: done}},
			mergecheck.CantTell, "", []mergecheck.Class{mergecheck.NoHash, mergecheck.NoHash})
	})

	t.Run("a bar with no hash beside landed bars", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")

		assertVerdict(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done, Commit: a}, {ID: bar2, Status: done}},
			mergecheck.Done, a, []mergecheck.Class{mergecheck.Landed, mergecheck.NoHash})
	})

	t.Run("an unfinished bar is listed", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: r.Dir, Bars: []mergecheck.Bar{
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

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: r.Dir, Bars: []mergecheck.Bar{
			{ID: bar1, Status: done, Commit: a},
			{ID: bar2, Status: "doing", Commit: b},
		}})
		if err != nil {
			t.Fatal(err)
		}

		if got.Verdict != mergecheck.Partly || got.Landing != b {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, mergecheck.Partly, b)
		}

		if want := []string{bar2}; !slices.Equal(got.Unfinished, want) {
			t.Errorf("Unfinished = %q, want %q", got.Unfinished, want)
		}
	})

	t.Run("a todo bar with no hash", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")

		assertVerdict(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done, Commit: a}, {ID: bar2, Status: todo}},
			mergecheck.Partly, a, []mergecheck.Class{mergecheck.Landed, mergecheck.NoHash})
	})

	t.Run("a commit answer places bars that haven't landed", func(t *testing.T) {
		for _, tt := range []struct {
			name  string
			class mergecheck.Class
			bar   func(r *gittest.Repo) string
		}{
			{name: "no hash", class: mergecheck.NoHash, bar: func(*gittest.Repo) string { return "" }},
			{name: "stale hash", class: mergecheck.Unresolvable, bar: func(*gittest.Repo) string {
				return "0123456789abcdef0123456789abcdef01234567"
			}},
			{name: "pushed to a branch", class: mergecheck.Pushed, bar: func(r *gittest.Repo) string {
				r.Git("checkout", "-b", "feat")
				b := r.Commit("feat(bit): one")
				r.Git("push", "origin", "feat")
				r.Git("checkout", "main")

				return b
			}},
			{name: "local only", class: mergecheck.Local, bar: func(r *gittest.Repo) string {
				r.Git("checkout", "-b", "feat")
				b := r.Commit("feat(bit): one")
				r.Git("checkout", "main")

				return b
			}},
		} {
			t.Run(tt.name, func(t *testing.T) {
				r := gittest.New(t)

				b := tt.bar(r)
				x := r.Commit("feat(bit): landed by hand")
				r.Git("push", "origin", "main")

				got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{
					Dir: r.Dir, Commit: x[:12], Bars: []mergecheck.Bar{{ID: bar1, Status: done, Commit: b}},
				})
				if err != nil {
					t.Fatal(err)
				}

				if got.Verdict != mergecheck.Done || got.Landing != x {
					t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, mergecheck.Done, x)
				}

				bar := got.Bars[0]
				if bar.Landing != x || !bar.Repoint || bar.Class != tt.class {
					t.Errorf("Bars[0] = %+v, want landing %q, repoint, class %q", bar, x, tt.class)
				}
			})
		}
	})

	t.Run("an answer not on trunk is refused", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		y := r.Commit("feat(bit): off trunk")

		_, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{
			Dir: r.Dir, Commit: y, Bars: []mergecheck.Bar{{ID: bar1, Status: done}},
		})
		if !errors.Is(err, mergecheck.ErrNotOnTrunk) {
			t.Errorf("err = %v, want %v", err, mergecheck.ErrNotOnTrunk)
		}
	})

	t.Run("an answer that isn't a commit is refused", func(t *testing.T) {
		const trunk = "6a1d3459c0ffee000000000000000000000000ab"

		for _, answer := range []string{"--all", "zz"} {
			fake := &recordingGit{results: map[string]string{
				"rev-parse --git-dir": ".git",
				"rev-parse --verify -q refs/remotes/origin/main^{commit}": trunk,
				"rev-list --first-parent " + trunk:                        trunk,
			}}

			_, err := mergecheck.Run(t.Context(), fake.run, mergecheck.Query{
				Dir: "/repo", Commit: answer, Bars: []mergecheck.Bar{{ID: bar1, Status: done}},
			})
			if !errors.Is(err, mergecheck.ErrBadAnswer) {
				t.Errorf("Commit %q: err = %v, want %v", answer, err, mergecheck.ErrBadAnswer)
			}

			for _, call := range fake.calls {
				if strings.Contains(call, answer) {
					t.Errorf("git call %q contains %q", call, answer)
				}
			}
		}
	})

	t.Run("a landed bar keeps its own commit", func(t *testing.T) {
		r := gittest.New(t)

		a := r.Commit("feat(bit): one")
		r.Git("push", "origin", "main")
		r.Git("checkout", "-b", "feat")
		b := r.Commit("feat(bit): two")
		r.Git("push", "origin", "feat")
		r.Git("checkout", "main")
		x := r.Commit("feat(bit): landed by hand")
		r.Git("push", "origin", "main")

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{
			Dir: r.Dir, Commit: x, Bars: []mergecheck.Bar{
				{ID: bar1, Status: done, Commit: a},
				{ID: bar2, Status: done, Commit: b},
				{ID: "BIT-1.3", Status: done},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		if bar := got.Bars[0]; bar.Repoint || bar.Landing != a {
			t.Errorf("Bars[0] = %+v, want no repoint, landing %q", bar, a)
		}

		for i, bar := range got.Bars[1:] {
			if !bar.Repoint || bar.Landing != x {
				t.Errorf("Bars[%d] = %+v, want repoint, landing %q", i+1, bar, x)
			}
		}
	})

	t.Run("a pr number finds its squash on trunk", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		r.Commit("feat(tui): overlay")
		r.Git("checkout", "main")
		r.Git("merge", "--squash", "feat")
		r.Git("commit", "-m", "Worktree bit 31 (#5)")
		s := r.Git("rev-parse", "HEAD")
		r.Git("push", "origin", "main")

		got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{
			Dir: r.Dir, PR: 5, Bars: []mergecheck.Bar{{ID: bar1, Status: done}},
		})
		if err != nil {
			t.Fatal(err)
		}

		if got.Verdict != mergecheck.Done || got.Landing != s {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, mergecheck.Done, s)
		}

		if !got.Bars[0].Repoint {
			t.Errorf("Bars[0] = %+v, want repoint", got.Bars[0])
		}
	})

	t.Run("a pr that isn't on trunk is refused", func(t *testing.T) {
		r := gittest.New(t)

		r.Commit("feat(bit): one (#5)")
		r.Git("push", "origin", "main")

		assertPRNotOnTrunk(t, r.Dir, 6)
	})

	t.Run("two commits ending in the same pr number are refused", func(t *testing.T) {
		r := gittest.New(t)

		x := r.Commit("x (#7)")
		y := r.Commit("y (#7)")
		r.Git("push", "origin", "main")

		_, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{
			Dir: r.Dir, PR: 7, Bars: []mergecheck.Bar{{ID: bar1, Status: done}},
		})

		var ambiguous *mergecheck.AmbiguousPRError
		if !errors.As(err, &ambiguous) {
			t.Fatalf("err = %v, want *AmbiguousPRError", err)
		}

		if !slices.Contains(ambiguous.SHAs, x) || !slices.Contains(ambiguous.SHAs, y) || len(ambiguous.SHAs) != 2 {
			t.Errorf("SHAs = %q, want %q and %q", ambiguous.SHAs, x, y)
		}
	})

	t.Run("pr 1 doesn't match pr 17", func(t *testing.T) {
		r := gittest.New(t)

		r.Commit("z (#17)")
		r.Git("push", "origin", "main")

		assertPRNotOnTrunk(t, r.Dir, 1)
	})

	t.Run("a bar listed in a multi commit squash lands at the squash", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		b1 := r.Commit(play)
		b2 := r.Commit(queue)
		r.Git("push", "origin", "feat")
		s := squash(r, "Worktree bit 31 (#5)", "* "+play+"\n\nbody one\n\n* "+queue+"\n\nbody two")

		got := check(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done, Commit: b1}, {ID: bar2, Status: done, Commit: b2}})

		if got.Verdict != mergecheck.Done || got.Landing != s {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, mergecheck.Done, s)
		}

		assertSquashed(t, got, s, s)
	})

	t.Run("a single commit squash matches on its subject", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		c := r.Commit("fix(tui): render overlay")
		r.Git("push", "origin", "feat")
		s := squash(r, "fix(tui): render overlay (#6)", "")

		assertSquashed(t, check(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done, Commit: c}}), s)
	})

	t.Run("the earliest squash after the bar wins", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		b1 := r.Commit(play)
		r.Git("push", "origin", "feat")
		first := squash(r, "Worktree bit 31 (#16)", "* "+play)
		r.Git("checkout", "feat")
		r.Commit("chore: more")
		squash(r, "Worktree bit 31 (#17)", "* "+play)

		assertSquashed(t, check(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done, Commit: b1}}), first)
	})

	t.Run("a track over two squashes lands at the newer", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		b1 := r.Commit(play)
		b2 := r.Commit(queue)
		r.Git("push", "origin", "feat")
		five := squash(r, "Worktree bit 31 (#5)", "* "+play+"\n\n* "+queue)
		r.Git("checkout", "feat")
		b3 := r.Commit("feat(tui): stop view")
		r.Git("push", "origin", "feat")
		six := squash(r, "feat(tui): stop view (#6)", "")

		got := check(t, r.Dir, []mergecheck.Bar{
			{ID: bar1, Status: done, Commit: b1},
			{ID: bar2, Status: done, Commit: b2},
			{ID: "BIT-1.3", Status: done, Commit: b3},
		})

		if got.Verdict != mergecheck.Done || got.Landing != six {
			t.Errorf("Verdict, Landing = %q, %q, want %q, %q", got.Verdict, got.Landing, mergecheck.Done, six)
		}

		assertSquashed(t, got, five, five, six)
	})

	t.Run("two bars with one shared subject claim one line each", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		t1 := r.Commit(tidy)
		t2 := r.Commit(tidy)
		r.Git("push", "origin", "feat")
		s := squash(r, "Tidy (#8)", "* "+tidy)

		got := check(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done, Commit: t1}, {ID: bar2, Status: done, Commit: t2}})

		if got.Verdict != mergecheck.Partly || got.Bars[0].Landing != s {
			t.Errorf("Verdict, Bars[0].Landing = %q, %q, want %q, %q", got.Verdict, got.Bars[0].Landing, mergecheck.Partly, s)
		}

		if want := []mergecheck.Class{mergecheck.Squash, mergecheck.Pushed}; !slices.Equal(classes(got), want) {
			t.Errorf("classes = %q, want %q", classes(got), want)
		}
	})

	t.Run("a subject listed twice lands both bars", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		t1 := r.Commit(tidy)
		t2 := r.Commit(tidy)
		r.Git("push", "origin", "feat")
		s := squash(r, "Tidy (#8)", "* "+tidy+"\n\n* "+tidy)

		assertSquashed(t, check(t, r.Dir, []mergecheck.Bar{
			{ID: bar1, Status: done, Commit: t1}, {ID: bar2, Status: done, Commit: t2},
		}), s, s)
	})

	t.Run("a bar passed over moves to a later squash", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		t1 := r.Commit(tidy)
		t2 := r.Commit(tidy)
		r.Git("push", "origin", "feat")
		eight := squash(r, "Tidy (#8)", "* "+tidy)
		r.Git("checkout", "feat")
		r.Commit("chore: more")
		nine := squash(r, "Tidy again (#9)", "* "+tidy)

		assertSquashed(t, check(t, r.Dir, []mergecheck.Bar{
			{ID: bar1, Status: done, Commit: t1}, {ID: bar2, Status: done, Commit: t2},
		}), eight, nine)
	})

	t.Run("a mention that isn't a list line doesn't match", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		b1 := r.Commit(play)
		r.Git("push", "origin", "feat")
		squash(r, "Worktree bit 31 (#5)", "see "+play)

		assertVerdict(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done, Commit: b1}},
			mergecheck.NotDone, "", []mergecheck.Class{mergecheck.Pushed})
	})

	t.Run("a direct commit with a matching bullet isn't a squash", func(t *testing.T) {
		r := gittest.New(t)

		r.Git("checkout", "-b", "feat")
		b1 := r.Commit(play)
		r.Git("push", "origin", "feat")
		r.Git("checkout", "main")
		r.Git("commit", "--allow-empty", "-m", "chore: notes", "-m", "* "+play)
		r.Git("push", "origin", "main")

		assertVerdict(t, r.Dir, []mergecheck.Bar{{ID: bar1, Status: done, Commit: b1}},
			mergecheck.NotDone, "", []mergecheck.Class{mergecheck.Pushed})
	})
}

func squash(r *gittest.Repo, subject, body string) string {
	r.Git("checkout", "main")
	r.Git("merge", "--squash", "feat")

	args := []string{"commit", "--allow-empty", "-m", subject}
	if body != "" {
		args = append(args, "-m", body)
	}

	r.Git(args...)
	r.Git("push", "origin", "main")

	return r.Git("rev-parse", "HEAD")
}

func check(t *testing.T, dir string, bars []mergecheck.Bar) mergecheck.Report {
	t.Helper()

	got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: dir, Bars: bars})
	if err != nil {
		t.Fatal(err)
	}

	return got
}

func assertSquashed(t *testing.T, got mergecheck.Report, want ...string) {
	t.Helper()

	landings := make([]string, 0, len(got.Bars))

	for i, bar := range got.Bars {
		landings = append(landings, bar.Landing)

		if bar.Class != mergecheck.Squash || !bar.Repoint {
			t.Errorf("Bars[%d] = %+v, want class %q, repoint", i, bar, mergecheck.Squash)
		}
	}

	if !slices.Equal(landings, want) {
		t.Errorf("bar landings = %q, want %q", landings, want)
	}
}

func assertPRNotOnTrunk(t *testing.T, dir string, pr int) {
	t.Helper()

	_, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{
		Dir: dir, PR: pr, Bars: []mergecheck.Bar{{ID: bar1, Status: done}},
	})
	if !errors.Is(err, mergecheck.ErrNotOnTrunk) {
		t.Errorf("PR %d: err = %v, want %v", pr, err, mergecheck.ErrNotOnTrunk)
	}
}

func classes(r mergecheck.Report) []mergecheck.Class {
	out := make([]mergecheck.Class, 0, len(r.Bars))
	for _, b := range r.Bars {
		out = append(out, b.Class)
	}

	return out
}

func assertVerdict(
	t *testing.T, dir string, bars []mergecheck.Bar,
	verdict mergecheck.Verdict, landingSHA string, want []mergecheck.Class,
) {
	t.Helper()

	got, err := mergecheck.Run(t.Context(), git.ExecRunner, mergecheck.Query{Dir: dir, Bars: bars})
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
