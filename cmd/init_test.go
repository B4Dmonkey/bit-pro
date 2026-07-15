package cmd

import (
	"os"
	"path/filepath"
	"testing"
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
				rootCmd.SetArgs([]string{"init"})
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
