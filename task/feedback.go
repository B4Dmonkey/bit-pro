package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/pathologize"
)

const feedbackSubdir = "feedback"

type noteRecord struct {
	Project   string    `json:"project"`
	ID        string    `json:"id"`
	Track     string    `json:"track"`
	Seq       int       `json:"seq"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Commits   []Commit  `json:"commits"`
	Content   string    `json:"content"`
}

func (s *Store) feedbackDir() string {
	return filepath.Join(s.root, feedbackSubdir)
}

func noteID(track string, seq int) string {
	return fmt.Sprintf("%s-%03d", track, seq)
}

func (s *Store) nextNoteSeq(track string) (int, error) {
	glob := track + "-*" + recordExt
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(track) + `-(\d+)\.json$`)

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

func (s *Store) resolveTrack(track string) (string, error) {
	if strings.Contains(track, "..") || strings.ContainsAny(track, `/\`) {
		return "", fmt.Errorf("track ID %q looks like a path", track)
	}

	track = NormalizeID(track)
	if !s.trackExists(track) {
		return "", fmt.Errorf("track %s does not exist", track)
	}

	return track, nil
}

func (s *Store) AddNote(track, body string, head Commit) (string, error) {
	track, err := s.resolveTrack(track)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(s.feedbackDir(), dirMode); err != nil {
		return "", fmt.Errorf("creating %s: %w", s.feedbackDir(), err)
	}

	seq, err := s.nextNoteSeq(track)
	if err != nil {
		return "", err
	}

	id := noteID(track, seq)

	path := pathologize.Join(s.feedbackDir(), id+bodyExt)
	if err := os.WriteFile(path, []byte(body), fileMode); err != nil {
		return "", fmt.Errorf("writing note for %s: %w", track, err)
	}

	commits := appendCommit(nil, head)
	if commits == nil {
		commits = []Commit{}
	}

	ts := s.now().UTC().Truncate(time.Second)
	rec := noteRecord{
		Project:   s.code,
		ID:        id,
		Track:     track,
		Seq:       seq,
		CreatedAt: ts,
		UpdatedAt: ts,
		Commits:   commits,
		Content:   filepath.Base(path),
	}

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling note %s: %w", id, err)
	}

	recPath := pathologize.Join(s.feedbackDir(), id+recordExt)
	if err := os.WriteFile(recPath, append(data, '\n'), fileMode); err != nil {
		return "", fmt.Errorf("writing %s: %w", recPath, err)
	}

	return path, nil
}
