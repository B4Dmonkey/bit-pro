package task

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreCounts_CountsCompletedTracks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T, s *Store)
		want  Counts
	}{
		{
			name: "a completed track and an active one",
			setup: func(t *testing.T, s *Store) {
				t.Helper()

				if err := s.Save(&Task{ID: acme1, Title: ttrack, Status: StatusTodo, Approved: true}); err != nil {
					t.Fatalf("Save: %v", err)
				}

				if err := os.MkdirAll(s.completedDir(), dirMode); err != nil {
					t.Fatalf("MkdirAll: %v", err)
				}

				for _, tk := range []*Task{
					{ID: "ACME-2", Title: ttrack, Status: StatusDone},
					{ID: "ACME-2.1", Title: tbar, Status: StatusDone},
				} {
					data, err := tk.Bytes()
					if err != nil {
						t.Fatalf("Bytes: %v", err)
					}

					if err := os.WriteFile(filepath.Join(s.completedDir(), tk.ID+".md"), data, fileMode); err != nil {
						t.Fatalf("WriteFile %s: %v", tk.ID, err)
					}
				}
			},
			want: Counts{Todo: 1, Completed: 1},
		},
		{
			name: "no completed directory",
			setup: func(t *testing.T, s *Store) {
				t.Helper()

				if err := s.Save(&Task{ID: acme1, Title: ttrack, Status: StatusTodo, Approved: true}); err != nil {
					t.Fatalf("Save: %v", err)
				}
			},
			want: Counts{Todo: 1},
		},
		{
			name: "only completed work",
			setup: func(t *testing.T, s *Store) {
				t.Helper()

				if err := os.MkdirAll(s.completedDir(), dirMode); err != nil {
					t.Fatalf("MkdirAll: %v", err)
				}

				data, err := (&Task{ID: acme1, Title: ttrack, Status: StatusDone}).Bytes()
				if err != nil {
					t.Fatalf("Bytes: %v", err)
				}

				if err := os.WriteFile(filepath.Join(s.completedDir(), acme1+".md"), data, fileMode); err != nil {
					t.Fatalf("WriteFile: %v", err)
				}
			},
			want: Counts{Completed: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := New(filepath.Join(t.TempDir(), ".bit"))
			tt.setup(t, s)

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
