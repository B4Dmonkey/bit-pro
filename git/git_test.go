package git

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/git/gittest"
)

const (
	caseGitFails    = "git fails"
	caseEmptyOutput = "empty output"
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
		{name: caseGitFails, result: fakeResult{err: errors.New("exit status 128")}, want: false},
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
		{name: caseEmptyOutput, result: fakeResult{}},
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

func TestIsRepo(t *testing.T) {
	const (
		dir    = "/repo"
		gitDir = "rev-parse --git-dir"
	)

	tests := []struct {
		name   string
		result fakeResult
		want   bool
	}{
		{name: "inside a repo", result: fakeResult{out: ".git\n"}, want: true},
		{name: caseGitFails, result: fakeResult{err: errors.New("exit status 128")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: map[string]fakeResult{gitDir: tt.result}}

			if got := IsRepo(t.Context(), fake.run, dir); got != tt.want {
				t.Errorf("IsRepo() = %v, want %v", got, tt.want)
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
		{name: caseGitFails, result: fakeResult{err: errors.New("exit status 128")}},
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

func TestRemoteContains(t *testing.T) {
	const (
		sha      = "6a1d3459c0ffee000000000000000000000000ab"
		dir      = "/repo"
		contains = "branch -r --contains " + sha
	)

	tests := []struct {
		name   string
		result fakeResult
		want   bool
	}{
		{name: "on a remote branch", result: fakeResult{out: "  origin/feat\n"}, want: true},
		{name: caseEmptyOutput, result: fakeResult{}},
		{name: caseGitFails, result: fakeResult{err: errors.New("exit status 129")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: map[string]fakeResult{contains: tt.result}}

			if got := RemoteContains(t.Context(), fake.run, dir, sha); got != tt.want {
				t.Errorf("RemoteContains() = %v, want %v", got, tt.want)
			}

			if want := []string{dir}; !slices.Equal(fake.dirs, want) {
				t.Errorf("runner dirs = %q, want %q", fake.dirs, want)
			}
		})
	}
}

func TestFirstParents(t *testing.T) {
	const (
		sha     = "6a1d3459c0ffee000000000000000000000000ab"
		other   = "0000000000000000000000000000000000000001"
		dir     = "/repo"
		revList = "rev-list --first-parent origin/main"
	)

	failed := errors.New("exit status 128")

	tests := []struct {
		name    string
		result  fakeResult
		want    []string
		wantErr error
	}{
		{name: "newest first", result: fakeResult{out: sha + "\n" + other + "\n"}, want: []string{sha, other}},
		{name: caseEmptyOutput, result: fakeResult{}, want: []string{}},
		{name: caseGitFails, result: fakeResult{err: failed}, wantErr: failed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: map[string]fakeResult{revList: tt.result}}

			got, err := FirstParents(t.Context(), fake.run, dir, "origin/main")

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("FirstParents() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil && (got == nil || !slices.Equal(got, tt.want)) {
				t.Errorf("FirstParents() = %#v, want %#v", got, tt.want)
			}

			if want := []string{dir}; !slices.Equal(fake.dirs, want) {
				t.Errorf("runner dirs = %q, want %q", fake.dirs, want)
			}
		})
	}
}

func TestAncestryPath(t *testing.T) {
	const (
		sha     = "6a1d3459c0ffee000000000000000000000000ab"
		other   = "0000000000000000000000000000000000000001"
		third   = "0000000000000000000000000000000000000002"
		dir     = "/repo"
		revList = "rev-list --ancestry-path " + sha + "..origin/main"
	)

	failed := errors.New("exit status 128")

	tests := []struct {
		name    string
		result  fakeResult
		want    []string
		wantErr error
	}{
		{name: "newest first", result: fakeResult{out: other + "\n" + third + "\n"}, want: []string{other, third}},
		{name: caseEmptyOutput, result: fakeResult{}, want: []string{}},
		{name: caseGitFails, result: fakeResult{err: failed}, wantErr: failed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: map[string]fakeResult{revList: tt.result}}

			got, err := AncestryPath(t.Context(), fake.run, dir, sha, "origin/main")

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("AncestryPath() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil && (got == nil || !slices.Equal(got, tt.want)) {
				t.Errorf("AncestryPath() = %#v, want %#v", got, tt.want)
			}

			if want := []string{dir}; !slices.Equal(fake.dirs, want) {
				t.Errorf("runner dirs = %q, want %q", fake.dirs, want)
			}
		})
	}
}

