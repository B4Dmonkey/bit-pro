package usage

import (
	"strconv"
	"strings"
	"testing"
)

func assistant(msgID, toolID string, out int) string {
	content := `[{"type":"text","text":"x"}]`
	if toolID != "" {
		content = `[{"type":"tool_use","id":"toolu_1","name":"mcp__bit__task_read","input":{"id":"` + toolID + `"}}]`
	}

	return `{"type":"assistant","requestId":"req_` + msgID + `","timestamp":"2026-10-20T14:00:00Z",` +
		`"version":"2.1.0","isSidechain":false,` +
		`"message":{"id":"` + msgID + `","model":"claude-opus-5-5","content":` + content +
		`,"usage":{"input_tokens":2,"cache_read_input_tokens":3,"output_tokens":` + strconv.Itoa(out) +
		`,"cache_creation":{"ephemeral_5m_input_tokens":4,"ephemeral_1h_input_tokens":5},"speed":"standard"}}}`
}

func TestParse(t *testing.T) {
	t.Run("fills an entry from an assistant line", func(t *testing.T) {
		entries, err := Parse(strings.NewReader(assistant("m1", "BIT-1", 7)))
		if err != nil {
			t.Fatalf("Parse() returned error: %v", err)
		}

		want := Entry{
			MessageID: "m1", RequestID: "req_m1", Timestamp: "2026-10-20T14:00:00Z",
			Model: "claude-opus-5-5", Version: "2.1.0", Speed: "standard", Track: "BIT-1",
			InputTokens: 2, CacheWrite5m: 4, CacheWrite1h: 5, CacheReadTokens: 3, OutputTokens: 7,
		}

		if len(entries) != 1 || entries[0] != want {
			t.Errorf("Parse() = %+v, want [%+v]", entries, want)
		}
	})

	t.Run("a bar id attributes to its track", func(t *testing.T) {
		entries, _ := Parse(strings.NewReader(assistant("m1", "BIT-58.20", 1)))

		if len(entries) != 1 || entries[0].Track != "BIT-58" {
			t.Errorf("Parse() = %+v, want track BIT-58", entries)
		}
	})

	t.Run("responses before the first reference go to the first track", func(t *testing.T) {
		in := strings.Join([]string{
			assistant("m1", "", 1), assistant("m2", "BIT-1", 2), assistant("m3", "BIT-2", 3), assistant("m4", "", 4),
		}, "\n")

		entries, err := Parse(strings.NewReader(in))
		if err != nil {
			t.Fatalf("Parse() returned error: %v", err)
		}

		var got []string
		for _, e := range entries {
			got = append(got, e.Track)
		}

		if want := "BIT-1,BIT-1,BIT-2,BIT-2"; strings.Join(got, ",") != want {
			t.Errorf("tracks = %v, want %s", got, want)
		}
	})

	t.Run("the last line per message id wins", func(t *testing.T) {
		in := assistant("m1", "BIT-1", 1) + "\n" + assistant("m1", "", 9)

		entries, _ := Parse(strings.NewReader(in))

		if len(entries) != 1 || entries[0].OutputTokens != 9 {
			t.Errorf("Parse() = %+v, want one entry with 9 output tokens", entries)
		}
	})

	t.Run("duplicate lines keep the last", func(t *testing.T) {
		in := strings.Join([]string{
			assistant("msg_01", "BIT-1", 12), assistant("msg_01", "BIT-1", 180), assistant("msg_01", "BIT-1", 180),
		}, "\n")

		entries, err := Parse(strings.NewReader(in))
		if err != nil {
			t.Fatalf("Parse() returned error: %v", err)
		}

		if len(entries) != 1 || entries[0].OutputTokens != 180 {
			t.Errorf("Parse() = %+v, want one entry with 180 output tokens", entries)
		}
	})

	t.Run("the same message id under another request id is a separate entry", func(t *testing.T) {
		other := strings.Replace(assistant("m1", "", 9), "req_m1", "req_other", 1)
		in := assistant("m1", "BIT-1", 1) + "\n" + other

		entries, _ := Parse(strings.NewReader(in))

		if len(entries) != 2 {
			t.Errorf("Parse() = %+v, want 2 entries", entries)
		}
	})

	t.Run("skips non-assistant lines", func(t *testing.T) {
		entries, err := Parse(strings.NewReader(`{"type":"user"}` + "\n"))
		if err != nil || len(entries) != 0 {
			t.Errorf("Parse() = %+v, %v, want none", entries, err)
		}
	})
}

func TestParseSession(t *testing.T) {
	t.Run("agent setting names the invoked agent", func(t *testing.T) {
		in := `{"type":"agent-setting","agentSetting":"bit:ruler","sessionId":"s1"}` + "\n" + assistant("m1", "", 1)

		got, err := ParseSession(strings.NewReader(in))
		if err != nil {
			t.Fatalf("ParseSession() returned error: %v", err)
		}

		if got != "bit:ruler" {
			t.Errorf("ParseSession() = %q, want bit:ruler", got)
		}
	})

	t.Run("no agent setting is main", func(t *testing.T) {
		got, err := ParseSession(strings.NewReader(assistant("m1", "", 1)))
		if err != nil {
			t.Fatalf("ParseSession() returned error: %v", err)
		}

		if got != "main" {
			t.Errorf("ParseSession() = %q, want main", got)
		}
	})
}
