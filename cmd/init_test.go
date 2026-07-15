package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestInitCmd(t *testing.T) {
	tests := []struct {
		name string
		runs int
	}{
		{name: "fresh directory", runs: 1},
		{name: "already initialized", runs: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)

			var err error
			for i := 0; i < tt.runs; i++ {
				rootCmd := NewRootCmd()
				rootCmd.SetArgs([]string{"init", "--prefix", "BIT"})
				err = rootCmd.Execute()
			}

			if err != nil {
				t.Fatalf("Execute() returned error: %v", err)
			}
			info, statErr := os.Stat(filepath.Join(dir, ".bit"))
			if statErr != nil {
				t.Fatalf("os.Stat(.bit) returned error: %v", statErr)
			}
			if !info.IsDir() {
				t.Errorf(".bit exists but is not a directory")
			}
		})
	}
}

func TestInitCmd_WritesConfigWithPrefix(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".bit", "config.toml"))
	if err != nil {
		t.Fatalf("os.ReadFile(.bit/config.toml) returned error: %v", err)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		t.Fatalf("toml.Decode returned error: %v", err)
	}
	if cfg.Prefix != "BIT" {
		t.Errorf("cfg.Prefix = %q, want %q", cfg.Prefix, "BIT")
	}
}

func TestInitCmd_ErrorsOnEmptyPrefix(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := NewRootCmd()
	rootCmd.SetIn(strings.NewReader("\n"))
	rootCmd.SetArgs([]string{"init"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Execute() returned nil error, want error for empty prefix")
	}

	if _, statErr := os.Stat(filepath.Join(dir, ".bit", "config.toml")); !os.IsNotExist(statErr) {
		t.Errorf("os.Stat(.bit/config.toml) = %v, want IsNotExist", statErr)
	}
}

func TestInitCmd_PromptsForPrefixWhenFlagOmitted(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	rootCmd := NewRootCmd()
	rootCmd.SetIn(strings.NewReader("BIT\n"))
	rootCmd.SetArgs([]string{"init"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".bit", "config.toml"))
	if err != nil {
		t.Fatalf("os.ReadFile(.bit/config.toml) returned error: %v", err)
	}

	var cfg Config
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		t.Fatalf("toml.Decode returned error: %v", err)
	}
	if cfg.Prefix != "BIT" {
		t.Errorf("cfg.Prefix = %q, want %q", cfg.Prefix, "BIT")
	}
}
