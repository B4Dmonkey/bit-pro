package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/B4Dmonkey/bit-pro/git"
	"github.com/B4Dmonkey/bit-pro/task"
)

func sessionDir(root string) (string, error) {
	if root != "" {
		return root, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting the working directory: %w", err)
	}

	return wd, nil
}

func sessionHead(ctx context.Context, run git.Runner, dir string) task.Commit {
	h := git.ReadHead(ctx, run, dir)

	return task.Commit{SHA: h.SHA, Branch: h.Branch, At: time.Now()}
}
