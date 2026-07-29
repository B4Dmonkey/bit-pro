package claude

import (
	"context"
	"errors"
	"slices"
	"testing"
)

type recorder struct {
	calls  [][]string
	failAt int
}

func newRecorder() *recorder {
	return &recorder{failAt: -1}
}

func (r *recorder) Run(_ context.Context, name string, args ...string) error {
	r.calls = append(r.calls, append([]string{name}, args...))
	if len(r.calls)-1 == r.failAt {
		return errors.New("claude exited 1")
	}
	return nil
}

func TestSyncPlugin_RefreshesThenUpdates(t *testing.T) {
	rec := newRecorder()

	if err := SyncPlugin(t.Context(), rec.Run); err != nil {
		t.Fatalf("SyncPlugin returned error: %v", err)
	}

	want := [][]string{
		{"claude", "plugin", "marketplace", "update", "bit-pro"},
		{"claude", "plugin", "update", "bit@bit-pro", "--scope", "project"},
	}
	if !slices.EqualFunc(rec.calls, want, slices.Equal) {
		t.Errorf("calls = %v, want %v", rec.calls, want)
	}
}

func TestSyncPlugin_StopsWhenTheCatalogRefreshFails(t *testing.T) {
	rec := newRecorder()
	rec.failAt = 0

	err := SyncPlugin(t.Context(), rec.Run)
	if err == nil {
		t.Fatal("SyncPlugin returned nil error, want non-nil")
	}

	if len(rec.calls) != 1 {
		t.Errorf("calls = %v, want only the marketplace refresh", rec.calls)
	}
}
