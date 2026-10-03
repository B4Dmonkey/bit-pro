package project

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
)

const (
	testCode  = "BIT"
	outerCode = "ACME"
	innerCode = "API"
)

func evalSymlinks(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks(%q) returned error: %v", path, err)
	}

	return resolved
}

func mkdirs(t *testing.T, root string, dirs ...string) {
	t.Helper()

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o750); err != nil {
			t.Fatalf("os.MkdirAll(%q) returned error: %v", d, err)
		}
	}
}

func TestResolve(t *testing.T) {
	single := func(code, sub string) func(root, real string) []Project {
		return func(root, _ string) []Project { return []Project{{Code: code, Path: filepath.Join(root, sub)}} }
	}

	nested := func(root, _ string) []Project {
		return []Project{{Code: innerCode, Path: filepath.Join(root, "api")}, {Code: outerCode, Path: root}}
	}

	under := func(sub string) func(root, real string) string {
		return func(root, _ string) string { return filepath.Join(root, sub) }
	}

	tests := []struct {
		name     string
		dirs     []string
		projects func(root, real string) []Project
		dir      func(root, real string) string
		wantCode string
		wantErr  error
		wantMsg  string
	}{
		{
			name:     "finds the project from a subfolder",
			dirs:     []string{"src/pkg"},
			projects: single(testCode, ""),
			dir:      under("src/pkg"),
			wantCode: testCode,
		},
		{
			name:     "sibling prefix is not a match",
			dirs:     []string{"app", "application"},
			projects: single(testCode, "app"),
			dir:      under("application"),
			wantErr:  ErrNotRegistered,
		},
		{
			name:     "nested registration longest wins",
			dirs:     []string{"api/x"},
			projects: nested,
			dir:      under("api/x"),
			wantCode: innerCode,
		},
		{
			name:     "nested registration outer folder",
			dirs:     []string{"api", "docs"},
			projects: nested,
			dir:      under("docs"),
			wantCode: outerCode,
		},
		{
			name: "removed longest match is refused",
			dirs: []string{"api/x"},
			projects: func(root, _ string) []Project {
				return []Project{{Code: outerCode, Path: root}, {Code: innerCode, Path: filepath.Join(root, "api"), Removed: true}}
			},
			dir:     under("api/x"),
			wantErr: ErrRemoved,
			wantMsg: "removed; run `bp add`",
		},
		{
			name:     "claude worktree",
			dirs:     []string{".claude/worktrees/hazy/sub"},
			projects: single(testCode, ""),
			dir:      under(".claude/worktrees/hazy/sub"),
			wantCode: testCode,
		},
		{
			name:     "missing dir",
			projects: single(testCode, ""),
			dir:      under(".claude/worktrees/wt"),
			wantCode: testCode,
		},
		{
			name:     "symlinked temp dir as registered",
			projects: single(testCode, ""),
			dir:      func(_, real string) string { return real },
			wantCode: testCode,
		},
		{
			name:     "symlinked temp dir as typed",
			projects: func(_, real string) []Project { return []Project{{Code: testCode, Path: real}} },
			dir:      func(root, _ string) string { return root },
			wantCode: testCode,
		},
		{
			name:     "case differs in the typed path",
			dirs:     []string{"Repo/x"},
			projects: single(testCode, "Repo"),
			dir:      under("repo/x"),
			wantCode: testCode,
		},
		{
			name:     "unregistered with bit above",
			dirs:     []string{".bit", "sub"},
			projects: func(_, _ string) []Project { return nil },
			dir:      under("sub"),
			wantErr:  ErrNeedsMigrate,
			wantMsg:  "run `bp migrate`",
		},
		{
			name:     "unregistered without bit",
			dirs:     []string{"sub"},
			projects: func(_, _ string) []Project { return nil },
			dir:      under("sub"),
			wantErr:  ErrNotRegistered,
			wantMsg:  "not a bit project; run `bp add`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			mkdirs(t, root, tt.dirs...)
			real := evalSymlinks(t, root)

			got, err := Resolve(tt.projects(root, real), tt.dir(root, real))

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Resolve() error = %v, want %v", err, tt.wantErr)
			}

			if err != nil && !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("Resolve() error = %q, want it to contain %q", err, tt.wantMsg)
			}

			if got.Code != tt.wantCode {
				t.Errorf("Resolve() code = %q, want %q", got.Code, tt.wantCode)
			}
		})
	}
}

func TestCanonicalPath(t *testing.T) {
	tests := []struct {
		name string
		path func(t *testing.T, root string) string
		want func(real string) string
	}{
		{
			name: "existing dir resolves symlinks",
			path: func(_ *testing.T, root string) string { return root },
			want: func(real string) string { return real },
		},
		{
			name: "missing relative path resolves its existing parent",
			path: func(t *testing.T, root string) string {
				t.Chdir(root)
				return "a/../b/c"
			},
			want: func(real string) string { return filepath.Join(real, "b", "c") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			want := tt.want(evalSymlinks(t, root))

			got, err := CanonicalPath(tt.path(t, root))
			if err != nil {
				t.Fatalf("CanonicalPath() returned error: %v", err)
			}

			if got != want {
				t.Errorf("CanonicalPath() = %q, want %q", got, want)
			}
		})
	}
}

func registerRoot(t *testing.T) (*orm.Queries, string) {
	t.Helper()

	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", "")

	root := t.TempDir()
	mkdirs(t, root, "sub")

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}

	t.Cleanup(func() { sqlDB.Close() })

	q := orm.New(sqlDB)
	if err := q.CreateProject(t.Context(), orm.CreateProjectParams{Path: root, Code: testCode}); err != nil {
		t.Fatalf("CreateProject() returned error: %v", err)
	}

	return q, root
}

func TestFind(t *testing.T) {
	t.Run("resolves through the registry", func(t *testing.T) {
		_, root := registerRoot(t)

		got, err := Find(t.Context(), filepath.Join(root, "sub"))
		if err != nil {
			t.Fatalf("Find() returned error: %v", err)
		}

		if got.Code != testCode {
			t.Errorf("Find() code = %q, want %q", got.Code, testCode)
		}
	})
	t.Run("refuses a removed project", func(t *testing.T) {
		q, root := registerRoot(t)

		projects, err := q.ListProjects(t.Context())
		if err != nil {
			t.Fatalf("ListProjects() returned error: %v", err)
		}

		if err := q.SetProjectRemoved(t.Context(), orm.SetProjectRemovedParams{Removed: 1, ID: projects[0].ID}); err != nil {
			t.Fatalf("SetProjectRemoved() returned error: %v", err)
		}

		if _, err := Find(t.Context(), filepath.Join(root, "sub")); !errors.Is(err, ErrRemoved) {
			t.Errorf("Find() error = %v, want %v", err, ErrRemoved)
		}
	})
}
