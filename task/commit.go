package task

import (
	"context"
	"time"

	"github.com/B4Dmonkey/bit-pro/git"
)

func CommitAt(ctx context.Context, run git.Runner, dir string) Commit {
	h := git.ReadHead(ctx, run, dir)

	return Commit{SHA: h.SHA, Branch: h.Branch, At: time.Now()}
}
