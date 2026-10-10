package gittest

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Env() []string {
	env := make([]string, 0, len(os.Environ()))

	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "GIT_") {
			env = append(env, kv)
		}
	}

	return env
}

func Run(t testing.TB, dir string, args ...string) string {
	t.Helper()

	bin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found")
	}

	cmd := exec.CommandContext(t.Context(), bin, args...)
	cmd.Dir = dir
	cmd.Env = Env()

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String())
}

func Isolate(t testing.TB) {
	t.Helper()

	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if !strings.HasPrefix(k, "GIT_") {
			continue
		}

		t.Setenv(k, "")

		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("unsetting %s: %v", k, err)
		}
	}
}

const epoch = 1759406400

type Repo struct {
	t      testing.TB
	Dir    string
	Origin string
	n      int
}

func New(t testing.TB) *Repo {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found")
	}

	Isolate(t)

	base := t.TempDir()
	r := &Repo{t: t, Dir: filepath.Join(base, "work"), Origin: filepath.Join(base, "origin.git")}

	Run(t, base, "init", "--bare", "-b", "main", r.Origin)
	Run(t, base, "init", "-b", "main", r.Dir)
	r.Git("remote", "add", "origin", r.Origin)
	r.Commit("chore: root")
	r.Git("push", "-u", "origin", "main")

	return r
}

func (r *Repo) Git(args ...string) string {
	r.t.Helper()

	full := append([]string{"-c", "user.name=bit", "-c", "user.email=bit@example.com"}, args...)

	cmd := exec.CommandContext(r.t.Context(), "git", full...)
	cmd.Dir = r.Dir

	date := fmt.Sprintf("%d +0000", epoch+r.n)
	r.n++

	cmd.Env = append(Env(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		r.t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String())
}

func (r *Repo) Commit(subject string) string {
	r.t.Helper()

	r.n++
	name := fmt.Sprintf("f%d.txt", r.n)

	if err := os.WriteFile(filepath.Join(r.Dir, name), []byte(subject+"\n"), 0o600); err != nil {
		r.t.Fatalf("writing %s: %v", name, err)
	}

	r.Git("add", name)
	r.Git("commit", "-m", subject)

	return r.Git("rev-parse", "HEAD")
}
