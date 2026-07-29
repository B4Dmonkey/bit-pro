package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pathologize"
)

const feedbackSubdir = "feedback"

func (s *Store) feedbackDir() string {
	return filepath.Join(s.root, feedbackSubdir)
}

func (s *Store) notePath(track string, seq int) string {
	return pathologize.Join(s.feedbackDir(), fmt.Sprintf("%s-%03d.md", track, seq))
}

func (s *Store) AddNote(track, body string) (string, error) {
	if err := os.MkdirAll(s.feedbackDir(), dirMode); err != nil {
		return "", fmt.Errorf("creating %s: %w", s.feedbackDir(), err)
	}
	path := s.notePath(track, 1)
	if err := os.WriteFile(path, []byte(body), fileMode); err != nil {
		return "", fmt.Errorf("writing note for %s: %w", track, err)
	}
	return path, nil
}
