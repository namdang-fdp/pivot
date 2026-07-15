package adapters

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/namdang-fdp/pivot/internal/core"
)

const validManifestYAML = `version: 1
project:
  id: centeros
  name: CenterOS
requirements:
  commands:
    - docker
  files:
    - .env
compose:
  files:
    - ops/compose.yml
  envFiles:
    - .env
ports:
  - name: postgres
    port: 5432
    protocol: tcp
    required: true
`

func TestYAMLManifestLoadValid(t *testing.T) {
	t.Parallel()
	root := writeManifestFixture(t, validManifestYAML)
	manifest, err := NewYAMLManifestRepository().Load(context.Background(), root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if manifest.Project.ID != "centeros" || manifest.Project.Name != "CenterOS" {
		t.Fatalf("unexpected project: %#v", manifest.Project)
	}
}

func TestYAMLManifestRejectsInvalidInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yaml    string
		message string
	}{
		{name: "unknown field", yaml: strings.Replace(validManifestYAML, "  name: CenterOS", "  name: CenterOS\n  unknown: true", 1), message: "field unknown not found"},
		{name: "duplicate key", yaml: strings.Replace(validManifestYAML, "version: 1", "version: 1\nversion: 1", 1), message: "duplicate"},
		{name: "unsupported version", yaml: strings.Replace(validManifestYAML, "version: 1", "version: 2", 1), message: "unsupported"},
		{name: "missing project", yaml: "version: 1\nrequirements:\n  commands: []\n  files: []\nports: []\n", message: "project is required"},
		{name: "missing name", yaml: strings.Replace(validManifestYAML, "  name: CenterOS", "  name: ''", 1), message: "project.name is required"},
		{name: "duplicate commands", yaml: strings.Replace(validManifestYAML, "    - docker", "    - docker\n    - docker", 1), message: "requirements.commands[1]"},
		{name: "duplicate files", yaml: strings.Replace(validManifestYAML, "  files:\n    - .env", "  files:\n    - .env\n    - ./.env", 1), message: "requirements.files[1]"},
		{name: "duplicate ports", yaml: strings.Replace(validManifestYAML, "  - name: postgres", "  - name: postgres\n    port: 1234\n    protocol: tcp\n    required: true\n  - name: postgres", 1), message: "ports[1].name"},
		{name: "invalid port", yaml: strings.Replace(validManifestYAML, "port: 5432", "port: 70000", 1), message: "between 1 and 65535"},
		{name: "unsupported protocol", yaml: strings.Replace(validManifestYAML, "protocol: tcp", "protocol: udp", 1), message: "only tcp"},
		{name: "absolute path", yaml: strings.Replace(validManifestYAML, "    - .env", "    - /tmp/.env", 1), message: "must be relative"},
		{name: "lexical traversal", yaml: strings.Replace(validManifestYAML, "    - .env", "    - ../.env", 1), message: "escapes the project root"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := writeManifestFixture(t, tt.yaml)
			_, err := NewYAMLManifestRepository().Load(context.Background(), root)
			if err == nil || !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("Load error = %v, want containing %q", err, tt.message)
			}
		})
	}
}

func TestYAMLManifestRejectsSymlinkEscape(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	yaml := strings.Replace(validManifestYAML, "    - .env", "    - escape/secret.env", 1)
	if err := os.WriteFile(filepath.Join(root, core.ManifestFilename), []byte(yaml), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	_, err := NewYAMLManifestRepository().Load(context.Background(), root)
	if err == nil || !strings.Contains(err.Error(), "resolves outside") {
		t.Fatalf("Load error = %v, want symlink escape", err)
	}
}

func TestYAMLManifestCreateRoundTrip(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	repository := NewYAMLManifestRepository()
	want := core.Manifest{
		Version:      1,
		Project:      core.Project{ID: "sample-project", Name: "Sample Project"},
		Requirements: core.Requirements{Commands: []string{}, Files: []string{}},
		Ports:        []core.Port{},
	}
	path, err := repository.Create(context.Background(), root, want)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if path != filepath.Join(root, core.ManifestFilename) {
		t.Fatalf("path = %q", path)
	}
	got, err := repository.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("Load generated manifest: %v", err)
	}
	if got.Project != want.Project || got.Version != 1 || len(got.Ports) != 0 {
		t.Fatalf("round trip = %#v, want %#v", got, want)
	}
	if _, err := repository.Create(context.Background(), root, want); err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("second Create error = %v", err)
	}
}

func writeManifestFixture(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ops"), 0o755); err != nil {
		t.Fatalf("mkdir ops: %v", err)
	}
	for _, name := range []string{".env", "ops/compose.yml"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("fixture\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, core.ManifestFilename), []byte(content), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return root
}
