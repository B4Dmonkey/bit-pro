package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/launchd"
)

const editedPlist = "not a plist"

func TestStartCmd_WritesThePlistOnlyWhenMissing(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		check    func(t *testing.T, home, plist string)
	}{
		{
			name:  "no plist on disk",
			check: assertEnrollsTheDaemon,
		},
		{
			name:     "a plist the operator edited",
			existing: editedPlist,
			check: func(t *testing.T, _, plist string) {
				t.Helper()

				if plist != editedPlist {
					t.Errorf("plist contents = %q, want %q", plist, editedPlist)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_DATA_HOME", "")

			path := filepath.Join(home, "Library", "LaunchAgents", launchd.Label+".plist")
			if tt.existing != "" {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("os.MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
				}

				if err := os.WriteFile(path, []byte(tt.existing), 0o600); err != nil {
					t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
				}
			}

			if _, err := runWithLaunchd(t, nothingLoaded, startCmdUse); err != nil {
				t.Fatalf("bp start returned error: %v", err)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
			}

			tt.check(t, home, string(data))
		})
	}
}

func assertEnrollsTheDaemon(t *testing.T, home, plist string) {
	t.Helper()

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() returned error: %v", err)
	}

	storeDir := filepath.Join(home, ".local", "share", "bit-pro")
	logPath := regexp.QuoteMeta(filepath.Join(storeDir, "daemon.log"))

	for _, want := range []string{launchd.Label, exe, "<string>serve</string>"} {
		if !strings.Contains(plist, want) {
			t.Errorf("plist does not contain %q:\n%s", want, plist)
		}
	}

	wants := []*regexp.Regexp{
		regexp.MustCompile(`<key>StandardOutPath</key>\s*<string>` + logPath + `</string>`),
		regexp.MustCompile(`<key>StandardErrorPath</key>\s*<string>` + logPath + `</string>`),
		regexp.MustCompile(`<key>RunAtLoad</key>\s*<true/>`),
		regexp.MustCompile(`<key>KeepAlive</key>\s*<dict>\s*<key>SuccessfulExit</key>\s*<false/>\s*</dict>`),
	}
	for _, want := range wants {
		if !want.MatchString(plist) {
			t.Errorf("plist does not match %s:\n%s", want, plist)
		}
	}

	info, err := os.Stat(storeDir)
	if err != nil {
		t.Fatalf("os.Stat(%q) returned error: %v", storeDir, err)
	}

	if !info.IsDir() {
		t.Errorf("%s is not a directory", storeDir)
	}
}
