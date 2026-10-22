package usage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Entry struct {
	MessageID       string
	RequestID       string
	Timestamp       string
	Model           string
	Version         string
	Speed           string
	IsSidechain     bool
	Track           string
	InputTokens     int64
	CacheWrite5m    int64
	CacheWrite1h    int64
	CacheReadTokens int64
	OutputTokens    int64
}

type line struct {
	Type        string `json:"type"`
	RequestID   string `json:"requestId"`
	Timestamp   string `json:"timestamp"`
	Version     string `json:"version"`
	IsSidechain bool   `json:"isSidechain"`
	Message     struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Content []struct {
			Type  string `json:"type"`
			Name  string `json:"name"`
			Input struct {
				ID string `json:"id"`
			} `json:"input"`
		} `json:"content"`
		Usage struct {
			InputTokens     int64  `json:"input_tokens"`
			CacheReadTokens int64  `json:"cache_read_input_tokens"`
			OutputTokens    int64  `json:"output_tokens"`
			Speed           string `json:"speed"`
			CacheCreation   struct {
				Ephemeral5m int64 `json:"ephemeral_5m_input_tokens"`
				Ephemeral1h int64 `json:"ephemeral_1h_input_tokens"`
			} `json:"cache_creation"`
		} `json:"usage"`
	} `json:"message"`
}

const taskToolPrefix = "mcp__bit__task_"

func Parse(r io.Reader) ([]Entry, error) {
	var (
		entries    []Entry
		index      = map[string]int{}
		track      string
		firstTrack string
	)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 64*1024*1024)

	for scanner.Scan() {
		raw := bytes.TrimSpace(scanner.Bytes())
		if len(raw) == 0 {
			continue
		}

		var l line
		if err := json.Unmarshal(raw, &l); err != nil {
			return nil, fmt.Errorf("parsing transcript line: %w", err)
		}

		if l.Type != "assistant" {
			continue
		}

		if ref := trackRef(l); ref != "" {
			track = ref
			if firstTrack == "" {
				firstTrack = ref
			}
		}

		e := l.entry(track)

		key := e.MessageID + "\x00" + e.RequestID
		if i, ok := index[key]; ok {
			entries[i] = e
			continue
		}

		index[key] = len(entries)
		entries = append(entries, e)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading transcript: %w", err)
	}

	for i := range entries {
		if entries[i].Track == "" {
			entries[i].Track = firstTrack
		}
	}

	return entries, nil
}

func trackOf(id string) string {
	if i := strings.LastIndex(id, "."); i >= 0 {
		return id[:i]
	}

	return id
}

func trackRef(l line) string {
	for _, c := range l.Message.Content {
		if c.Type == "tool_use" && strings.HasPrefix(c.Name, taskToolPrefix) && c.Input.ID != "" {
			return trackOf(c.Input.ID)
		}
	}

	return ""
}

func (l line) entry(track string) Entry {
	u := l.Message.Usage

	return Entry{
		MessageID:       l.Message.ID,
		RequestID:       l.RequestID,
		Timestamp:       l.Timestamp,
		Model:           l.Message.Model,
		Version:         l.Version,
		Speed:           u.Speed,
		IsSidechain:     l.IsSidechain,
		Track:           track,
		InputTokens:     u.InputTokens,
		CacheWrite5m:    u.CacheCreation.Ephemeral5m,
		CacheWrite1h:    u.CacheCreation.Ephemeral1h,
		CacheReadTokens: u.CacheReadTokens,
		OutputTokens:    u.OutputTokens,
	}
}
