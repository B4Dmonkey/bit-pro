package project

import (
	"context"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/db/orm"
	"github.com/B4Dmonkey/bit-pro/store"
	"github.com/B4Dmonkey/bit-pro/task"
)

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

type Stores struct {
	codes  []string
	stores map[string]*task.Store
}

func OpenAll(ctx context.Context) (*Stores, error) {
	sqlDB, err := db.Open()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	projects, err := Load(ctx, orm.New(sqlDB))
	if err != nil {
		return nil, err
	}

	s := &Stores{codes: make([]string, 0, len(projects)), stores: make(map[string]*task.Store, len(projects))}
	for _, p := range projects {
		if p.Removed {
			continue
		}

		ts, err := StoreFor(p)
		if err != nil {
			return nil, err
		}

		s.codes = append(s.codes, p.Code)
		s.stores[p.Code] = ts
	}

	slices.Sort(s.codes)

	return s, nil
}

func (s *Stores) List() ([]*task.Task, error) {
	var all []*task.Task

	for _, code := range s.codes {
		tasks, err := s.stores[code].List()
		if err != nil {
			return nil, fmt.Errorf("listing %s: %w", code, err)
		}

		all = append(all, tasks...)
	}

	return all, nil
}

func (s *Stores) For(id string) (*task.Store, error) {
	code, _, _ := strings.Cut(task.NormalizeID(id), "-")

	ts, ok := s.stores[code]
	if !ok {
		return nil, fmt.Errorf("task %s: %w", id, fs.ErrNotExist)
	}

	return ts, nil
}
