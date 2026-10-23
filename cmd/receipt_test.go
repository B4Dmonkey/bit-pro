package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/db"
	"github.com/B4Dmonkey/bit-pro/project"
)

const (
	receiptSession = "0d6f0c2b-1a26-4dbc-b324-a6401ead76a8"

	usageBlock = `"usage":{"input_tokens":2,"cache_creation_input_tokens":19817,"cache_read_input_tokens":27335,` +
		`"output_tokens":180,"cache_creation":{"ephemeral_5m_input_tokens":19817,"ephemeral_1h_input_tokens":0},` +
		`"speed":"standard"}`

	toolUseLine = `{"type":"assistant","sessionId":"s1","requestId":"req_01","version":"2.1.0","isSidechain":false,` +
		`"timestamp":"2026-10-20T14:00:00Z","message":{"id":"msg_01","model":"claude-opus-5-5","role":"assistant",` +
		`"content":[{"type":"tool_use","id":"toolu_01","name":"mcp__bit__task_read","input":{"id":"BIT-1"}}],` +
		usageBlock + `}}`

	textOnlyLine = `{"type":"assistant","sessionId":"s1","requestId":"req_02","version":"2.1.0","isSidechain":false,` +
		`"timestamp":"2026-10-20T14:01:00Z","message":{"id":"msg_02","model":"claude-opus-5-5","role":"assistant",` +
		`"content":[{"type":"text","text":"hello"}],` + usageBlock + `}}`
)

func writeTranscript(t *testing.T, dir string, lines ...string) {
	t.Helper()

	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(dir, "t.jsonl"), []byte(body), 0o600); err != nil {
		t.Fatalf("writing transcript: %v", err)
	}
}

func countUsage(t *testing.T) int {
	t.Helper()

	sqlDB, err := db.Open()
	if err != nil {
		t.Fatalf("db.Open() returned error: %v", err)
	}
	defer sqlDB.Close()

	var n int
	if err := sqlDB.QueryRow("SELECT count(*) FROM usage").Scan(&n); err != nil {
		t.Fatalf("counting usage: %v", err)
	}

	return n
}

func receiptPayload(dir string) string {
	return receiptPayloadFor(receiptSession, dir, "t.jsonl")
}

func receiptPayloadFor(session, dir, transcript string) string {
	return fmt.Sprintf(
		`{"session_id":%q,"transcript_path":%q,"cwd":%q,"hook_event_name":"Stop"}`,
		session, filepath.Join(dir, transcript), dir,
	)
}

