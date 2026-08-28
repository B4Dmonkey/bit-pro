package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func writeInstalledPlugins(t *testing.T, contents string) string {
	t.Helper()

	home := t.TempDir()
	path := filepath.Join(home, ".claude", "plugins", "installed_plugins.json")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) returned error: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) returned error: %v", path, err)
	}

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
		{"this project", twoProjects, thisProject, "0.1.0", true},
		{"another project", twoProjects, "/p/b", "0.2.0", true},
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
