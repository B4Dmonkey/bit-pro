package git

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/git/gittest"
)

type fakeResult struct {
	out string
	err error
}

type fakeGit struct {
	results map[string]fakeResult
	dirs    []string
}

func (f *fakeGit) run(_ context.Context, dir string, args ...string) (string, error) {
	f.dirs = append(f.dirs, dir)

	r, ok := f.results[strings.Join(args, " ")]
	if !ok {
		return "", errors.New("unexpected git " + strings.Join(args, " "))
	}

	return r.out, r.err
}

func TestReadHead(t *testing.T) {
	const (
		sha         = "6a1d3459c0ffee000000000000000000000000ab"
		dir         = "/repo/.claude/worktrees/wt"
		revParse    = "rev-parse HEAD"
		symbolicRef = "symbolic-ref --short -q HEAD"
	)

	failed := errors.New("exit status 128")

	tests := []struct {
		name    string
		results map[string]fakeResult
		want    Head
	}{
		{
			name: "on a branch",
			results: map[string]fakeResult{
				revParse:    {out: sha + "\n"},
				symbolicRef: {out: "v2\n"},
			},
			want: Head{SHA: sha, Branch: "v2"},
		},
		{
			name: "detached head",
			results: map[string]fakeResult{
				revParse:    {out: sha + "\n"},
				symbolicRef: {err: failed},
			},
			want: Head{SHA: sha},
		},
		{
			name: "no commits yet",
			results: map[string]fakeResult{
				revParse:    {err: failed},
				symbolicRef: {out: "main\n"},
			},
			want: Head{Branch: "main"},
		},
		{
			name: "not a git folder",
			results: map[string]fakeResult{
				revParse:    {err: failed},
				symbolicRef: {err: failed},
			},
			want: Head{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: tt.results}

			got := ReadHead(t.Context(), fake.run, dir)

			if got != tt.want {
				t.Errorf("ReadHead() = %+v, want %+v", got, tt.want)
			}

			if want := []string{dir, dir}; !slices.Equal(fake.dirs, want) {
				t.Errorf("runner dirs = %q, want %q", fake.dirs, want)
			}
		})
	}
}

func TestExecRunner(t *testing.T) {
	t.Run("reads a real repo", func(t *testing.T) {
		gittest.Isolate(t)

		dir := t.TempDir()
		gittest.Run(t, dir, "init", "-b", "main")
		gittest.Run(t, dir, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "--allow-empty", "-m", "x")

		got := ReadHead(t.Context(), ExecRunner, dir)

		want := Head{SHA: gittest.Run(t, dir, "rev-parse", "HEAD"), Branch: "main"}
		if got != want {
			t.Errorf("ReadHead() = %+v, want %+v", got, want)
		}

		if len(got.SHA) != 40 {
			t.Errorf("SHA = %q, want 40 characters", got.SHA)
		}
	})

	t.Run("a failing command returns an error", func(t *testing.T) {
		gittest.Isolate(t)

		_, err := ExecRunner(t.Context(), t.TempDir(), "rev-parse", "HEAD")

		if err == nil || !strings.Contains(err.Error(), "rev-parse") {
			t.Errorf("ExecRunner() error = %v, want one containing rev-parse", err)
		}
	})
}

func TestTracks(t *testing.T) {
	const (
		dir     = "/repo"
		lsFiles = "ls-files -- .bit"
	)

	tests := []struct {
		name   string
		result fakeResult
		want   bool
	}{
		{name: "tracked files", result: fakeResult{out: ".bit/config.toml\n"}, want: true},
		{name: "nothing tracked", result: fakeResult{}, want: false},
		{name: "git fails", result: fakeResult{err: errors.New("exit status 128")}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: map[string]fakeResult{lsFiles: tt.result}}

			if got := Tracks(t.Context(), fake.run, dir, ".bit"); got != tt.want {
				t.Errorf("Tracks() = %v, want %v", got, tt.want)
			}

			if want := []string{dir}; !slices.Equal(fake.dirs, want) {
				t.Errorf("runner dirs = %q, want %q", fake.dirs, want)
			}
		})
	}
}

func TestResolveCommit(t *testing.T) {
	const (
		sha      = "6a1d3459c0ffee000000000000000000000000ab"
		dir      = "/repo"
		rev      = "refs/remotes/origin/main"
		revParse = "rev-parse --verify -q " + rev + "^{commit}"
	)

	tests := []struct {
		name   string
		result fakeResult
		want   string
		wantOK bool
	}{
		{name: "resolves", result: fakeResult{out: sha + "\n"}, want: sha, wantOK: true},
		{name: "rev-parse fails", result: fakeResult{err: errors.New("exit status 1")}},
		{name: "empty output", result: fakeResult{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: map[string]fakeResult{revParse: tt.result}}

			got, ok := ResolveCommit(t.Context(), fake.run, dir, rev)

			if got != tt.want || ok != tt.wantOK {
				t.Errorf("ResolveCommit() = (%q, %v), want (%q, %v)", got, ok, tt.want, tt.wantOK)
			}

			if want := []string{dir}; !slices.Equal(fake.dirs, want) {
				t.Errorf("runner dirs = %q, want %q", fake.dirs, want)
			}
		})
	}
}

func TestIsAncestor(t *testing.T) {
	const (
		sha       = "6a1d3459c0ffee000000000000000000000000ab"
		other     = "0000000000000000000000000000000000000001"
		dir       = "/repo"
		mergeBase = "merge-base " + sha + " origin/main"
	)

	tests := []struct {
		name   string
		result fakeResult
		want   bool
	}{
		{name: "ancestor", result: fakeResult{out: sha + "\n"}, want: true},
		{name: "another merge base", result: fakeResult{out: other + "\n"}},
		{name: "git fails", result: fakeResult{err: errors.New("exit status 128")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: map[string]fakeResult{mergeBase: tt.result}}

			if got := IsAncestor(t.Context(), fake.run, dir, sha, "origin/main"); got != tt.want {
				t.Errorf("IsAncestor() = %v, want %v", got, tt.want)
			}

			if want := []string{dir}; !slices.Equal(fake.dirs, want) {
				t.Errorf("runner dirs = %q, want %q", fake.dirs, want)
			}
		})
	}
}
