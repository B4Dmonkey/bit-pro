package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pathologize"
	"gopkg.in/yaml.v3"
)

type Task struct {
	ID     string `yaml:"id"`
	Title  string `yaml:"title"`
	Status string `yaml:"status"`
	Body   string `yaml:"-"`
}

func taskPath(id string) string {
	return pathologize.Join(tasksDir, id+".md")
}

func loadTask(id string) (*Task, error) {
	data, err := os.ReadFile(taskPath(id))
	if err != nil {
		return nil, fmt.Errorf("loading task %s: %w", id, err)
	}
	return parseTask(data)
}

func parseTask(data []byte) (*Task, error) {
	s := string(data)
	if !strings.HasPrefix(s, "---\n") {
		return nil, errors.New("task file missing frontmatter delimiter")
	}
	rest := s[len("---\n"):]
	idx := strings.Index(rest, "\n---\n")
	if idx == -1 {
		return nil, errors.New("task file missing closing frontmatter delimiter")
	}
	var t Task
	if err := yaml.Unmarshal([]byte(rest[:idx+1]), &t); err != nil {
		return nil, fmt.Errorf("parsing task frontmatter: %w", err)
	}
	t.Body = rest[idx+len("\n---\n"):]
	return &t, nil
}

func (t *Task) save() error {
	header, err := yaml.Marshal(t)
	if err != nil {
		return fmt.Errorf("marshaling task %s: %w", t.ID, err)
	}
	content := "---\n" + string(header) + "---\n" + t.Body
	if err := os.WriteFile(taskPath(t.ID), []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing task %s: %w", t.ID, err)
	}
	return nil
}
