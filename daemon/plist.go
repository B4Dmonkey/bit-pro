package daemon

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const (
	agentDirMode  = 0o755
	plistFileMode = 0o644
)

var plistTemplate = template.Must(template.New("plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{.Label}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.Exe}}</string>
		<string>serve</string>
		<string>daemon</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<dict>
		<key>SuccessfulExit</key>
		<false/>
	</dict>
	<key>StandardOutPath</key>
	<string>{{.LogPath}}</string>
	<key>StandardErrorPath</key>
	<string>{{.LogPath}}</string>
</dict>
</plist>
`))

func PlistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locating the home directory: %w", err)
	}

	dir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(dir, agentDirMode); err != nil {
		return "", fmt.Errorf("creating %s: %w", dir, err)
	}

	return filepath.Join(dir, Label+".plist"), nil
}

func Plist(exe, logPath string) []byte {
	fields := struct {
		Label   string
		Exe     string
		LogPath string
	}{Label: Label, Exe: exe, LogPath: logPath}

	var rendered bytes.Buffer

	_ = plistTemplate.Execute(&rendered, fields)

	return rendered.Bytes()
}

func WritePlist(path string, data []byte) error {
	if err := os.WriteFile(path, data, plistFileMode); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}
