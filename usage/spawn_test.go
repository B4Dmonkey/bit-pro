package usage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func launch(toolID, agentID string) string {
	return `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"` + toolID +
		`","content":"launched"}]},"toolUseResult":{"agentId":"` + agentID + `","isAsync":true,"status":"async_launched"}}`
}

func writeLines(t *testing.T, path string, lines ...string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("creating dir: %v", err)
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func TestFindSpawn(t *testing.T) {
	t.Run("nested agent follows the chain", func(t *testing.T) {
		dir := t.TempDir()
		main := filepath.Join(dir, "s1.jsonl")
		subs := filepath.Join(dir, "s1", "subagents")

		writeLines(t, main, assistant("msg_01", "BIT-1", 1), launch("toolu_a", "A"))
		writeLines(t, filepath.Join(subs, "agent-A.jsonl"), assistant("msg_02", "", 1), launch("toolu_b", "B"))
		writeLines(t, filepath.Join(subs, "agent-B.jsonl"), assistant("msg_03", "", 1))

		got, ok := FindSpawn(main, "B")
		if !ok {
			t.Fatal("FindSpawn(B) ok = false, want true")
		}

		if got.Track != trackOne || got.ParentAgentID != "A" {
			t.Errorf("FindSpawn(B) = %+v, want Track BIT-1 ParentAgentID A", got)
		}
	})

	t.Run("main spawn carries the bar in effect", func(t *testing.T) {
		dir := t.TempDir()
		main := filepath.Join(dir, "s1.jsonl")

		writeLines(t, main, assistant("msg_01", "BIT-1.1", 1), launch("toolu_a", "A"))

		got, ok := FindSpawn(main, "A")
		if !ok || got.Track != trackOne || got.Bar != "BIT-1.1" || got.ParentAgentID != "" {
			t.Errorf("FindSpawn(A) = %+v, %v, want BIT-1 BIT-1.1 main", got, ok)
		}
	})

	t.Run("unknown agent", func(t *testing.T) {
		dir := t.TempDir()
		main := filepath.Join(dir, "s1.jsonl")

		writeLines(t, main, assistant("msg_01", "BIT-1", 1), launch("toolu_a", "A"))

		if got, ok := FindSpawn(main, "Z"); ok {
			t.Errorf("FindSpawn(Z) = %+v, true, want ok false", got)
		}
	})
}
