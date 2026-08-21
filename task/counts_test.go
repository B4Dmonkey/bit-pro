package task

import (
	"path/filepath"
	"testing"
)

func TestStoreCounts_Buckets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		approved bool
		status   string
		want     Counts
	}{
		{name: "approved todo", approved: true, status: StatusTodo, want: Counts{Todo: 1}},
		{name: "approved doing", approved: true, status: StatusDoing, want: Counts{Todo: 1}},
		{name: "approved done", approved: true, status: StatusDone, want: Counts{Done: 1}},
		{name: "unapproved todo", approved: false, status: StatusTodo, want: Counts{Backlog: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(filepath.Join(t.TempDir(), ".bit"))
			if err := s.Save(&Task{ID: tid1, Title: ttrack, Status: tt.status, Approved: tt.approved}); err != nil {
				t.Fatalf("Save() returned error: %v", err)
			}

			got, err := s.Counts()
			if err != nil {
				t.Fatalf("Counts() returned error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Counts() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