func TestReceiptCmd(t *testing.T) {
	t.Run("unregistered cwd is a silent no-op", func(t *testing.T) {
		t.Setenv("HOME", t.TempDir())
		t.Setenv("XDG_DATA_HOME", "")
		dir := t.TempDir()

		out, err := runWithStdin(t, receiptPayload(dir), "receipt")
		if err != nil {
			t.Fatalf("bp receipt returned error: %v", err)
		}

		if out != "" {
			t.Errorf("output = %q, want empty", out)
		}
	})

	t.Run("hidden from help", func(t *testing.T) {
		out := mustRun(t, "--help")

		if strings.Contains(out, "receipt") {
			t.Errorf("help output lists receipt:\n%s", out)
		}
	})

	t.Run("removed project is a silent no-op", func(t *testing.T) {
		dir := initProject(t, testPrefix)
		markRemoved(t, testPrefix)

		out, err := runWithStdin(t, receiptPayload(dir), "receipt")
		if err != nil {
			t.Fatalf("bp receipt returned error: %v", err)
		}

		if out != "" {
			t.Errorf("output = %q, want empty", out)
		}
	})

	t.Run("stores a row per response for the referenced track", func(t *testing.T) {
		dir := initProject(t, testPrefix)
		createTask(t, "Track", "## Why\nx\n")

		canonical, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
		}

		writeTranscript(t, canonical, toolUseLine)

		out, err := runWithStdin(t, receiptPayload(canonical), "receipt")
		if err != nil {
			t.Fatalf("bp receipt returned error: %v", err)
		}

		if out != "" {
			t.Errorf("output = %q, want empty", out)
		}

		sqlDB, err := db.Open()
		if err != nil {
			t.Fatalf("db.Open() returned error: %v", err)
		}
		defer sqlDB.Close()

		var (
			track, bar, session, messageID, requestID, model string
			in, w5, w1, read, outTok                         int64
			rows                                             int
		)

		if err := sqlDB.QueryRow("SELECT count(*) FROM usage").Scan(&rows); err != nil {
			t.Fatalf("counting usage: %v", err)
		}

		if rows != 1 {
			t.Fatalf("usage rows = %d, want 1", rows)
		}

		err = sqlDB.QueryRow(`SELECT track, bar, session_id, message_id, request_id, model,
			input_tokens, cache_write_5m_tokens, cache_write_1h_tokens, cache_read_tokens, output_tokens
			FROM usage`).Scan(&track, &bar, &session, &messageID, &requestID, &model, &in, &w5, &w1, &read, &outTok)
		if err != nil {
			t.Fatalf("reading usage row: %v", err)
		}

		if track != "BIT-1" || bar != "" || session != receiptSession || messageID != "msg_01" ||
			requestID != "req_01" || model != "claude-opus-5-5" {
			t.Errorf("row = %q %q %q %q %q %q", track, bar, session, messageID, requestID, model)
		}

		if in != 2 || w5 != 19817 || w1 != 0 || read != 27335 || outTok != 180 {
			t.Errorf("tokens = %d/%d/%d/%d/%d, want 2/19817/0/27335/180", in, w5, w1, read, outTok)
		}
	})

	t.Run("session with no track reference stores nothing", func(t *testing.T) {
		dir := initProject(t, testPrefix)

		canonical, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
		}

		writeTranscript(t, canonical, textOnlyLine)

		out, err := runWithStdin(t, receiptPayload(canonical), "receipt")
		if err != nil {
			t.Fatalf("bp receipt returned error: %v", err)
		}

		if out != "" {
			t.Errorf("output = %q, want empty", out)
		}

		if n := countUsage(t); n != 0 {
			t.Errorf("usage rows = %d, want 0", n)
		}
	})

	t.Run("sidechain copy does not replace the original", func(t *testing.T) {
		dir := initProject(t, testPrefix)
		createTask(t, "Track", "## Why\nx\n")

		canonical, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
		}

		copyLine := strings.Replace(
			strings.Replace(toolUseLine, `"isSidechain":false`, `"isSidechain":true`, 1),
			`"id":"BIT-1"`, `"id":"BIT-2"`, 1,
		)

		writeTranscript(t, canonical, toolUseLine)

		if err := os.WriteFile(filepath.Join(canonical, "s2.jsonl"), []byte(copyLine+"\n"), 0o600); err != nil {
			t.Fatalf("writing transcript: %v", err)
		}

		for _, payload := range []string{
			receiptPayloadFor("s1", canonical, "t.jsonl"),
			receiptPayloadFor("s2", canonical, "s2.jsonl"),
		} {
			if _, err := runWithStdin(t, payload, "receipt"); err != nil {
				t.Fatalf("bp receipt returned error: %v", err)
			}
		}

		sqlDB, err := db.Open()
		if err != nil {
			t.Fatalf("db.Open() returned error: %v", err)
		}
		defer sqlDB.Close()

		var (
			rows, sidechain int
			session, track  string
		)

		if err := sqlDB.QueryRow("SELECT count(*) FROM usage WHERE message_id = 'msg_01'").Scan(&rows); err != nil {
			t.Fatalf("counting usage: %v", err)
		}

		if rows != 1 {
			t.Fatalf("usage rows = %d, want 1", rows)
		}

		if err := sqlDB.QueryRow("SELECT session_id, track, is_sidechain FROM usage WHERE message_id = 'msg_01'").
			Scan(&session, &track, &sidechain); err != nil {
			t.Fatalf("reading usage row: %v", err)
		}

		if session != "s1" || track != "BIT-1" || sidechain != 0 {
			t.Errorf("row = %q %q %d, want s1 BIT-1 0", session, track, sidechain)
		}
	})

	t.Run("repeated stop on a grown transcript adds only new responses", func(t *testing.T) {
		dir := initProject(t, testPrefix)
		createTask(t, "Track", "## Why\nx\n")

		canonical, err := project.CanonicalPath(dir)
		if err != nil {
			t.Fatalf("CanonicalPath(%q) returned error: %v", dir, err)
		}

		writeTranscript(t, canonical, toolUseLine)

		if _, err := runWithStdin(t, receiptPayload(canonical), "receipt"); err != nil {
			t.Fatalf("bp receipt returned error: %v", err)
		}

		writeTranscript(t, canonical, toolUseLine, textOnlyLine)

		if _, err := runWithStdin(t, receiptPayload(canonical), "receipt"); err != nil {
			t.Fatalf("bp receipt returned error: %v", err)
		}

		if n := countUsage(t); n != 2 {
			t.Errorf("usage rows = %d, want 2", n)
		}
	})
}
