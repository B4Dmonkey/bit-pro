package task

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const delim = "---"

type Task struct {
	ID         string   `yaml:"id"`
	Title      string   `yaml:"title"`
	Status     string   `yaml:"status"`
	Phase      int      `yaml:"phase,omitempty"`
	PhaseLabel string   `yaml:"phase_label,omitempty"`
	Order      []string `yaml:"order,omitempty"`
	Body       string   `yaml:"-"`
}

func Parse(data []byte) (*Task, error) {
	s := string(data)
	if !strings.HasPrefix(s, delim+"\n") {
		return nil, errors.New("task file missing frontmatter delimiter")
	}
	rest := s[len(delim)+1:]

	idx := strings.Index(rest, "\n"+delim+"\n")
	if idx == -1 {
		return nil, errors.New("task file missing closing frontmatter delimiter")
	}

	var t Task
	if err := yaml.Unmarshal([]byte(rest[:idx+1]), &t); err != nil {
		return nil, fmt.Errorf("parsing task frontmatter: %w", err)
	}
	t.ID = normalizeID(t.ID)
	t.Body = rest[idx+len("\n"+delim+"\n"):]
	return &t, nil
}

func (t *Task) Bytes() ([]byte, error) {
	header, err := yaml.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("marshaling task %s: %w", t.ID, err)
	}
	return []byte(delim + "\n" + string(header) + delim + "\n" + t.Body), nil
}
