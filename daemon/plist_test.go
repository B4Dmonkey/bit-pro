package daemon

import (
	"regexp"
	"strings"
	"testing"
)

func TestPlist_ProgramArgumentsIncludesDaemon(t *testing.T) {
	plist := string(Plist("/usr/local/bin/bp", "/tmp/daemon.log", "/usr/local/bin/claude"))

	for _, want := range []string{"<string>serve</string>", "<string>daemon</string>"} {
		if !strings.Contains(plist, want) {
			t.Errorf("plist does not contain %q:\n%s", want, plist)
		}
	}

	serveIdx := strings.Index(plist, "<string>serve</string>")

	daemonIdx := strings.Index(plist, "<string>daemon</string>")
	if serveIdx < 0 || daemonIdx < 0 || serveIdx > daemonIdx {
		t.Errorf("expected <string>serve</string> before <string>daemon</string> in plist:\n%s", plist)
	}

	count := strings.Count(plist, "<string>serve</string>") + strings.Count(plist, "<string>daemon</string>")
	if count != 2 {
		t.Errorf("ProgramArguments count = %d, want 2 (separate serve and daemon entries):\n%s", count, plist)
	}
}

func TestPlist_KeepAliveRestartsOnCrashOnly(t *testing.T) {
	plist := string(Plist("/usr/local/bin/bp", "/tmp/daemon.log", "/usr/local/bin/claude"))

	onCrashOnly := regexp.MustCompile(`<key>KeepAlive</key>\s*<dict>\s*<key>SuccessfulExit</key>\s*<false/>\s*</dict>`)
	if !onCrashOnly.MatchString(plist) {
		t.Errorf("plist does not match %s:\n%s", onCrashOnly, plist)
	}

	always := regexp.MustCompile(`<key>KeepAlive</key>\s*<true/>`)
	if always.MatchString(plist) {
		t.Errorf("plist restarts unconditionally, matching %s:\n%s", always, plist)
	}
}
