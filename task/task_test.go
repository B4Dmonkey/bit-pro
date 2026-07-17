package task

import (
	"strings"
	"testing"
)

func TestTaskBytes_WritesFrontmatterThenBody(t *testing.T) {
	t.Parallel()

	tk := &Task{
		ID:     "BIT-1",
		Title:  "Set up init wizard",
		Status: "todo",
		Body:   "Add flags for prefix capture.",
	}

	got, err := tk.Bytes()
	if err != nil {
		t.Fatalf("Bytes() returned error: %v", err)
	}

	want := "---\nid: BIT-1\ntitle: Set up init wizard\nstatus: todo\n---\nAdd flags for prefix capture."
	if string(got) != want {
		t.Errorf("Bytes() = %q, want %q", got, want)
	}
}

func TestParse_RoundTripsThroughBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		task Task
	}{
		{name: "single line body", task: Task{ID: "BIT-1", Title: "One", Status: "todo", Body: "Body."}},
		{name: "multi line body", task: Task{ID: "BIT-2", Title: "Two", Status: "doing", Body: "Line one.\nLine two.\n"}},
		{name: "empty body", task: Task{ID: "BIT-3", Title: "Three", Status: "done", Body: ""}},
		{name: "body containing a delimiter", task: Task{ID: "BIT-4", Title: "Four", Status: "todo", Body: "before\n---\nafter"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := tt.task.Bytes()
			if err != nil {
				t.Fatalf("Bytes() returned error: %v", err)
			}

			got, err := Parse(data)
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}
			if *got != tt.task {
				t.Errorf("Parse(Bytes()) = %+v, want %+v", *got, tt.task)
			}
		})
	}
}

func TestParse_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    string
		wantErr string
	}{
		{name: "no frontmatter at all", data: "just a body\n", wantErr: "missing frontmatter delimiter"},
		{name: "opening delimiter only", data: "---\nid: BIT-1\n", wantErr: "missing closing frontmatter delimiter"},
		{name: "empty input", data: "", wantErr: "missing frontmatter delimiter"},
		{name: "malformed yaml frontmatter", data: "---\nid: [unclosed\n---\nbody", wantErr: "parsing task frontmatter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := Parse([]byte(tt.data))
			if err == nil {
				t.Fatalf("Parse(%q) returned nil error, want %q", tt.data, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Parse(%q) error = %q, want it to contain %q", tt.data, err, tt.wantErr)
			}
		})
	}
}
