package adapters

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/namdang-fdp/pivot/internal/core"
	"gopkg.in/yaml.v3"
)

// YAMLManifestRepository stores strict project manifests in project roots.
type YAMLManifestRepository struct{}

// NewYAMLManifestRepository creates a strict YAML manifest adapter.
func NewYAMLManifestRepository() *YAMLManifestRepository { return &YAMLManifestRepository{} }

type manifestDTO struct {
	Version      int              `yaml:"version"`
	Project      *projectDTO      `yaml:"project"`
	Requirements *requirementsDTO `yaml:"requirements"`
	Compose      *composeDTO      `yaml:"compose,omitempty"`
	Ports        *[]portDTO       `yaml:"ports"`
}

type projectDTO struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type requirementsDTO struct {
	Commands *[]string `yaml:"commands"`
	Files    *[]string `yaml:"files"`
}

type composeDTO struct {
	Files    []string `yaml:"files"`
	EnvFiles []string `yaml:"envFiles"`
}

type portDTO struct {
	Name     string `yaml:"name"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"`
	Required *bool  `yaml:"required"`
}

// Load strictly decodes and validates a project's manifest.
func (r *YAMLManifestRepository) Load(_ context.Context, root string) (core.Manifest, error) {
	canonicalRoot, err := NewHostFilesystem().CanonicalDirectory(root)
	if err != nil {
		return core.Manifest{}, err
	}
	manifestPath := filepath.Join(canonicalRoot, core.ManifestFilename)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return core.Manifest{}, fmt.Errorf("read manifest %q: %w", manifestPath, err)
	}
	manifest, err := decodeManifest(data, canonicalRoot)
	if err != nil {
		return core.Manifest{}, fmt.Errorf("validate manifest %q: %w", manifestPath, err)
	}
	return manifest, nil
}

// Create atomically creates a validated manifest without overwriting an existing one.
func (r *YAMLManifestRepository) Create(_ context.Context, root string, manifest core.Manifest) (string, error) {
	canonicalRoot, err := NewHostFilesystem().CanonicalDirectory(root)
	if err != nil {
		return "", err
	}
	destination := filepath.Join(canonicalRoot, core.ManifestFilename)
	if _, err := os.Lstat(destination); err == nil {
		return "", fmt.Errorf("refusing to overwrite existing manifest %q", destination)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect manifest path %q: %w", destination, err)
	}

	data, err := encodeManifest(manifest)
	if err != nil {
		return "", err
	}
	if _, err := decodeManifest(data, canonicalRoot); err != nil {
		return "", fmt.Errorf("generated manifest failed validation: %w", err)
	}

	temp, err := os.CreateTemp(canonicalRoot, ".pivot.yaml.tmp-*")
	if err != nil {
		return "", fmt.Errorf("create temporary manifest: %w", err)
	}
	tempName := temp.Name()
	cleanup := func() {
		if tempName != "" {
			_ = os.Remove(tempName)
		}
	}
	defer cleanup()
	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return "", fmt.Errorf("set manifest permissions: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return "", fmt.Errorf("write temporary manifest: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return "", fmt.Errorf("sync temporary manifest: %w", err)
	}
	if err := temp.Close(); err != nil {
		return "", fmt.Errorf("close temporary manifest: %w", err)
	}
	if err := os.Link(tempName, destination); err != nil {
		if errors.Is(err, os.ErrExist) {
			return "", fmt.Errorf("refusing to overwrite existing manifest %q", destination)
		}
		return "", fmt.Errorf("publish manifest %q atomically: %w", destination, err)
	}
	if err := os.Remove(tempName); err != nil {
		return "", fmt.Errorf("remove temporary manifest %q: %w", tempName, err)
	}
	tempName = ""
	if err := syncDirectory(canonicalRoot); err != nil {
		return "", err
	}
	return destination, nil
}

