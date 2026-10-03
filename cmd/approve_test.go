package cmd

import (
	"testing"
)

func TestApproveCmd(t *testing.T) {
	t.Run("sets approved true", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")

		mustRun(t, "approve", "BIT-1")

		got, err := projectStore(t).Load("BIT-1")
		if err != nil {
			t.Fatalf("loading BIT-1: %v", err)
		}

		if !got.Approved {
			t.Error("expected Approved = true, got false")
		}
	})

	t.Run("unapprove clears", func(t *testing.T) {
		initProject(t, "BIT")
		createTask(t, "Track", "...")

		mustRun(t, "approve", "BIT-1")
		mustRun(t, "unapprove", "BIT-1")

		got, err := projectStore(t).Load("BIT-1")
		if err != nil {
			t.Fatalf("loading BIT-1: %v", err)
		}

		if got.Approved {
			t.Error("expected Approved = false after unapprove, got true")
		}
	})

	t.Run("errors on unknown id", func(t *testing.T) {
		initProject(t, "BIT")

		_, err := run(t, "approve", "BIT-99")
		if err == nil {
			t.Error("expected error for unknown task BIT-99, got nil")
		}
	})
}
