package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/pathologize"
)

const retroSubdir = "retro"

type retroRecord struct {
	Project   string    `json:"project"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Commits   []Commit  `json:"commits"`
	Content   string    `json:"content"`
}

func (s *Store) retroDir() string {
	return filepath.Join(s.data, retroSubdir)
}

func retroName(code, name string) string {
	if strings.HasPrefix(name, code+"-") {
		return name
	}

	return code + "-" + name
}

func validateRetroName(name string) error {
	switch {
	case strings.Trim(name, ". ") == "":
		return fmt.Errorf("retro name %q is empty", name)
	case strings.Contains(name, ".."):
		return fmt.Errorf("retro name %q contains \"..\"", name)
	case strings.ContainsAny(name, `/\`):
		return fmt.Errorf("retro name %q contains a path separator", name)
	}

	return nil
}

func readRetroRecord(path string) (retroRecord, error) {
	var rec retroRecord

	data, err := os.ReadFile(path)
	if err != nil {
		return rec, err
	}

	if err := json.Unmarshal(data, &rec); err != nil {
		return rec, fmt.Errorf("parsing %s: %w", path, err)
	}

	return rec, nil
}

func (s *Store) WriteRetro(name, body string, head Commit) (string, error) {
	if s.data == "" {
		return "", errNoDataRoot
	}

	if err := validateRetroName(name); err != nil {
		return "", err
	}

	name = retroName(s.code, name)

	if err := os.MkdirAll(s.retroDir(), dirMode); err != nil {
		return "", fmt.Errorf("creating %s: %w", s.retroDir(), err)
	}

	recPath := pathologize.Join(s.retroDir(), name+recordExt)

	rec, err := readRetroRecord(recPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("reading %s: %w", recPath, err)
	}

	path := pathologize.Join(s.retroDir(), name+bodyExt)
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
	rec.Name = name
	rec.UpdatedAt = ts
	rec.Content = filepath.Base(path)

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling retro %s: %w", name, err)
	}

	if err := os.WriteFile(recPath, append(data, '\n'), fileMode); err != nil {
		return "", fmt.Errorf("writing %s: %w", recPath, err)
	}

	return name, nil
}

type Proposal struct {
	Name    string `json:"name"`
	Project string `json:"project"`
}

func (s *Store) ListRetro() ([]Proposal, error) {
	paths, err := filepath.Glob(filepath.Join(s.retroDir(), "*"+recordExt))
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", s.retroDir(), err)
	}

	proposals := make([]Proposal, 0, len(paths))

	for _, path := range paths {
		rec, err := readRetroRecord(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}

		proposals = append(proposals, Proposal{Name: rec.Name, Project: rec.Project})
	}

	slices.SortFunc(proposals, func(a, b Proposal) int { return strings.Compare(a.Name, b.Name) })

	return proposals, nil
}
