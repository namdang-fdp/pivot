package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/namdang-fdp/pivot/internal/core"
	"github.com/namdang-fdp/pivot/internal/ports"
)

// InitProjectService creates explicit project manifests.
type InitProjectService struct {
	manifests ports.ManifestRepository
	files     ports.Filesystem
}

// NewInitProjectService constructs the explicit manifest creation use case.
func NewInitProjectService(manifests ports.ManifestRepository, files ports.Filesystem) *InitProjectService {
	return &InitProjectService{manifests: manifests, files: files}
}

// InitProjectRequest is the input to explicit manifest creation.
type InitProjectRequest struct {
	Path string
	ID   string
	Name string
}

// InitProjectResult describes a newly created manifest.
type InitProjectResult struct {
	ManifestPath string
	ProjectPath  string
}

// InitProject creates only a minimal strict manifest.
func (s *InitProjectService) InitProject(ctx context.Context, request InitProjectRequest) (InitProjectResult, error) {
	// pivot init ~/Projects/abc OR cd Projects/abc and pivot init
	projectPath := request.Path
	if projectPath == "" {
		var err error
		projectPath, err = s.files.CurrentDirectory()
		if err != nil {
			return InitProjectResult{}, err
		}
	}
	canonical, err := s.files.CanonicalDirectory(projectPath)
	if err != nil {
		return InitProjectResult{}, err
	}
	base := filepath.Base(canonical)
	var id core.ProjectID
	// using folder name if not input id
	// input an ID through pivot init --id "id"
	if request.ID == "" {
		id, err = core.SlugifyProjectID(base)
		if err != nil {
			return InitProjectResult{}, fmt.Errorf(
				"derive project ID from directory name %q: %w; use --id to provide one explicitly",
				base,
				err,
			)
		}
	} else {
		id, err = core.ParseProjectID(request.ID)
		if err != nil {
			return InitProjectResult{}, fmt.Errorf(
				"invalid explicit project ID %q: %w",
				request.ID,
				err,
			)
		}
	}
	name := request.Name
	if name == "" {
		name = base
	}
	if strings.TrimSpace(name) == "" {
		return InitProjectResult{}, fmt.Errorf("project name must not be empty")
	}
	manifest := core.Manifest{
		Version:      1,
		Project:      core.Project{ID: id, Name: name},
		Requirements: core.Requirements{Commands: []string{}, Files: []string{}},
		Ports:        []core.Port{},
	}
	manifestPath, err := s.manifests.Create(ctx, canonical, manifest)
	if err != nil {
		return InitProjectResult{}, err
	}
	return InitProjectResult{ManifestPath: manifestPath, ProjectPath: canonical}, nil
}
