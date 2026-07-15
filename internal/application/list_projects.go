package application

import (
	"context"
	"path/filepath"
	"sort"

	"github.com/namdang-fdp/pivot/internal/core"
	"github.com/namdang-fdp/pivot/internal/ports"
)

// ListProjectsService lists registered project roots.
type ListProjectsService struct {
	registry ports.ProjectRegistry
	files    ports.Filesystem
}

// NewListProjectsService constructs the registered project listing use case.
func NewListProjectsService(registry ports.ProjectRegistry, files ports.Filesystem) *ListProjectsService {
	return &ListProjectsService{registry: registry, files: files}
}

// ListedProject is a registry entry enriched with local availability.
type ListedProject struct {
	ID        core.ProjectID
	Name      string
	Path      string
	Manifest  string
	Available bool
}

// ListProjects returns registered projects in stable ID order.
func (s *ListProjectsService) ListProjects(ctx context.Context) ([]ListedProject, error) {
	projects, err := s.registry.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ListedProject, 0, len(projects))
	for _, project := range projects {
		manifest := filepath.Join(project.Path, core.ManifestFilename)
		result = append(result, ListedProject{
			ID: project.ID, Name: project.Name, Path: project.Path, Manifest: manifest,
			Available: s.files.DirectoryExists(project.Path) && s.files.FileExists(manifest),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}
