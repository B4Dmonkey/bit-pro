package project

import (
	"context"

	"github.com/B4Dmonkey/bit-pro/store"
	"github.com/B4Dmonkey/bit-pro/task"
)

func OpenStore(ctx context.Context, dir string) (*task.Store, error) {
	p, err := Find(ctx, dir)
	if err != nil {
		return nil, err
	}

	root, err := store.ProjectDir(p.Code)
	if err != nil {
		return nil, err
	}

	return task.NewProject(root, p.Code), nil
}