func TestPRCommits(t *testing.T) {
	const (
		a   = "6a1d3459c0ffee000000000000000000000000ab"
		b   = "0000000000000000000000000000000000000001"
		c   = "0000000000000000000000000000000000000002"
		dir = "/repo"
		log = "log --first-parent --format=%H%x00%s origin/main"
	)

	failed := errors.New("exit status 128")

	tests := []struct {
		name    string
		result  fakeResult
		want    []string
		wantErr error
	}{
		{
			name:   "keeps subjects ending in the pr number",
			result: fakeResult{out: a + "\x00x (#5)\n" + b + "\x00Revert \"x (#5)\"\n" + c + "\x00y (#15)"},
			want:   []string{a},
		},
		{name: caseEmptyOutput, result: fakeResult{}, want: []string{}},
		{name: caseGitFails, result: fakeResult{err: failed}, wantErr: failed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: map[string]fakeResult{log: tt.result}}

			got, err := PRCommits(t.Context(), fake.run, dir, "origin/main", 5)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("PRCommits() error = %v, want %v", err, tt.wantErr)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("PRCommits() = %#v, want %#v", got, tt.want)
			}

			if want := []string{dir}; !slices.Equal(fake.dirs, want) {
				t.Errorf("runner dirs = %q, want %q", fake.dirs, want)
			}
		})
	}
}

func TestIsShallow(t *testing.T) {
	const (
		dir     = "/repo"
		shallow = "rev-parse --is-shallow-repository"
	)

	tests := []struct {
		name   string
		result fakeResult
		want   bool
	}{
		{name: "a shallow clone", result: fakeResult{out: "true\n"}, want: true},
		{name: "a full clone", result: fakeResult{out: "false\n"}},
		{name: caseGitFails, result: fakeResult{err: errors.New("exit status 128")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: map[string]fakeResult{shallow: tt.result}}

			if got := IsShallow(t.Context(), fake.run, dir); got != tt.want {
				t.Errorf("IsShallow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSquashes(t *testing.T) {
	const (
		a   = "6a1d3459c0ffee000000000000000000000000ab"
		b   = "0000000000000000000000000000000000000001"
		dir = "/repo"
		log = "log --first-parent --format=%H%x00%ct%x00%s%x00%b%x1e origin/main"
	)

	failed := errors.New("exit status 128")

	tests := []struct {
		name    string
		result  fakeResult
		want    []Squash
		wantErr error
	}{
		{
			name: "keeps only the squash",
			result: fakeResult{out: a + "\x001759406402\x00Worktree bit 31 (#5)\x00* one\n\nbody\n\n* two\n  * nested\n\x1e\n" +
				b + "\x001759406401\x00chore: notes\x00* one\n\x1e"},
			want: []Squash{{SHA: a, Time: 1759406402, Subject: "Worktree bit 31", Listed: []string{"one", "two"}}},
		},
		{name: caseEmptyOutput, result: fakeResult{}, want: []Squash{}},
		{name: caseGitFails, result: fakeResult{err: failed}, wantErr: failed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: map[string]fakeResult{log: tt.result}}

			got, err := Squashes(t.Context(), fake.run, dir, "origin/main")

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Squashes() error = %v, want %v", err, tt.wantErr)
			}

			if !slices.EqualFunc(got, tt.want, func(x, y Squash) bool {
				return x.SHA == y.SHA && x.Time == y.Time && x.Subject == y.Subject && slices.Equal(x.Listed, y.Listed)
			}) {
				t.Errorf("Squashes() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSubject(t *testing.T) {
	const (
		dir = "/repo"
		sha = "6a1d3459c0ffee000000000000000000000000ab"
		log = "log -1 --format=%s%x00%ct " + sha
	)

	tests := []struct {
		name        string
		result      fakeResult
		wantSubject string
		wantTime    int64
		wantOK      bool
	}{
		{name: "a commit", result: fakeResult{out: "x\x001759406400"}, wantSubject: "x", wantTime: 1759406400, wantOK: true},
		{name: caseGitFails, result: fakeResult{err: errors.New("exit status 128")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeGit{results: map[string]fakeResult{log: tt.result}}

			subject, at, ok := Subject(t.Context(), fake.run, dir, sha)
			if subject != tt.wantSubject || at != tt.wantTime || ok != tt.wantOK {
				t.Errorf("Subject() = %q, %d, %v, want %q, %d, %v",
					subject, at, ok, tt.wantSubject, tt.wantTime, tt.wantOK)
			}
		})
	}
}