func decodeManifest(data []byte, root string) (core.Manifest, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return core.Manifest{}, fmt.Errorf("decode YAML: %w", err)
	}
	if err := rejectDuplicateYAMLKeys(&document, "manifest"); err != nil {
		return core.Manifest{}, err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var dto manifestDTO
	if err := decoder.Decode(&dto); err != nil {
		return core.Manifest{}, fmt.Errorf("decode strict YAML: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return core.Manifest{}, fmt.Errorf("multiple YAML documents are not supported")
		}
		return core.Manifest{}, fmt.Errorf("decode YAML: %w", err)
	}
	if dto.Project == nil {
		return core.Manifest{}, fmt.Errorf("project is required")
	}
	if dto.Requirements == nil {
		return core.Manifest{}, fmt.Errorf("requirements is required")
	}
	if dto.Ports == nil {
		return core.Manifest{}, fmt.Errorf("ports is required")
	}
	if dto.Requirements.Commands == nil {
		return core.Manifest{}, fmt.Errorf("requirements.commands is required")
	}
	if dto.Requirements.Files == nil {
		return core.Manifest{}, fmt.Errorf("requirements.files is required")
	}

	id, err := core.ParseProjectID(dto.Project.ID)
	if err != nil {
		return core.Manifest{}, err
	}
	manifest := core.Manifest{
		Version: dto.Version,
		Project: core.Project{ID: id, Name: dto.Project.Name},
		Requirements: core.Requirements{
			Commands: append([]string(nil), (*dto.Requirements.Commands)...),
			Files:    append([]string(nil), (*dto.Requirements.Files)...),
		},
		Ports: make([]core.Port, 0, len(*dto.Ports)),
	}
	if dto.Compose != nil {
		manifest.Compose.Files = append([]string(nil), dto.Compose.Files...)
		manifest.Compose.EnvFiles = append([]string(nil), dto.Compose.EnvFiles...)
	}
	for i, value := range *dto.Ports {
		if value.Required == nil {
			return core.Manifest{}, fmt.Errorf("ports[%d].required is required", i)
		}
		manifest.Ports = append(manifest.Ports, core.Port{
			Name: value.Name, Port: value.Port, Protocol: value.Protocol, Required: *value.Required,
		})
	}
	if err := manifest.Validate(); err != nil {
		return core.Manifest{}, err
	}
	for _, value := range manifest.ManifestPaths() {
		if err := validatePathWithinRoot(root, value.Field, value.Path); err != nil {
			return core.Manifest{}, err
		}
	}
	return manifest, nil
}

func encodeManifest(manifest core.Manifest) ([]byte, error) {
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	ports := make([]portDTO, 0, len(manifest.Ports))
	for _, value := range manifest.Ports {
		required := value.Required
		ports = append(ports, portDTO{Name: value.Name, Port: value.Port, Protocol: value.Protocol, Required: &required})
	}
	commands := append([]string(nil), manifest.Requirements.Commands...)
	files := append([]string(nil), manifest.Requirements.Files...)
	dto := manifestDTO{
		Version: manifest.Version,
		Project: &projectDTO{ID: string(manifest.Project.ID), Name: manifest.Project.Name},
		Requirements: &requirementsDTO{
			Commands: &commands,
			Files:    &files,
		},
		Ports: &ports,
	}
	if len(manifest.Compose.Files) > 0 || len(manifest.Compose.EnvFiles) > 0 {
		dto.Compose = &composeDTO{Files: manifest.Compose.Files, EnvFiles: manifest.Compose.EnvFiles}
	}
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(dto); err != nil {
		return nil, fmt.Errorf("encode manifest: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("finish manifest encoding: %w", err)
	}
	return output.Bytes(), nil
}

func rejectDuplicateYAMLKeys(node *yaml.Node, location string) error {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.MappingNode {
		seen := make(map[string]struct{}, len(node.Content)/2)
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if _, exists := seen[key.Value]; exists {
				return fmt.Errorf("duplicate YAML key %q at %s (line %d)", key.Value, location, key.Line)
			}
			seen[key.Value] = struct{}{}
			if err := rejectDuplicateYAMLKeys(node.Content[i+1], location+"."+key.Value); err != nil {
				return err
			}
		}
		return nil
	}
	for _, child := range node.Content {
		if err := rejectDuplicateYAMLKeys(child, location); err != nil {
			return err
		}
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open directory %q for sync: %w", path, err)
	}
	defer func() { _ = directory.Close() }()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync directory %q: %w", path, err)
	}
	return nil
}
