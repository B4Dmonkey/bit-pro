package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Runner func(ctx context.Context, dir string, args ...string) (string, error)

func ExecRunner(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return strings.TrimSpace(stdout.String()), nil
}

type Head struct {
	SHA    string
	Branch string
}

func ReadHead(ctx context.Context, run Runner, dir string) Head {
	var h Head

	if sha, err := run(ctx, dir, "rev-parse", "HEAD"); err == nil {
		h.SHA = strings.TrimSpace(sha)
	}

	if branch, err := run(ctx, dir, "symbolic-ref", "--short", "-q", "HEAD"); err == nil {
		h.Branch = strings.TrimSpace(branch)
	}

	return h
}

func Tracks(ctx context.Context, run Runner, dir, path string) bool {
	out, err := run(ctx, dir, "ls-files", "--", path)

	return err == nil && strings.TrimSpace(out) != ""
}

func IsRepo(ctx context.Context, run Runner, dir string) bool {
	_, err := run(ctx, dir, "rev-parse", "--git-dir")

	return err == nil
}

func ResolveCommit(ctx context.Context, run Runner, dir, rev string) (string, bool) {
	out, err := run(ctx, dir, "rev-parse", "--verify", "-q", rev+"^{commit}")
	if err != nil {
		return "", false
	}

	sha := strings.TrimSpace(out)

	return sha, sha != ""
}

func IsAncestor(ctx context.Context, run Runner, dir, sha, ref string) bool {
	out, err := run(ctx, dir, "merge-base", sha, ref)

	return err == nil && strings.TrimSpace(out) == sha
}

func FirstParents(ctx context.Context, run Runner, dir, ref string) ([]string, error) {
	out, err := run(ctx, dir, "rev-list", "--first-parent", ref)
	if err != nil {
		return nil, fmt.Errorf("first parents of %s: %w", ref, err)
	}

	return strings.Fields(out), nil
}

func AncestryPath(ctx context.Context, run Runner, dir, sha, ref string) ([]string, error) {
	out, err := run(ctx, dir, "rev-list", "--ancestry-path", sha+".."+ref)
	if err != nil {
		return nil, fmt.Errorf("ancestry path %s..%s: %w", sha, ref, err)
	}

	return strings.Fields(out), nil
}

func RemoteContains(ctx context.Context, run Runner, dir, sha string) bool {
	out, err := run(ctx, dir, "branch", "-r", "--contains", sha)

	return err == nil && strings.TrimSpace(out) != ""
}

func IsShallow(ctx context.Context, run Runner, dir string) bool {
	out, err := run(ctx, dir, "rev-parse", "--is-shallow-repository")

	return err == nil && strings.TrimSpace(out) == "true"
}

func PRCommits(ctx context.Context, run Runner, dir, ref string, n int) ([]string, error) {
	out, err := run(ctx, dir, "log", "--first-parent", "--format=%H%x00%s", ref)
	if err != nil {
		return nil, fmt.Errorf("pr #%d commits on %s: %w", n, ref, err)
	}

	suffix := fmt.Sprintf(" (#%d)", n)
	shas := []string{}

	for line := range strings.Lines(out) {
		sha, subject, ok := strings.Cut(strings.TrimSuffix(line, "\n"), "\x00")
		if ok && strings.HasSuffix(subject, suffix) {
			shas = append(shas, sha)
		}
	}

	return shas, nil
}

type Squash struct {
	SHA     string
	Time    int64
	Subject string
	Listed  []string
}

var prSuffix = regexp.MustCompile(` \(#\d+\)$`)

func Squashes(ctx context.Context, run Runner, dir, ref string) ([]Squash, error) {
	out, err := run(ctx, dir, "log", "--first-parent", "--format=%H%x00%ct%x00%s%x00%b%x1e", ref)
	if err != nil {
		return nil, fmt.Errorf("squashes on %s: %w", ref, err)
	}

	squashes := []Squash{}

	for record := range strings.SplitSeq(out, "\x1e") {
		fields := strings.SplitN(strings.TrimSpace(record), "\x00", 4)
		if len(fields) < 3 || !prSuffix.MatchString(fields[2]) {
			continue
		}

		at, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("squash %s time %q: %w", fields[0], fields[1], err)
		}

		s := Squash{SHA: fields[0], Time: at, Subject: prSuffix.ReplaceAllString(fields[2], "")}

		if len(fields) == 4 {
			s.Listed = listed(fields[3])
		}

		squashes = append(squashes, s)
	}

	return squashes, nil
}

func listed(body string) []string {
	var out []string

	for line := range strings.Lines(body) {
		if item, ok := strings.CutPrefix(line, "* "); ok {
			out = append(out, strings.TrimRight(item, " \t\r\n"))
		}
	}

	return out
}

func Subject(ctx context.Context, run Runner, dir, sha string) (string, int64, bool) {
	out, err := run(ctx, dir, "log", "-1", "--format=%s%x00%ct", sha)
	if err != nil {
		return "", 0, false
	}

	subject, ct, ok := strings.Cut(strings.TrimSpace(out), "\x00")
	if !ok {
		return "", 0, false
	}

	at, err := strconv.ParseInt(ct, 10, 64)
	if err != nil {
		return "", 0, false
	}

	return subject, at, true
}
