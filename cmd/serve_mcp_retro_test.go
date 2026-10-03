package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const testRetroBody = "## Proposal 1\n\n**Pattern:** ...\n"

func TestRetroWriteHandler(t *testing.T) {
	t.Run("stores a proposal under the code-prefixed name", func(t *testing.T) {
		dir := t.TempDir()
		seedCodedProject(t, dir, testOwnCode, testOwnTrack)

		got := callTool(t, mcpSession(t, dir), retroWriteTool, map[string]any{
			"name":      "album-proposals",
			testBodyKey: testRetroBody,
		})

		const want = "BIT-album-proposals"
		if got["name"] != want {
			t.Errorf("name = %v, want %q", got["name"], want)
		}

		retroDir := filepath.Join(dataDir(t), "retro")

		body, err := os.ReadFile(filepath.Join(retroDir, want+".md"))
		if err != nil {
			t.Fatal(err)
		}

		if string(body) != testRetroBody {
			t.Errorf("body = %q, want %q", body, testRetroBody)
		}

		raw, err := os.ReadFile(filepath.Join(retroDir, want+".json"))
		if err != nil {
			t.Fatal(err)
		}

		var rec map[string]any
		if err := json.Unmarshal(raw, &rec); err != nil {
			t.Fatal(err)
		}

		for key, wantVal := range map[string]any{
			"project": testOwnCode,
			"name":    want,
			"commits": []any{},
			"content": want + ".md",
		} {
			if !reflect.DeepEqual(rec[key], wantVal) {
				t.Errorf("record[%q] = %#v, want %#v", key, rec[key], wantVal)
			}
		}
	})
}
