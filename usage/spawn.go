package usage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Spawn struct {
	ParentAgentID string
	Track         string
	Bar           string
}

func (s Spawn) Target() string {
	if s.Bar != "" {
		return s.Bar
	}

	return s.Track
}

func FindSpawn(mainPath, agentID string) (Spawn, bool) {
	session := strings.TrimSuffix(filepath.Base(mainPath), filepath.Ext(mainPath))
	subDir := filepath.Join(filepath.Dir(mainPath), session, "subagents")

	candidates := []string{mainPath}
	if matches, err := filepath.Glob(filepath.Join(subDir, "agent-*.jsonl")); err == nil {
		candidates = append(candidates, matches...)
	}

	for _, path := range candidates {
		parent := ""
		if path != mainPath {
			parent = strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "agent-"), ".jsonl")
		}

		if parent == agentID {
			continue
		}

		seed := ""

		if parent != "" {
			if above, ok := FindSpawn(mainPath, parent); ok {
				seed = above.Target()
			}
		}

		if target, ok := spawnTarget(path, agentID, seed); ok {
			return Spawn{ParentAgentID: parent, Track: trackOf(target), Bar: barOf(target)}, true
		}
	}

	return Spawn{}, false
}

func spawnTarget(path, agentID, seed string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()

	p := newParser(seed)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 64*1024*1024)

	for scanner.Scan() {
		raw := bytes.TrimSpace(scanner.Bytes())
		if len(raw) == 0 {
			continue
		}

		var l line
		if json.Unmarshal(raw, &l) != nil {
			continue
		}

		p.add(l)

		if l.launchedAgent() == agentID {
			return p.target, true
		}
	}

	return "", false
}

func (l line) launchedAgent() string {
	var res struct {
		AgentID string `json:"agentId"`
	}

	if json.Unmarshal(l.ToolUseResult, &res) != nil {
		return ""
	}

	return res.AgentID
}
