package project

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrInvalidCode  = errors.New("must be letters and digits, starting with a letter")
	ErrReservedCode = errors.New("is reserved")
)

var codePattern = regexp.MustCompile(`^[A-Z][A-Z0-9]*$`)

var reservedCodes = map[string]bool{"FEEDBACK": true, "RETRO": true}

func ValidateCode(code string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))

	if !codePattern.MatchString(normalized) {
		return "", fmt.Errorf("project code %q: %w", code, ErrInvalidCode)
	}

	if reservedCodes[normalized] {
		return "", fmt.Errorf("project code %q: %w", code, ErrReservedCode)
	}

	return normalized, nil
}
