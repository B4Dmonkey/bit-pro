package usage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Entry struct {
	MessageID       string
	RequestID       string
	Entry           int
	Timestamp       string
	Model           string
	Version         string
	Speed           string
	IsSidechain     bool
	Track           string
	Bar             string
	InputTokens     int64
	CacheWrite5m    int64
	CacheWrite1h    int64
	CacheReadTokens int64
	OutputTokens    int64
}

type line struct {
	Type         string `json:"type"`
	AgentSetting string `json:"agentSetting"`
	RequestID    string `json:"requestId"`
	Timestamp    string `json:"timestamp"`
	Version      string `json:"version"`
	IsSidechain  bool   `json:"isSidechain"`

	ToolUseResult json.RawMessage `json:"toolUseResult"`
	Message       struct {
		ID      string  `json:"id"`
		Model   string  `json:"model"`
		Content content `json:"content"`
		Usage   struct {
			InputTokens     int64  `json:"input_tokens"`
			CacheReadTokens int64  `json:"cache_read_input_tokens"`
			OutputTokens    int64  `json:"output_tokens"`
			Speed           string `json:"speed"`
			CacheCreation   struct {
				Ephemeral5m int64 `json:"ephemeral_5m_input_tokens"`
				Ephemeral1h int64 `json:"ephemeral_1h_input_tokens"`
			} `json:"cache_creation"`
			Iterations []iteration `json:"iterations"`
		} `json:"usage"`
	} `json:"message"`
}

type contentItem struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	ID        string `json:"id"`
	ToolUseID string `json:"tool_use_id"`
	Input     struct {
		ID string `json:"id"`
	} `json:"input"`
}

type content []contentItem

func (c *content) UnmarshalJSON(b []byte) error {
	var items []contentItem
	if err := json.Unmarshal(b, &items); err != nil {
		*c = nil

		return nil //nolint:nilerr // string content carries no tool items
	}

	*c = items

	return nil
}

type iteration struct {
	Type            string `json:"type"`
	Model           string `json:"model"`
	InputTokens     int64  `json:"input_tokens"`
	CacheReadTokens int64  `json:"cache_read_input_tokens"`
	OutputTokens    int64  `json:"output_tokens"`
	CacheCreation   struct {
		Ephemeral5m int64 `json:"ephemeral_5m_input_tokens"`
		Ephemeral1h int64 `json:"ephemeral_1h_input_tokens"`
	} `json:"cache_creation"`
}

const advisorIteration = "advisor_message"

const (
	taskToolPrefix = "mcp__bit__task_"
	taskCreateTool = "mcp__bit__task_create"
	mainAgent      = "main"
)

func ParseSession(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 64*1024*1024)

	for scanner.Scan() {
		raw := bytes.TrimSpace(scanner.Bytes())
		if len(raw) == 0 {
			continue
		}

		var l line
		if err := json.Unmarshal(raw, &l); err != nil {
			return "", fmt.Errorf("parsing transcript line: %w", err)
		}

		if l.Type == "agent-setting" && l.AgentSetting != "" {
			return l.AgentSetting, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading transcript: %w", err)
	}

	return mainAgent, nil
}

type parser struct {
	entries    []Entry
	index      map[string]int
	creates    map[string]bool
	target     string
	firstTrack string
}

func Parse(r io.Reader) ([]Entry, error) {
	p := &parser{index: map[string]int{}, creates: map[string]bool{}}

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

		p.add(l)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading transcript: %w", err)
	}

	for i := range p.entries {
		if p.entries[i].Track == "" {
			p.entries[i].Track = p.firstTrack
		}
	}

	return p.entries, nil
}

func (p *parser) add(l line) {
	switch l.Type {
	case "user":
		if id := l.createdID(p.creates); id != "" {
			p.target = id
		}
	case "assistant":
		p.addAssistant(l)
	}
}

func (p *parser) addAssistant(l line) {
	for _, c := range l.Message.Content {
		if c.Type == "tool_use" && c.Name == taskCreateTool {
			p.creates[c.ID] = true
		}
	}

	if ref := taskRef(l); ref != "" {
		p.target = ref
		if p.firstTrack == "" {
			p.firstTrack = trackOf(ref)
		}
	}

	for _, e := range l.entries(trackOf(p.target), barOf(p.target)) {
		key := e.MessageID + "\x00" + e.RequestID + "\x00" + strconv.Itoa(e.Entry)
		if i, ok := p.index[key]; ok {
			p.entries[i] = e
			continue
		}

		p.index[key] = len(p.entries)
		p.entries = append(p.entries, e)
	}
}

func trackOf(id string) string {
	if i := strings.LastIndex(id, "."); i >= 0 {
		return id[:i]
	}

	return id
}

func barOf(id string) string {
	if strings.Contains(id, ".") {
		return id
	}

	return ""
}

func (l line) createdID(creates map[string]bool) string {
	for _, c := range l.Message.Content {
		if c.Type != "tool_result" || !creates[c.ToolUseID] {
			continue
		}

		var res struct {
			ID string `json:"id"`
		}

		if json.Unmarshal(l.ToolUseResult, &res) == nil {
			return res.ID
		}
	}

	return ""
}

func taskRef(l line) string {
	for _, c := range l.Message.Content {
		if c.Type == "tool_use" && strings.HasPrefix(c.Name, taskToolPrefix) && c.Input.ID != "" {
			return c.Input.ID
		}
	}

	return ""
}

func (l line) entries(track, bar string) []Entry {
	entries := []Entry{l.entry(track, bar)}

	for i, it := range l.Message.Usage.Iterations {
		if it.Type != advisorIteration {
			continue
		}

		e := l.entry(track, bar)
		e.Entry = i + 1
		e.Model = it.Model
		e.InputTokens = it.InputTokens
		e.CacheWrite5m = it.CacheCreation.Ephemeral5m
		e.CacheWrite1h = it.CacheCreation.Ephemeral1h
		e.CacheReadTokens = it.CacheReadTokens
		e.OutputTokens = it.OutputTokens
		entries = append(entries, e)
	}

	return entries
}

func (l line) entry(track, bar string) Entry {
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
		Bar:             bar,
		InputTokens:     u.InputTokens,
		CacheWrite5m:    u.CacheCreation.Ephemeral5m,
		CacheWrite1h:    u.CacheCreation.Ephemeral1h,
		CacheReadTokens: u.CacheReadTokens,
		OutputTokens:    u.OutputTokens,
	}
}
