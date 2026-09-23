package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pathologize"
)

const researchSubdir = "research"

func (s *Store) researchDir(track string) string {
	return filepath.Join(s.root, researchSubdir, track)
}

func (s *Store) researchPath(track, topic string) string {
	return pathologize.Join(s.researchDir(track), pathologize.Clean(topic)+".md")
}

func (s *Store) WriteResearch(track, topic, body string) (string, error) {
	track = NormalizeID(track)

	if err := os.MkdirAll(s.researchDir(track), dirMode); err != nil {
		return "", fmt.Errorf("creating %s: %w", s.researchDir(track), err)
	}

	path := s.researchPath(track, topic)
	if err := os.WriteFile(path, []byte(body), fileMode); err != nil {
		return "", fmt.Errorf("writing %s: %w", path, err)
	}

	return path, nil
}
