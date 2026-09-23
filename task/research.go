package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pathologize"
)

const researchSubdir = "research"

func (s *Store) researchDir(track string) string {
	return filepath.Join(s.root, researchSubdir, track)
}

func (s *Store) researchPath(track, topic string) string {
	return pathologize.Join(s.researchDir(track), strings.TrimLeft(pathologize.Clean(topic), ".")+".md")
}

func validateTopic(topic string) error {
	switch {
	case strings.Trim(topic, ". ") == "":
		return fmt.Errorf("research topic %q is empty", topic)
	case strings.Contains(topic, ".."):
		return fmt.Errorf("research topic %q contains \"..\"", topic)
	case strings.ContainsAny(topic, `/\`):
		return fmt.Errorf("research topic %q contains a path separator", topic)
	}

	return nil
}

func (s *Store) WriteResearch(track, topic, body string) (string, error) {
	track = NormalizeID(track)
	if !s.trackExists(track) {
		return "", fmt.Errorf("track %s does not exist", track)
	}

	if err := validateTopic(topic); err != nil {
		return "", err
	}

	if err := os.MkdirAll(s.researchDir(track), dirMode); err != nil {
		return "", fmt.Errorf("creating %s: %w", s.researchDir(track), err)
	}

	path := s.researchPath(track, topic)
	if err := os.WriteFile(path, []byte(body), fileMode); err != nil {
		return "", fmt.Errorf("writing %s: %w", path, err)
	}

	return path, nil
}

func (s *Store) ReadResearch(track, topic string) (string, error) {
	track = NormalizeID(track)
	if !s.trackExists(track) {
		return "", fmt.Errorf("track %s does not exist", track)
	}

	if err := validateTopic(topic); err != nil {
		return "", err
	}

	body, err := os.ReadFile(s.researchPath(track, topic))
	if err != nil {
		return "", fmt.Errorf("reading research topic %s for %s: %w", topic, track, err)
	}

	return string(body), nil
}
