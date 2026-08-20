package task

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/pathologize"
)

const feedbackSubdir = "feedback"

func (s *Store) feedbackDir() string {
	return filepath.Join(s.root, feedbackSubdir)
}

func (s *Store) notePath(track string, seq int) string {
	return pathologize.Join(s.feedbackDir(), fmt.Sprintf("%s-%03d.md", track, seq))
}

func (s *Store) nextNoteSeq(track string) (int, error) {
	glob := track + "-*.md"
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(track) + `-(\d+)\.md$`)

	highest, err := highestSuffix(s.feedbackDir(), glob, re)
	if err != nil {
		return 0, fmt.Errorf("scanning %s for existing notes: %w", s.feedbackDir(), err)
	}

	return highest + 1, nil
}

func (s *Store) trackExists(track string) bool {
	for _, path := range []string{s.Path(track), s.completedPath(track), s.archivePath(track)} {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

func (s *Store) AddNote(track, body string) (string, error) {
	track = NormalizeID(track)
	if !s.trackExists(track) {
		return "", fmt.Errorf("track %s does not exist", track)
	}

	if err := os.MkdirAll(s.feedbackDir(), dirMode); err != nil {
		return "", fmt.Errorf("creating %s: %w", s.feedbackDir(), err)
	}

	seq, err := s.nextNoteSeq(track)
	if err != nil {
		return "", err
	}

	path := s.notePath(track, seq)
	if err := os.WriteFile(path, []byte(body), fileMode); err != nil {
		return "", fmt.Errorf("writing note for %s: %w", track, err)
	}

	return path, nil
}
