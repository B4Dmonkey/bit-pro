package task

import (
	"fmt"
	"os"
)

type Counts struct {
	Backlog, Todo, Done, Completed int
}

func (s *Store) Counts() (Counts, error) {
	if _, err := os.Stat(s.root); err != nil {
		return Counts{}, fmt.Errorf("reading %s: %w", s.root, err)
	}

	tasks, err := s.List()
	if err != nil {
		return Counts{}, err
	}

	var c Counts

	for _, t := range tasks {
		if _, ok := barParent(t.ID); ok {
			continue
		}

		if !t.Approved {
			c.Backlog++
		} else if t.Status == StatusDone {
			c.Done++
		} else {
			c.Todo++
		}
	}

	completed, err := s.listCompleted()
	if err != nil {
		return Counts{}, err
	}

	for _, t := range completed {
		if _, ok := barParent(t.ID); ok {
			continue
		}

		c.Completed++
	}

	return c, nil
}
