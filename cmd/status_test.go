package cmd

import "testing"

func TestStatusCmd_ReportsNotRunning(t *testing.T) {
	out, err := run(t, statusCmdUse)
	if err != nil {
		t.Fatalf("bp status returned error: %v", err)
	}

	if want := "not running\n"; out != want {
		t.Errorf("bp status output = %q, want %q", out, want)
	}
}
