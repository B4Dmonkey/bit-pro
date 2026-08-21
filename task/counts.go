package task

type Counts struct {
	Backlog, Todo, Done, Completed int
}

func (s *Store) Counts() (Counts, error) {
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

	return c, nil
}
