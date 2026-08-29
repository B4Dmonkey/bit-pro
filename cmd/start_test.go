package cmd

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/B4Dmonkey/bit-pro/daemon"
)

const (
	editedPlist       = "not a plist"
	stalePlist        = "<string>serve</string>"
	startedPID        = "started (pid 4242)\n"
	alreadyRunningPID = "already running (pid 4242)\n"
)

func TestStartCmd_EnrollsOrRepairsThePlist(t *testing.T) {
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
			check:    assertEnrollsTheDaemon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_DATA_HOME", "")

			stubClaudeOnPath(t)

			path := filepath.Join(home, "Library", "LaunchAgents", daemon.Label+".plist")
			if tt.existing != "" {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("os.MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
				}

				if err := os.WriteFile(path, []byte(tt.existing), 0o600); err != nil {
					t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
				}
			}

			if _, err := runWithDaemon(t, nothingLoaded, startCmdUse); err != nil {
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

	for _, want := range []string{daemon.Label, exe, "<string>serve</string>", "<string>daemon</string>"} {
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

func TestStartCmd_RepairsAStalePlist(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	stubClaudeOnPath(t)

	path := filepath.Join(home, "Library", "LaunchAgents", daemon.Label+".plist")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(stalePlist), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
	}

	var calls []string

	out, err := runWithDaemon(t, recordingLaunchctl(&calls, bootstrapCall(home), "", 113), startCmdUse)
	if err != nil {
		t.Fatalf("bp start returned error: %v", err)
	}

	bootout := slices.Index(calls, bootoutCall())
	bootstrap := slices.Index(calls, bootstrapCall(home))

	if bootout < 0 || bootstrap < 0 || bootout > bootstrap {
		t.Errorf("launchctl calls = %v, want bootout before bootstrap", calls)
	}

	if out != startedPID {
		t.Errorf("bp start output = %q, want %q", out, startedPID)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
	}

	assertEnrollsTheDaemon(t, home, string(data))
}

func TestStartCmd_LogPathFollowsXDGDataHome(t *testing.T) {
	home := t.TempDir()
	data := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", data)

	stubClaudeOnPath(t)

	if _, err := runWithDaemon(t, nothingLoaded, startCmdUse); err != nil {
		t.Fatalf("bp start returned error: %v", err)
	}

	path := filepath.Join(home, "Library", "LaunchAgents", daemon.Label+".plist")

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
	return "launchctl enable gui/" + strconv.Itoa(os.Getuid()) + "/" + daemon.Label
}

func bootstrapCall(home string) string {
	plist := filepath.Join(home, "Library", "LaunchAgents", daemon.Label+".plist")

	return "launchctl bootstrap gui/" + strconv.Itoa(os.Getuid()) + " " + plist
}

func recordingLaunchctl(calls *[]string, action, out string, code int) daemon.Runner {
	return func(_ context.Context, name string, args ...string) (string, int, error) {
		*calls = append(*calls, strings.Join(append([]string{name}, args...), " "))

		switch args[0] {
		case printDisabled:
			return disabledStore(), 0, nil
		case listSubcmd:
			if action != "" && !slices.Contains(*calls, action) {
				return out, code, nil
			}

			return launchctlDict, 0, nil
		default:
			return "", 0, nil
		}
	}
}

func assertNoCallContains(t *testing.T, calls []string, unwanted string) {
	t.Helper()

	if slices.ContainsFunc(calls, func(c string) bool { return strings.Contains(c, unwanted) }) {
		t.Errorf("launchctl calls = %v, want no %s call", calls, unwanted)
	}
}

func TestStartCmd_EnablesBeforeBootstrapping(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	stubClaudeOnPath(t)

	var calls []string

	out, err := runWithDaemon(t, recordingLaunchctl(&calls, bootstrapCall(home), "", 113), startCmdUse)
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

func kickstartCall() string {
	return "launchctl kickstart gui/" + strconv.Itoa(os.Getuid()) + "/" + daemon.Label
}

func TestStartCmd_ReconcilesTheStateItFinds(t *testing.T) {
	tests := []struct {
		name   string
		out    string
		code   int
		action func(home string) string
		want   string
	}{
		{
			name: "already running",
			out:  launchctlDict,
			want: alreadyRunningPID,
		},
		{
			name:   "loaded but not running",
			out:    strings.ReplaceAll(launchctlDict, "\t\"PID\" = 4242;\n", ""),
			action: func(string) string { return kickstartCall() },
			want:   startedPID,
		},
		{
			name:   "not loaded",
			code:   113,
			action: bootstrapCall,
			want:   startedPID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("XDG_DATA_HOME", "")

			stubClaudeOnPath(t)

			var action string
			if tt.action != nil {
				action = tt.action(home)
			}

			var calls []string

			out, err := runWithDaemon(t, recordingLaunchctl(&calls, action, tt.out, tt.code), startCmdUse)
			if err != nil {
				t.Fatalf("bp start returned error: %v", err)
			}

			if out != tt.want {
				t.Errorf("bp start output = %q, want %q", out, tt.want)
			}

			if action != "" && !slices.Contains(calls, action) {
				t.Errorf("launchctl calls = %v, want %q", calls, action)
			}

			for _, unwanted := range []string{"bootstrap", "kickstart"} {
				if !strings.Contains(action, unwanted) {
					assertNoCallContains(t, calls, unwanted)
				}
			}
		})
	}
}

func stubClaudeOnPath(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	bin := filepath.Join(dir, "claude")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) returned error: %v", bin, err)
	}

	if err := os.Chmod(bin, 0o700); err != nil {
		t.Fatalf("os.Chmod(%q) returned error: %v", bin, err)
	}

	t.Setenv("PATH", dir)

	return bin
}

func TestStartCmd_PinsTheResolvedClaudeInThePlist(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")

	bin := stubClaudeOnPath(t)

	if _, err := runWithDaemon(t, nothingLoaded, startCmdUse); err != nil {
		t.Fatalf("bp start returned error: %v", err)
	}

	path := filepath.Join(home, "Library", "LaunchAgents", daemon.Label+".plist")

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) returned error: %v", path, err)
	}

	want := regexp.MustCompile(`<key>EnvironmentVariables</key>\s*<dict>\s*<key>BP_CLAUDE</key>\s*<string>` +
		regexp.QuoteMeta(bin) + `</string>\s*</dict>`)
	if !want.MatchString(string(contents)) {
		t.Errorf("plist does not match %s:\n%s", want, contents)
	}
}

func TestStartCmd_FailsWhenClaudeIsNotOnThePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("PATH", t.TempDir())

	_, err := runWithDaemon(t, nothingLoaded, startCmdUse)
	if err == nil {
		t.Fatal("bp start returned nil error, want an error naming claude")
	}

	if !strings.Contains(err.Error(), "claude") {
		t.Errorf("bp start error = %q, want it to name claude", err)
	}

	path := filepath.Join(home, "Library", "LaunchAgents", daemon.Label+".plist")
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("os.Stat(%q) error = %v, want fs.ErrNotExist", path, err)
	}
}
