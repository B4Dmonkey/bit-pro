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

func TestRootCmd_Version(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()

	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	want := "bit version 0.1.0-dev\n"
	if buf.String() != want {
		t.Errorf("version output = %q, want %q", buf.String(), want)
	}
}
