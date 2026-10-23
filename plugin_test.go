package main

import (
	"encoding/json"
	"maps"
	"os"
	"slices"
	"testing"
)

type pluginHooks struct {
	Hooks map[string][]struct {
		Hooks []struct {
			Type    string `json:"type"`
			Command string `json:"command"`
			Async   bool   `json:"async"`
		} `json:"hooks"`
	} `json:"hooks"`
}

func TestPluginHooks(t *testing.T) {
	t.Run("registers bp receipt on both stop events", func(t *testing.T) {
		raw, err := os.ReadFile("bit/hooks/hooks.json")
		if err != nil {
			t.Fatal(err)
		}

		var got pluginHooks
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatal(err)
		}

		events := slices.Sorted(maps.Keys(got.Hooks))
		if want := []string{"Stop", "SubagentStop"}; !slices.Equal(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}

		for _, event := range events {
			groups := got.Hooks[event]
			if len(groups) != 1 || len(groups[0].Hooks) != 1 {
				t.Fatalf("%s: want one group with one hook, got %+v", event, groups)
			}

			h := groups[0].Hooks[0]
			if h.Type != "command" || h.Command != "bp receipt" || !h.Async {
				t.Errorf("%s hook = %+v, want async command \"bp receipt\"", event, h)
			}
		}
	})
}
