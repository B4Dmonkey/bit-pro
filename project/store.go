package project

import (
	"context"
	"fmt"
	"os"

	"github.com/B4Dmonkey/bit-pro/store"
	"github.com/B4Dmonkey/bit-pro/task"
)

func OpenCurrent(ctx context.Context) (*task.Store, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting the working directory: %w", err)
	}

	return OpenStore(ctx, wd)
}

func OpenStore(ctx context.Context, dir string) (*task.Store, error) {
	p, err := Find(ctx, dir)
	if err != nil {
		return nil, err
	}

	return StoreFor(p)
}

func StoreFor(p Project) (*task.Store, error) {
	root, err := store.ProjectDir(p.Code)
	if err != nil {
		return nil, err
	}

	data, err := store.Dir()
	if err != nil {
		return nil, err
	}

	return task.NewProject(root, p.Code).WithDataRoot(data), nil
}
