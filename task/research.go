package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pathologize"
)

const researchSubdir = "research"

type researchRecord struct {
	Project   string    `json:"project"`
	Track     string    `json:"track"`
	Topic     string    `json:"topic"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Commits   []Commit  `json:"commits"`
	Content   string    `json:"content"`
}

func (s *Store) researchDir(track string) string {
	return filepath.Join(s.root, researchSubdir, track)
}

func topicStem(topic string) string {
	return strings.TrimLeft(pathologize.Clean(topic), ".")
}

func (s *Store) researchPath(track, topic string) string {
	return pathologize.Join(s.researchDir(track), topicStem(topic)+bodyExt)
}

func (s *Store) researchRecordPath(track, topic string) string {
	return pathologize.Join(s.researchDir(track), topicStem(topic)+recordExt)
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

func readResearchRecord(path string) (researchRecord, error) {
	var rec researchRecord

	data, err := os.ReadFile(path)
	if err != nil {
		return rec, err
	}

	if err := json.Unmarshal(data, &rec); err != nil {
		return rec, fmt.Errorf("parsing %s: %w", path, err)
	}

	return rec, nil
}

func (s *Store) WriteResearch(track, topic, body string, head Commit) (string, error) {
	track, err := s.resolveTrack(track)
	if err != nil {
		return "", err
	}

	if err := validateTopic(topic); err != nil {
		return "", err
	}

	if err := os.MkdirAll(s.researchDir(track), dirMode); err != nil {
		return "", fmt.Errorf("creating %s: %w", s.researchDir(track), err)
	}

	recPath := s.researchRecordPath(track, topic)

	rec, err := readResearchRecord(recPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("reading %s: %w", recPath, err)
	}

	path := s.researchPath(track, topic)
	if err := os.WriteFile(path, []byte(body), fileMode); err != nil {
		return "", fmt.Errorf("writing %s: %w", path, err)
	}

	ts := s.now().UTC().Truncate(time.Second)
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = ts
	}

	rec.Commits = appendCommit(rec.Commits, head)
	if rec.Commits == nil {
		rec.Commits = []Commit{}
	}

	rec.Project = s.code
	rec.Track = track
	rec.Topic = topicStem(topic)
	rec.UpdatedAt = ts
	rec.Content = filepath.Base(path)

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling research topic %s: %w", topic, err)
	}

	if err := os.WriteFile(recPath, append(data, '\n'), fileMode); err != nil {
		return "", fmt.Errorf("writing %s: %w", recPath, err)
	}

	return path, nil
}

func (s *Store) ReadResearch(track, topic string) (string, error) {
	track, err := s.resolveTrack(track)
	if err != nil {
		return "", err
	}

	if err := validateTopic(topic); err != nil {
		return "", err
	}

	rec, err := readResearchRecord(s.researchRecordPath(track, topic))
	if err != nil {
		return "", fmt.Errorf("reading research topic %s for %s: %w", topic, track, err)
	}

	body, err := os.ReadFile(pathologize.Join(s.researchDir(track), rec.Content))
	if err != nil {
		return "", fmt.Errorf("reading research topic %s for %s: %w", topic, track, err)
	}

	return string(body), nil
}

func (s *Store) ResearchTopics(track string) ([]string, error) {
	track, err := s.resolveTrack(track)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.researchDir(track))
	if errors.Is(err, fs.ErrNotExist) {
		return []string{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("listing research for %s: %w", track, err)
	}

	var topics []string

	for _, e := range entries {
		if e.Type().IsRegular() && strings.HasSuffix(e.Name(), recordExt) {
			topics = append(topics, strings.TrimSuffix(e.Name(), recordExt))
		}
	}

	return topics, nil
}
