package adapters

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/namdang-fdp/pivot/internal/core"
)

func TestYAMLRegistryEmpty(t *testing.T) {
	t.Parallel()
	projects, err := NewYAMLProjectRegistryAt(filepath.Join(t.TempDir(), "pivot", "projects.yaml")).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("projects = %#v, want empty", projects)
	}
}

func TestYAMLRegistryUpdateAndDeterministicOrdering(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "pivot", "projects.yaml")
	registry := NewYAMLProjectRegistryAt(path)
	projects := []core.RegisteredProject{
		{ID: "smartpark", Name: "SmartPark", Path: filepath.Join(t.TempDir(), "smartpark")},
		{ID: "centeros", Name: "CenterOS", Path: filepath.Join(t.TempDir(), "centeros")},
	}
	if err := registry.Update(context.Background(), func([]core.RegisteredProject) ([]core.RegisteredProject, error) { return projects, nil }); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, err := registry.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 || got[0].ID != "centeros" || got[1].ID != "smartpark" {
		t.Fatalf("projects = %#v", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry: %v", err)
	}
	if strings.Index(string(data), "centeros:") > strings.Index(string(data), "smartpark:") {
		t.Fatalf("registry is not deterministically ordered:\n%s", data)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat registry: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("registry mode = %v, want 0600", info.Mode().Perm())
	}
	directoryInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat registry directory: %v", err)
	}
	if directoryInfo.Mode().Perm() != 0o700 {
		t.Fatalf("registry directory mode = %v, want 0700", directoryInfo.Mode().Perm())
	}
	lockInfo, err := os.Stat(path + ".lock")
	if err != nil {
		t.Fatalf("stat registry lock: %v", err)
	}
	if lockInfo.Mode().Perm() != 0o600 {
		t.Fatalf("registry lock mode = %v, want 0600", lockInfo.Mode().Perm())
	}
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".projects.yaml.tmp-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("temporary registry files = %v, err = %v", matches, err)
	}
}

func TestYAMLRegistryRejectsUnsupportedAndCorruptData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		message string
	}{
		{name: "unsupported", content: "version: 2\nprojects: {}\n", message: "unsupported"},
		{name: "unknown", content: "version: 1\nprojects: {}\nunknown: true\n", message: "field unknown"},
		{name: "corrupt", content: "version: 1\nprojects:\n  bad:\n    name: Bad\n    path: relative\n", message: "clean absolute path"},
		{name: "duplicate key", content: "version: 1\nversion: 1\nprojects: {}\n", message: "duplicate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "projects.yaml")
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("write registry: %v", err)
			}
			_, err := NewYAMLProjectRegistryAt(path).List(context.Background())
			if err == nil || !strings.Contains(err.Error(), tt.message) {
				t.Fatalf("List error = %v, want containing %q", err, tt.message)
			}
		})
	}
}

func TestYAMLRegistryRespectsXDGConfigHome(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	registry := NewYAMLProjectRegistry()
	projectPath := filepath.Join(t.TempDir(), "project")
	if err := registry.Update(context.Background(), func([]core.RegisteredProject) ([]core.RegisteredProject, error) {
		return []core.RegisteredProject{{ID: "project", Name: "Project", Path: projectPath}}, nil
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(xdg, "pivot", "projects.yaml")); err != nil {
		t.Fatalf("XDG registry: %v", err)
	}
}

func TestYAMLRegistryConcurrentUpdates(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "pivot", "projects.yaml")
	const count = 16
	var wait sync.WaitGroup
	errorsChannel := make(chan error, count)
	for i := range count {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			registry := NewYAMLProjectRegistryAt(path)
			id := core.ProjectID(fmt.Sprintf("project-%d", index))
			err := registry.Update(context.Background(), func(projects []core.RegisteredProject) ([]core.RegisteredProject, error) {
				return append(projects, core.RegisteredProject{ID: id, Name: string(id), Path: filepath.Join(filepath.Dir(path), "roots", string(id))}), nil
			})
			errorsChannel <- err
		}(i)
	}
	wait.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent Update: %v", err)
		}
	}
	projects, err := NewYAMLProjectRegistryAt(path).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(projects) != count {
		t.Fatalf("project count = %d, want %d", len(projects), count)
	}
}
