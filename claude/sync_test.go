package claude

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
)

type recorder struct {
	calls [][]string
	errs  map[int]error
}

func newRecorder(errs map[int]error) *recorder {
	return &recorder{errs: errs}
}

func (r *recorder) Run(_ context.Context, name string, args ...string) error {
	r.calls = append(r.calls, append([]string{name}, args...))
	return r.errs[len(r.calls)-1]
}

func TestSyncPlugin_RefreshesThenUpdates(t *testing.T) {
	rec := newRecorder(nil)

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

func TestSyncPlugin_ReportsWhenInstallAlsoFails(t *testing.T) {
	rec := newRecorder(map[int]error{
		1: errors.New("plugin bit@bit-pro is not installed"),
		2: errors.New("marketplace bit-pro is unreachable"),
	})

	err := SyncPlugin(t.Context(), rec.Run)
	if err == nil {
		t.Fatal("SyncPlugin returned nil error, want non-nil")
	}

	if !strings.Contains(err.Error(), "marketplace bit-pro is unreachable") {
		t.Errorf("err = %v, want it to carry the install failure", err)
	}
	if strings.Contains(err.Error(), "is not installed") {
		t.Errorf("err = %v, want it not to carry the swallowed update failure", err)
	}
}

func TestSyncPlugin_StopsWhenTheCatalogRefreshFails(t *testing.T) {
	rec := newRecorder(map[int]error{0: errors.New("claude exited 1")})

	err := SyncPlugin(t.Context(), rec.Run)
	if err == nil {
		t.Fatal("SyncPlugin returned nil error, want non-nil")
	}

	if len(rec.calls) != 1 {
		t.Errorf("calls = %v, want only the marketplace refresh", rec.calls)
	}
}

func TestSyncPlugin_FallsBackToInstall(t *testing.T) {
	rec := newRecorder(map[int]error{1: errors.New("plugin bit@bit-pro is not installed")})

	if err := SyncPlugin(t.Context(), rec.Run); err != nil {
		t.Fatalf("SyncPlugin returned error: %v", err)
	}

	want := [][]string{
		{"claude", "plugin", "marketplace", "update", "bit-pro"},
		{"claude", "plugin", "update", "bit@bit-pro", "--scope", "project"},
		{"claude", "plugin", "install", "bit@bit-pro", "--scope", "project"},
	}
	if !slices.EqualFunc(rec.calls, want, slices.Equal) {
		t.Errorf("calls = %v, want %v", rec.calls, want)
	}
}
