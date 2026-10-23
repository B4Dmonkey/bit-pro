package usage

import (
	"strconv"
	"strings"
	"testing"
)

const (
	opusModel  = "claude-opus-5-5"
	trackOne   = "BIT-1"
	rulerAgent = "bit:ruler"
	barOne     = "BIT-1.1"
)

func assistant(msgID, toolID string, out int) string {
	content := `[{"type":"text","text":"x"}]`
	if toolID != "" {
		content = `[{"type":"tool_use","id":"toolu_1","name":"mcp__bit__task_read","input":{"id":"` + toolID + `"}}]`
	}

	return `{"type":"assistant","requestId":"req_` + msgID + `","timestamp":"2026-10-20T14:00:00Z",` +
		`"version":"2.1.0","isSidechain":false,` +
		`"message":{"id":"` + msgID + `","model":"` + opusModel + `","content":` + content +
		`,"usage":{"input_tokens":2,"cache_read_input_tokens":3,"output_tokens":` + strconv.Itoa(out) +
		`,"cache_creation":{"ephemeral_5m_input_tokens":4,"ephemeral_1h_input_tokens":5},"speed":"standard"}}}`
}

func TestParse(t *testing.T) {
	t.Run("skill spans end at the next skill or human prompt", func(t *testing.T) {
		skillUse := `{"type":"assistant","requestId":"req_msg_03","timestamp":"2026-10-20T14:00:00Z",` +
			`"message":{"id":"msg_03","model":"` + opusModel + `","content":[{"type":"tool_use","id":"toolu_07",` +
			`"name":"Skill","input":{"skill":"bit:scope"}}],"usage":{"output_tokens":1}}}`

		in := strings.Join([]string{
			assistant("msg_01", "BIT-1", 1),
			`{"type":"user","uuid":"u1","message":{"content":"<command-message>bit:analyze</command-message>` +
				`<command-name>/bit:analyze</command-name>"}}`,
			assistant("msg_02", "", 1),
			skillUse,
			assistant("msg_04", "", 1),
			`{"type":"user","uuid":"u2","message":{"content":"looks good"}}`,
			assistant("msg_05", "", 1),
		}, "\n")

		entries, err := Parse(strings.NewReader(in), "")
		if err != nil {
			t.Fatalf("Parse() returned error: %v", err)
		}

		var skills, invs []string
		for _, e := range entries {
			skills = append(skills, e.Skill)
			invs = append(invs, e.SkillInvocation)
		}

		if want := ",bit:analyze,bit:analyze,bit:scope,"; strings.Join(skills, ",") != want {
			t.Errorf("skills = %q, want %q", strings.Join(skills, ","), want)
		}

		if want := ",u1,u1,toolu_07,"; strings.Join(invs, ",") != want {
			t.Errorf("invocations = %q, want %q", strings.Join(invs, ","), want)
		}
	})

	t.Run("fills an entry from an assistant line", func(t *testing.T) {
		entries, err := Parse(strings.NewReader(assistant("m1", "BIT-1", 7)), "")
		if err != nil {
			t.Fatalf("Parse() returned error: %v", err)
		}

		want := Entry{
			MessageID: "m1", RequestID: "req_m1", Timestamp: "2026-10-20T14:00:00Z",
			Model: opusModel, Version: "2.1.0", Speed: "standard", Track: trackOne,
			InputTokens: 2, CacheWrite5m: 4, CacheWrite1h: 5, CacheReadTokens: 3, OutputTokens: 7,
		}

		if len(entries) != 1 || entries[0] != want {
			t.Errorf("Parse() = %+v, want [%+v]", entries, want)
		}
	})

	t.Run("a bar id attributes to its track", func(t *testing.T) {
		entries, _ := Parse(strings.NewReader(assistant("m1", "BIT-58.20", 1)), "")

		if len(entries) != 1 || entries[0].Track != "BIT-58" {
			t.Errorf("Parse() = %+v, want track BIT-58", entries)
		}
	})

	t.Run("responses before the first reference go to the first track", func(t *testing.T) {
		in := strings.Join([]string{
			assistant("m1", "", 1), assistant("m2", "BIT-1", 2), assistant("m3", "BIT-2", 3), assistant("m4", "", 4),
		}, "\n")

		entries, err := Parse(strings.NewReader(in), "")
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

		entries, _ := Parse(strings.NewReader(in), "")

		if len(entries) != 1 || entries[0].OutputTokens != 9 {
			t.Errorf("Parse() = %+v, want one entry with 9 output tokens", entries)
		}
	})

	t.Run("duplicate lines keep the last", func(t *testing.T) {
		in := strings.Join([]string{
			assistant("msg_01", "BIT-1", 12), assistant("msg_01", "BIT-1", 180), assistant("msg_01", "BIT-1", 180),
		}, "\n")

		entries, err := Parse(strings.NewReader(in), "")
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

		entries, _ := Parse(strings.NewReader(in), "")

		if len(entries) != 2 {
			t.Errorf("Parse() = %+v, want 2 entries", entries)
		}
	})

	t.Run("advisor iteration becomes its own entry", func(t *testing.T) {
		in := `{"type":"assistant","requestId":"req_01","timestamp":"2026-10-20T14:00:00Z","version":"2.1.0",` +
			`"isSidechain":false,"message":{"id":"msg_01","model":"` + opusModel + `","content":[],` +
			`"usage":{"input_tokens":2,"output_tokens":180,"iterations":[` +
			`{"type":"message","model":"` + opusModel + `","input_tokens":2,"output_tokens":180},` +
			`{"type":"advisor_message","model":"claude-fable-5","input_tokens":1000,"output_tokens":50}]}}}`

		entries, err := Parse(strings.NewReader(in), "")
		if err != nil {
			t.Fatalf("Parse() returned error: %v", err)
		}

		if len(entries) != 2 {
			t.Fatalf("Parse() = %+v, want 2 entries", entries)
		}

		if e := entries[0]; e.Entry != 0 || e.Model != opusModel || e.InputTokens != 2 || e.OutputTokens != 180 {
			t.Errorf("entries[0] = %+v, want opus in 2 out 180", e)
		}

		if e := entries[1]; e.Entry != 2 || e.Model != "claude-fable-5" || e.InputTokens != 1000 ||
			e.OutputTokens != 50 || e.MessageID != "msg_01" || e.RequestID != "req_01" {
			t.Errorf("entries[1] = %+v, want advisor in 1000 out 50", e)
		}
	})

	t.Run("bar targets from read and create", func(t *testing.T) {
		create := strings.Replace(assistant("msg_03", "", 20), `[{"type":"text","text":"x"}]`,
			`[{"type":"tool_use","id":"toolu_05","name":"mcp__bit__task_create","input":{"parent":"BIT-1","title":"t"}}]`, 1)
		result := `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_05","content":"x"}]},` +
			`"toolUseResult":{"id":"BIT-1.2"}}`
		in := strings.Join([]string{
			assistant("msg_01", "BIT-1", 5), assistant("msg_02", barOne, 10), create, result, assistant("msg_04", "", 100),
		}, "\n")

		entries, err := Parse(strings.NewReader(in), "")
		if err != nil {
			t.Fatalf("Parse() returned error: %v", err)
		}

		var bars, tracks []string
		for _, e := range entries {
			bars = append(bars, e.Bar)
			tracks = append(tracks, e.Track)
		}

		if want := ",BIT-1.1,BIT-1.1,BIT-1.2"; strings.Join(bars, ",") != want {
			t.Errorf("bars = %q, want %q", strings.Join(bars, ","), want)
		}

		if want := "BIT-1,BIT-1,BIT-1,BIT-1"; strings.Join(tracks, ",") != want {
			t.Errorf("tracks = %q, want %q", strings.Join(tracks, ","), want)
		}
	})

	t.Run("skips non-assistant lines", func(t *testing.T) {
		entries, err := Parse(strings.NewReader(`{"type":"user"}`+"\n"), "")
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

		if got != rulerAgent {
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
