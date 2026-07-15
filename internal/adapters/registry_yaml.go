package adapters

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/namdang-fdp/pivot/internal/core"
	"gopkg.in/yaml.v3"
)

const registryVersion = 1

// YAMLProjectRegistry stores the machine registry as strict, locked YAML.
type YAMLProjectRegistry struct {
	explicitPath string
}

// NewYAMLProjectRegistry creates an XDG-aware project registry adapter.
func NewYAMLProjectRegistry() *YAMLProjectRegistry { return &YAMLProjectRegistry{} }

// NewYAMLProjectRegistryAt creates a registry at an explicit path for tests and embedding.
func NewYAMLProjectRegistryAt(path string) *YAMLProjectRegistry {
	return &YAMLProjectRegistry{explicitPath: path}
}

type registryDTO struct {
	Version  int                          `yaml:"version"`
	Projects *map[string]registryEntryDTO `yaml:"projects"`
}

type registryEntryDTO struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// List strictly reads registered projects without creating registry files.
func (r *YAMLProjectRegistry) List(_ context.Context) ([]core.RegisteredProject, error) {
	path, err := r.path()
	if err != nil {
		return nil, err
	}
	return r.loadUnlocked(path)
}

// Update runs a registry transformation under an inter-process lock and writes atomically.
func (r *YAMLProjectRegistry) Update(_ context.Context, update func([]core.RegisteredProject) ([]core.RegisteredProject, error)) error {
	path, err := r.path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create Pivot config directory %q: %w", filepath.Dir(path), err)
	}
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("secure Pivot config directory %q: %w", filepath.Dir(path), err)
	}
	lock, err := acquireFileLock(path + ".lock")
	if err != nil {
		return err
	}
	defer func() { _ = releaseFileLock(lock) }()

	projects, err := r.loadUnlocked(path)
	if err != nil {
		return err
	}
	updated, err := update(append([]core.RegisteredProject(nil), projects...))
	if err != nil {
		return err
	}
	if err := validateRegisteredProjects(updated); err != nil {
		return fmt.Errorf("refuse invalid registry update: %w", err)
	}
	return writeRegistryAtomic(path, updated)
}

func (r *YAMLProjectRegistry) path() (string, error) {
	if r.explicitPath != "" {
		abs, err := filepath.Abs(r.explicitPath)
		if err != nil {
			return "", fmt.Errorf("resolve registry path: %w", err)
		}
		return filepath.Clean(abs), nil
	}
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		if !filepath.IsAbs(configHome) {
			return "", fmt.Errorf("XDG_CONFIG_HOME must be an absolute path, got %q", configHome)
		}
		return filepath.Join(filepath.Clean(configHome), "pivot", "projects.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory for Pivot registry: %w", err)
	}
	return filepath.Join(home, ".config", "pivot", "projects.yaml"), nil
}

func (r *YAMLProjectRegistry) loadUnlocked(path string) ([]core.RegisteredProject, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []core.RegisteredProject{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read registry %q: %w", path, err)
	}

	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("decode registry %q: %w", path, err)
	}
	if err := rejectDuplicateYAMLKeys(&document, "registry"); err != nil {
		return nil, fmt.Errorf("decode registry %q: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var dto registryDTO
	if err := decoder.Decode(&dto); err != nil {
		return nil, fmt.Errorf("decode strict registry %q: %w", path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, fmt.Errorf("decode registry %q: multiple YAML documents are not supported", path)
		}
		return nil, fmt.Errorf("decode registry %q: %w", path, err)
	}
	if dto.Version == 0 {
		return nil, fmt.Errorf("registry %q: version is required", path)
	}
	if dto.Version != registryVersion {
		return nil, fmt.Errorf("registry %q: version %d is unsupported: only version 1 is supported", path, dto.Version)
	}
	if dto.Projects == nil {
		return nil, fmt.Errorf("registry %q: projects is required", path)
	}

	projects := make([]core.RegisteredProject, 0, len(*dto.Projects))
	for rawID, entry := range *dto.Projects {
		id, err := core.ParseProjectID(rawID)
		if err != nil {
			return nil, fmt.Errorf("registry %q: projects.%s: %w", path, rawID, err)
		}
		projects = append(projects, core.RegisteredProject{ID: id, Name: entry.Name, Path: entry.Path})
	}
	if err := validateRegisteredProjects(projects); err != nil {
		return nil, fmt.Errorf("registry %q is corrupt: %w", path, err)
	}
	sortProjects(projects)
	return projects, nil
}

func validateRegisteredProjects(projects []core.RegisteredProject) error {
	ids := make(map[core.ProjectID]struct{}, len(projects))
	paths := make(map[string]core.ProjectID, len(projects))
	for _, project := range projects {
		if _, err := core.ParseProjectID(string(project.ID)); err != nil {
			return err
		}
		if strings.TrimSpace(project.Name) == "" {
			return fmt.Errorf("projects.%s.name is required", project.ID)
		}
		if !filepath.IsAbs(project.Path) || filepath.Clean(project.Path) != project.Path {
			return fmt.Errorf("projects.%s.path %q must be a clean absolute path", project.ID, project.Path)
		}
		if _, exists := ids[project.ID]; exists {
			return fmt.Errorf("project ID %q is duplicated", project.ID)
		}
		ids[project.ID] = struct{}{}
		if owner, exists := paths[project.Path]; exists {
			return fmt.Errorf("project path %q is assigned to both %q and %q", project.Path, owner, project.ID)
		}
		paths[project.Path] = project.ID
	}
	return nil
}

func writeRegistryAtomic(path string, projects []core.RegisteredProject) error {
	sortProjects(projects)
	entries := make(map[string]registryEntryDTO, len(projects))
	for _, project := range projects {
		entries[string(project.ID)] = registryEntryDTO{Name: project.Name, Path: project.Path}
	}
	dto := registryDTO{Version: registryVersion, Projects: &entries}
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(dto); err != nil {
		return fmt.Errorf("encode registry: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("finish registry encoding: %w", err)
	}

	directory := filepath.Dir(path)
	temp, err := os.CreateTemp(directory, ".projects.yaml.tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary registry: %w", err)
	}
	tempName := temp.Name()
	defer func() { _ = os.Remove(tempName) }()
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("set registry permissions: %w", err)
	}
	if _, err := temp.Write(output.Bytes()); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write temporary registry: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync temporary registry: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary registry: %w", err)
	}
	if err := os.Rename(tempName, path); err != nil {
		return fmt.Errorf("replace registry %q atomically: %w", path, err)
	}
	return syncDirectory(directory)
}

func acquireFileLock(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open registry lock %q: %w", path, err)
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("secure registry lock %q: %w", path, err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("lock registry %q: %w", path, err)
	}
	return file, nil
}

func releaseFileLock(file *os.File) error {
	unlockErr := syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	closeErr := file.Close()
	if unlockErr != nil {
		return unlockErr
	}
	return closeErr
}

func sortProjects(projects []core.RegisteredProject) {
	sort.Slice(projects, func(i, j int) bool { return projects[i].ID < projects[j].ID })
}
