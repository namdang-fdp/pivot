package core

import (
	"fmt"
	"path"
	"strings"
)

// ManifestFilename is the project-root manifest name.
const ManifestFilename = ".pivot.yaml"

// Manifest is Pivot's versioned project declaration.
type Manifest struct {
	Version      int
	Project      Project
	Requirements Requirements
	Compose      Compose
	Ports        []Port
}

// Project contains project identity declared by a manifest.
type Project struct {
	ID   ProjectID
	Name string
}

// Requirements declares host commands and project files needed by a project.
type Requirements struct {
	Commands []string
	Files    []string
}

// Compose declares files consumed by Docker Compose in a future lifecycle slice.
type Compose struct {
	Files    []string
	EnvFiles []string
}

// Port declares a local TCP port required by the project.
type Port struct {
	Name     string
	Port     int
	Protocol string
	Required bool
}

// Validate checks all infrastructure-independent manifest invariants.
func (m Manifest) Validate() error {
	if m.Version == 0 {
		return fmt.Errorf("version is required")
	}
	if m.Version != 1 {
		return fmt.Errorf("version %d is unsupported: only version 1 is supported", m.Version)
	}
	if _, err := ParseProjectID(string(m.Project.ID)); err != nil {
		return err
	}
	if strings.TrimSpace(m.Project.Name) == "" {
		return fmt.Errorf("project.name is required")
	}
	if err := validateUniqueNonEmpty("requirements.commands", m.Requirements.Commands, false); err != nil {
		return err
	}
	if err := validateUniqueNonEmpty("requirements.files", m.Requirements.Files, true); err != nil {
		return err
	}
	if err := validateUniqueNonEmpty("compose.files", m.Compose.Files, true); err != nil {
		return err
	}
	if err := validateUniqueNonEmpty("compose.envFiles", m.Compose.EnvFiles, true); err != nil {
		return err
	}

	portNames := make(map[string]struct{}, len(m.Ports))
	for i, port := range m.Ports {
		field := fmt.Sprintf("ports[%d]", i)
		if strings.TrimSpace(port.Name) == "" {
			return fmt.Errorf("%s.name is required", field)
		}
		if _, exists := portNames[port.Name]; exists {
			return fmt.Errorf("%s.name %q is duplicated", field, port.Name)
		}
		portNames[port.Name] = struct{}{}
		if port.Port < 1 || port.Port > 65535 {
			return fmt.Errorf("%s.port must be between 1 and 65535", field)
		}
		if port.Protocol != "tcp" {
			return fmt.Errorf("%s.protocol %q is unsupported: only tcp is supported", field, port.Protocol)
		}
	}
	return nil
}

func validateUniqueNonEmpty(field string, values []string, paths bool) error {
	seen := make(map[string]struct{}, len(values))
	for i, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s[%d] must not be empty", field, i)
		}
		key := value
		if paths {
			if path.IsAbs(value) {
				return fmt.Errorf("%s[%d] %q must be relative to the project root", field, i, value)
			}
			key = path.Clean(value)
			if key == "." || key == ".." || strings.HasPrefix(key, "../") {
				return fmt.Errorf("%s[%d] %q escapes the project root", field, i, value)
			}
		}
		if _, exists := seen[key]; exists {
			return fmt.Errorf("%s[%d] %q is duplicated", field, i, value)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// ManifestPaths returns field names and project-relative paths in stable order.
func (m Manifest) ManifestPaths() []ManifestPath {
	paths := make([]ManifestPath, 0, len(m.Requirements.Files)+len(m.Compose.Files)+len(m.Compose.EnvFiles))
	for i, value := range m.Requirements.Files {
		paths = append(paths, ManifestPath{Field: fmt.Sprintf("requirements.files[%d]", i), Path: value})
	}
	for i, value := range m.Compose.Files {
		paths = append(paths, ManifestPath{Field: fmt.Sprintf("compose.files[%d]", i), Path: value})
	}
	for i, value := range m.Compose.EnvFiles {
		paths = append(paths, ManifestPath{Field: fmt.Sprintf("compose.envFiles[%d]", i), Path: value})
	}
	return paths
}

// ManifestPath identifies one path-bearing manifest field.
type ManifestPath struct {
	Field string
	Path  string
}
