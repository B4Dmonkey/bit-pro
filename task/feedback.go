package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/spf13/pathologize"
)

const feedbackSubdir = "feedback"

var errNoDataRoot = errors.New("store has no data root")

var ErrOtherProject = errors.New("note belongs to another project")

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
	return filepath.Join(s.data, feedbackSubdir)
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
	if s.data == "" {
		return "", errNoDataRoot
	}

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

	for {
		path, err := s.writeNote(track, seq, body, head)
		if errors.Is(err, fs.ErrExist) {
			seq++

			continue
		}

		return path, err
	}
}

func (s *Store) writeNote(track string, seq int, body string, head Commit) (string, error) {
	id := noteID(track, seq)

	path := pathologize.Join(s.feedbackDir(), id+bodyExt)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fileMode)
	if err != nil {
		return "", fmt.Errorf("creating note %s: %w", id, err)
	}

	if _, err := f.WriteString(body); err != nil {
		_ = f.Close()

		return "", fmt.Errorf("writing note %s: %w", id, err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("closing note %s: %w", id, err)
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

func (s *Store) ListNotes(track string) ([]string, error) {
	if track != "" {
		var err error

		track, err = s.resolveTrack(track)
		if err != nil {
			return nil, err
		}
	}

	paths, err := filepath.Glob(filepath.Join(s.feedbackDir(), "*"+recordExt))
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", s.feedbackDir(), err)
	}

	var recs []noteRecord

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		var rec noteRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		if rec.Project == s.code && (track == "" || rec.Track == track) {
			recs = append(recs, rec)
		}
	}

	slices.SortFunc(recs, func(a, b noteRecord) int {
		if c := compareIDs(b.Track, a.Track); c != 0 {
			return c
		}

		return a.Seq - b.Seq
	})

	ids := make([]string, 0, len(recs))
	for _, rec := range recs {
		ids = append(ids, rec.ID)
	}

	return ids, nil
}

func (s *Store) ReadNote(id string) (string, error) {
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return "", fmt.Errorf("note ID %q looks like a path", id)
	}

	id = NormalizeID(id)
	dir := s.feedbackDir()

	data, err := os.ReadFile(pathologize.Join(dir, id+recordExt))
	if err != nil {
		return "", fmt.Errorf("reading note %s: %w", id, err)
	}

	var rec noteRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return "", fmt.Errorf("parsing note %s: %w", id, err)
	}

	if rec.Project != s.code {
		return "", fmt.Errorf("%s: %w", id, ErrOtherProject)
	}

	body, err := os.ReadFile(pathologize.Join(dir, rec.Content))
	if err != nil {
		return "", fmt.Errorf("reading note %s body: %w", id, err)
	}

	return string(body), nil
}
