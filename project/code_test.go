package project

import (
	"errors"
	"testing"
)

func TestValidateCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		want    string
		wantErr error
	}{
		{name: "lowercase is uppercased", code: "bit", want: "BIT"},
		{name: "digits after a letter", code: "ACME2", want: "ACME2"},
		{name: "surrounding space is trimmed", code: "  ex ", want: "EX"},
		{name: "empty", code: "", wantErr: ErrInvalidCode},
		{name: "leading digit", code: "2BIT", wantErr: ErrInvalidCode},
		{name: "hyphen", code: "BIT-PRO", wantErr: ErrInvalidCode},
		{name: "underscore", code: "BIT_1", wantErr: ErrInvalidCode},
		{name: "traversal", code: "../X", wantErr: ErrInvalidCode},
		{name: "reserved feedback in lowercase", code: "feedback", wantErr: ErrReservedCode},
		{name: "reserved retro in mixed case", code: "Retro", wantErr: ErrReservedCode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateCode(tt.code)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateCode(%q) error = %v, want %v", tt.code, err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("ValidateCode(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}
