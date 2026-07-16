package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaskCreateCmd_WritesFirstTask(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	createCmd := NewRootCmd()
	createCmd.SetArgs([]string{"task", "create", "Set up init wizard", "--description", "Add flags for prefix capture."})
	if err := createCmd.Execute(); err != nil {
		t.Fatalf("task create Execute() returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".bit", "tasks", "BIT-1.md"))
	if err != nil {
		t.Fatalf("os.ReadFile(.bit/tasks/BIT-1.md) returned error: %v", err)
	}

	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) != 3 {
		t.Fatalf("content = %q, want three `---`-delimited parts", string(data))
	}
	frontmatter, body := parts[1], parts[2]

	if !strings.Contains(frontmatter, "id: BIT-1") {
		t.Errorf("frontmatter = %q, want to contain %q", frontmatter, "id: BIT-1")
	}
	if !strings.Contains(frontmatter, "title: Set up init wizard") {
		t.Errorf("frontmatter = %q, want to contain %q", frontmatter, "title: Set up init wizard")
	}
	if !strings.Contains(frontmatter, "status: todo") {
		t.Errorf("frontmatter = %q, want to contain %q", frontmatter, "status: todo")
	}
	if !strings.Contains(body, "Add flags for prefix capture.") {
		t.Errorf("body = %q, want to contain %q", body, "Add flags for prefix capture.")
	}
}

func TestTaskCreateCmd_ErrorsWithoutTitle(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	createCmd := NewRootCmd()
	createCmd.SetArgs([]string{"task", "create"})
	if err := createCmd.Execute(); err == nil {
		t.Fatal("Execute() returned nil error, want error for missing title argument")
	}
}

func TestTaskCreateCmd_AssignsNextIDWhenTasksExist(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	initCmd := NewRootCmd()
	initCmd.SetArgs([]string{"init", "--prefix", "BIT"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init Execute() returned error: %v", err)
	}

	tasksPath := filepath.Join(dir, ".bit", "tasks")
	if err := os.MkdirAll(tasksPath, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%s) returned error: %v", tasksPath, err)
	}
	existing := map[string]string{
		"BIT-1.md": "---\nid: BIT-1\ntitle: First\nstatus: todo\n---\n",
		"BIT-3.md": "---\nid: BIT-3\ntitle: Third\nstatus: todo\n---\n",
	}
	for name, content := range existing {
		if err := os.WriteFile(filepath.Join(tasksPath, name), []byte(content), 0o644); err != nil {
			t.Fatalf("os.WriteFile(%s) returned error: %v", name, err)
		}
	}

	createCmd := NewRootCmd()
	createCmd.SetArgs([]string{"task", "create", "Third real task", "--description", "..."})
	if err := createCmd.Execute(); err != nil {
		t.Fatalf("task create Execute() returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tasksPath, "BIT-4.md")); err != nil {
		t.Errorf("os.Stat(BIT-4.md) returned error: %v, want file to exist", err)
	}

	for name, want := range existing {
		got, err := os.ReadFile(filepath.Join(tasksPath, name))
		if err != nil {
			t.Fatalf("os.ReadFile(%s) returned error: %v", name, err)
		}
		if string(got) != want {
			t.Errorf("%s content = %q, want unchanged %q", name, got, want)
		}
	}
}

func TestTaskCreateCmd_ErrorsWithoutConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	createCmd := NewRootCmd()
	createCmd.SetArgs([]string{"task", "create", "Foo"})
	if err := createCmd.Execute(); err == nil {
		t.Fatal("Execute() returned nil error, want error when config.toml is absent")
	}

	if _, statErr := os.Stat(filepath.Join(dir, ".bit", "tasks")); !os.IsNotExist(statErr) {
		t.Errorf("os.Stat(.bit/tasks) = %v, want IsNotExist", statErr)
	}
}
