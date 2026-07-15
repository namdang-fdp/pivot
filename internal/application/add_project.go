package application

import (
	"context"
	"fmt"

	"github.com/namdang-fdp/pivot/internal/core"
	"github.com/namdang-fdp/pivot/internal/ports"
)

// AddProjectService registers explicit project roots.
type AddProjectService struct {
	manifests ports.ManifestRepository
	registry  ports.ProjectRegistry
	files     ports.Filesystem
}

// NewAddProjectService constructs the explicit project registration use case.
func NewAddProjectService(manifests ports.ManifestRepository, registry ports.ProjectRegistry, files ports.Filesystem) *AddProjectService {
	return &AddProjectService{manifests: manifests, registry: registry, files: files}
}

// AddProjectResult describes an add or idempotent registration.
type AddProjectResult struct {
	Project           core.RegisteredProject
	AlreadyRegistered bool
}

// AddProject validates and registers one explicit project root.
func (s *AddProjectService) AddProject(ctx context.Context, path string) (AddProjectResult, error) {
	if path == "" {
		var err error
		path, err = s.files.CurrentDirectory()
		if err != nil {
			return AddProjectResult{}, err
		}
	}
	canonical, err := s.files.CanonicalDirectory(path)
	if err != nil {
		return AddProjectResult{}, err
	}
	manifest, err := s.manifests.Load(ctx, canonical)
	if err != nil {
		return AddProjectResult{}, err
	}
	project := core.RegisteredProject{ID: manifest.Project.ID, Name: manifest.Project.Name, Path: canonical}
	result := AddProjectResult{Project: project}
	err = s.registry.Update(ctx, func(projects []core.RegisteredProject) ([]core.RegisteredProject, error) {
		for _, existing := range projects {
			if existing.ID == project.ID && existing.Path == project.Path {
				result.AlreadyRegistered = true
				return projects, nil
			}
			if existing.ID == project.ID {
				return nil, fmt.Errorf("project ID %q is already registered at %q; registry was not changed", project.ID, existing.Path)
			}
			if existing.Path == project.Path {
				return nil, fmt.Errorf("project path %q is already registered as %q; registry was not changed", project.Path, existing.ID)
			}
		}
		return append(projects, project), nil
	})
	if err != nil {
		return AddProjectResult{}, err
	}
	return result, nil
}
