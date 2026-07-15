package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_Help(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()

	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "bit") {
		t.Errorf("help output missing command name %q, got:\n%s", "bit", buf.String())
	}
}
