package cmd

import (
	"testing"

	"github.com/B4Dmonkey/bit-pro/task"
)

func TestApproveCmd_SetsApprovedTrue(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")

	mustRun(t, "approve", "BIT-1")

	got, err := task.New(".bit").Load("BIT-1")
	if err != nil {
		t.Fatalf("loading BIT-1: %v", err)
	}
	if !got.Approved {
		t.Error("expected Approved = true, got false")
	}
}

func TestApproveCmd_UnapproveClears(t *testing.T) {
	initProject(t, "BIT")
	createTask(t, "Track", "...")

	mustRun(t, "approve", "BIT-1")
	mustRun(t, "unapprove", "BIT-1")

	got, err := task.New(".bit").Load("BIT-1")
	if err != nil {
		t.Fatalf("loading BIT-1: %v", err)
	}
	if got.Approved {
		t.Error("expected Approved = false after unapprove, got true")
	}
}

func TestApproveCmd_ErrorsOnUnknownID(t *testing.T) {
	initProject(t, "BIT")

	_, err := run(t, "approve", "BIT-99")
	if err == nil {
		t.Error("expected error for unknown task BIT-99, got nil")
	}
}
