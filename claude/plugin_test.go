package claude

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	verInstalled = "0.1.0"
	verLatest    = "0.2.0"
)

func writeFixture(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
	}
}

func installRecordAt(t *testing.T, home, contents string) {
	t.Helper()

	writeFixture(t, filepath.Join(home, ".claude", "plugins", "installed_plugins.json"), contents)
}

func writeInstalledPlugins(t *testing.T, contents string) string {
	t.Helper()

	home := t.TempDir()
	installRecordAt(t, home, contents)

	return home
}

func TestInstalledVersion(t *testing.T) {
	userOnly := writeInstalledPlugins(t, `{"plugins": {"bit@bit-pro": [{"scope": "user", "version": "0.1.0"}]}}`)
	userBesideProject := writeInstalledPlugins(t, `{"plugins": {"bit@bit-pro": [
		{"scope": "project", "projectPath": "/p/a", "version": "0.1.0"},
		{"scope": "user", "version": "0.2.0"}
	]}}`)
	projectsOnly := writeInstalledPlugins(t, `{"plugins": {"bit@bit-pro": [
		{"scope": "project", "projectPath": "/p/a", "version": "0.1.0"},
		{"scope": "project", "projectPath": "/p/b", "version": "0.2.0"}
	]}}`)
	otherPlugin := writeInstalledPlugins(t, `{"plugins": {"go@go-skills": [{"scope": "user", "version": "3.0.0"}]}}`)
	malformed := writeInstalledPlugins(t, `{`)
	empty := writeInstalledPlugins(t, `{"plugins": {}}`)
	missing := t.TempDir()

	tests := []struct {
		name   string
		home   string
		want   string
		wantOK bool
	}{
		{"user only", userOnly, verInstalled, true},
		{"user beside a project install", userBesideProject, verLatest, true},
		{"project installs only", projectsOnly, "", false},
		{"another plugin's user install only", otherPlugin, "", false},
		{"file absent", missing, "", false},
		{"file malformed", malformed, "", false},
		{"no plugins recorded", empty, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := InstalledVersion(tt.home)

			if got != tt.want || ok != tt.wantOK {
				t.Errorf("InstalledVersion(%q) = (%q, %v), want (%q, %v)", tt.home, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func marketplaceManifestAt(t *testing.T, home, contents string) {
	t.Helper()

	writeFixture(t, filepath.Join(home, ".claude", "plugins", "marketplaces", "bit-pro",
		"bit", ".claude-plugin", "plugin.json"), contents)
}

func writeMarketplaceManifest(t *testing.T, contents string) string {
	t.Helper()

	home := t.TempDir()
	marketplaceManifestAt(t, home, contents)

	return home
}

func TestLatestVersion(t *testing.T) {
	versioned := writeMarketplaceManifest(t, `{"name": "bit", "version": "0.2.0"}`)
	unversioned := writeMarketplaceManifest(t, `{"name": "bit", "author": {"name": "josiah"}}`)
	malformed := writeMarketplaceManifest(t, `{`)
	missing := t.TempDir()

	tests := []struct {
		name   string
		home   string
		want   string
		wantOK bool
	}{
		{"manifest declares a version", versioned, verLatest, true},
		{"manifest declares no version", unversioned, "", false},
		{"manifest malformed", malformed, "", false},
		{"manifest absent", missing, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := LatestVersion(tt.home)

			if got != tt.want || ok != tt.wantOK {
				t.Errorf("LatestVersion(%q) = (%q, %v), want (%q, %v)", tt.home, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

const userInstallRecord = `{"plugins": {"bit@bit-pro": [{"scope": "user", "version": "0.1.0"}]}}`

func TestPluginState(t *testing.T) {
	t.Run("reports the user install", func(t *testing.T) {
		home := t.TempDir()

		installRecordAt(t, home, userInstallRecord)
		marketplaceManifestAt(t, home, `{"name": "bit", "version": "0.2.0"}`)

		installed, latest, ok := PluginState(home)

		if installed != verInstalled || latest != verLatest || !ok {
			t.Errorf("PluginState(%q) = (%q, %q, %v), want (%q, %q, %v)",
				home, installed, latest, ok, verInstalled, verLatest, true)
		}
	})

	t.Run("silent when either read fails", func(t *testing.T) {
		noClone := t.TempDir()
		installRecordAt(t, noClone, userInstallRecord)

		noRecord := t.TempDir()
		marketplaceManifestAt(t, noRecord, `{"name": "bit", "version": "0.2.0"}`)

		tests := []struct {
			name string
			home string
		}{
			{"no marketplace clone", noClone},
			{"no install record", noRecord},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				installed, latest, ok := PluginState(tt.home)

				if installed != "" || latest != "" || ok {
					t.Errorf("PluginState(%q) = (%q, %q, %v), want (%q, %q, %v)",
						tt.home, installed, latest, ok, "", "", false)
				}
			})
		}
	})
}

func TestStart(t *testing.T) {
	t.Run("does not wait for the child", func(t *testing.T) {
		began := time.Now()

		if err := start("sleep", "3"); err != nil {
			t.Fatalf("start(sleep 3) returned error: %v", err)
		}

		if elapsed := time.Since(began); elapsed >= time.Second {
			t.Errorf("start took %v, want it to return without waiting for the child", elapsed)
		}
	})

	t.Run("missing binary is silent", func(t *testing.T) {
		if err := start("bp-no-such-binary-exists"); err != nil {
			t.Errorf("start of a missing binary returned error %v, want nil", err)
		}
	})
}
