package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
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
