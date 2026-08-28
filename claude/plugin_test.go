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
	const thisProject = "/p/a"

	twoProjects := writeInstalledPlugins(t, `{"plugins": {
		"go@go-skills": [{"scope": "user", "version": "3.0.0"}],
		"bit@bit-pro": [
			{"scope": "project", "projectPath": "/p/a", "version": "0.1.0"},
			{"scope": "project", "projectPath": "/p/b", "version": "0.2.0"}
		]
	}}`)
	malformed := writeInstalledPlugins(t, `{`)
	userScope := writeInstalledPlugins(t, `{"plugins": {"bit@bit-pro": [{"scope": "user", "version": "0.1.0"}]}}`)
	empty := writeInstalledPlugins(t, `{"plugins": {}}`)
	missing := t.TempDir()

	tests := []struct {
		name        string
		home        string
		projectRoot string
		want        string
		wantOK      bool
	}{
		{"this project", twoProjects, thisProject, verInstalled, true},
		{"another project", twoProjects, "/p/b", verLatest, true},
		{"no matching project", twoProjects, "/p/c", "", false},
		{"file absent", missing, thisProject, "", false},
		{"file malformed", malformed, thisProject, "", false},
		{"record has no project path", userScope, thisProject, "", false},
		{"no plugins recorded", empty, thisProject, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := InstalledVersion(tt.home, tt.projectRoot)

			if got != tt.want || ok != tt.wantOK {
				t.Errorf("InstalledVersion(%q, %q) = (%q, %v), want (%q, %v)", tt.home, tt.projectRoot, got, ok, tt.want, tt.wantOK)
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

func installRecordFor(projectRoot string) string {
	return `{"plugins": {"bit@bit-pro": [
		{"scope": "project", "projectPath": "` + projectRoot + `", "version": "0.1.0"}
	]}}`
}

func TestPluginState_ReportsThisProject(t *testing.T) {
	home := t.TempDir()
	projectRoot := t.TempDir()

	installRecordAt(t, home, installRecordFor(projectRoot))
	marketplaceManifestAt(t, home, `{"name": "bit", "version": "0.2.0"}`)

	installed, latest, ok := PluginState(home, projectRoot)

	if installed != verInstalled || latest != verLatest || !ok {
		t.Errorf("PluginState(%q, %q) = (%q, %q, %v), want (%q, %q, %v)",
			home, projectRoot, installed, latest, ok, verInstalled, verLatest, true)
	}
}

func TestPluginState_SilentWhenEitherReadFails(t *testing.T) {
	projectRoot := t.TempDir()

	noClone := t.TempDir()
	installRecordAt(t, noClone, installRecordFor(projectRoot))

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
			installed, latest, ok := PluginState(tt.home, projectRoot)

			if installed != "" || latest != "" || ok {
				t.Errorf("PluginState(%q, %q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.home, projectRoot, installed, latest, ok, "", "", false)
			}
		})
	}
}

func TestStart_DoesNotWaitForTheChild(t *testing.T) {
	began := time.Now()

	if err := start("sleep", "3"); err != nil {
		t.Fatalf("start(sleep 3) returned error: %v", err)
	}

	if elapsed := time.Since(began); elapsed >= time.Second {
		t.Errorf("start took %v, want it to return without waiting for the child", elapsed)
	}
}

func TestStart_MissingBinaryIsSilent(t *testing.T) {
	if err := start("bp-no-such-binary-exists"); err != nil {
		t.Errorf("start of a missing binary returned error %v, want nil", err)
	}
}
