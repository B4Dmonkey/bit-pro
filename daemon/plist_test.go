package daemon

import (
	"regexp"
	"testing"
)

func TestPlist_KeepAliveRestartsOnCrashOnly(t *testing.T) {
	plist := string(Plist("/usr/local/bin/bp", "/tmp/daemon.log"))

	onCrashOnly := regexp.MustCompile(`<key>KeepAlive</key>\s*<dict>\s*<key>SuccessfulExit</key>\s*<false/>\s*</dict>`)
	if !onCrashOnly.MatchString(plist) {
		t.Errorf("plist does not match %s:\n%s", onCrashOnly, plist)
	}

	always := regexp.MustCompile(`<key>KeepAlive</key>\s*<true/>`)
	if always.MatchString(plist) {
		t.Errorf("plist restarts unconditionally, matching %s:\n%s", always, plist)
	}
}
