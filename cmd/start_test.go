package cmd

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/launchd"
)

const (
	editedPlist = "not a plist"
	startedPID  = "started (pid 4242)\n"
)

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

func TestStartCmd_LogPathFollowsXDGDataHome(t *testing.T) {
	home := t.TempDir()
	data := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", data)

	if _, err := runWithLaunchd(t, nothingLoaded, startCmdUse); err != nil {
		t.Fatalf("bp start returned error: %v", err)
	}

	path := filepath.Join(home, "Library", "LaunchAgents", launchd.Label+".plist")

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
	}

	plist := string(contents)

	want := filepath.Join(data, "bit-pro", "daemon.log")
	if !strings.Contains(plist, want) {
		t.Errorf("plist does not contain %q:\n%s", want, plist)
	}

	if strings.Contains(plist, filepath.Join(".local", "share")) {
		t.Errorf("plist still points at the home fallback:\n%s", plist)
	}
}

func enableCall() string {
	return "launchctl enable gui/" + strconv.Itoa(os.Getuid()) + "/" + launchd.Label
}

func bootstrapCall(home string) string {
	plist := filepath.Join(home, "Library", "LaunchAgents", launchd.Label+".plist")

	return "launchctl bootstrap gui/" + strconv.Itoa(os.Getuid()) + " " + plist
}

func TestStartCmd_EnablesBeforeBootstrapping(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	var calls []string

	lc := func(_ context.Context, name string, args ...string) (string, int, error) {
		calls = append(calls, strings.Join(append([]string{name}, args...), " "))

		switch args[0] {
		case printDisabled:
			return disabledStore(), 0, nil
		case "list":
			if !slices.Contains(calls, bootstrapCall(home)) {
				return "", 113, nil
			}

			return launchctlDict, 0, nil
		default:
			return "", 0, nil
		}
	}

	out, err := runWithLaunchd(t, lc, startCmdUse)
	if err != nil {
		t.Fatalf("bp start returned error: %v", err)
	}

	if out != startedPID {
		t.Errorf("bp start output = %q, want %q", out, startedPID)
	}

	enable := slices.Index(calls, enableCall())
	bootstrap := slices.Index(calls, bootstrapCall(home))

	if enable < 0 || bootstrap < 0 || enable > bootstrap {
		t.Errorf("launchctl calls = %v, want enable before bootstrap", calls)
	}
}
