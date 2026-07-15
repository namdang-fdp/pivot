package application

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/namdang-fdp/pivot/internal/core"
	"github.com/namdang-fdp/pivot/internal/ports"
)

// DoctorProjectService performs read-only project diagnostics.
type DoctorProjectService struct {
	manifests ports.ManifestRepository
	registry  ports.ProjectRegistry
	files     ports.Filesystem
	commands  ports.CommandInspector
	ports     ports.PortInspector
}

// NewDoctorProjectService constructs the read-only project diagnostics use case.
func NewDoctorProjectService(
	manifests ports.ManifestRepository,
	registry ports.ProjectRegistry,
	files ports.Filesystem,
	commands ports.CommandInspector,
	portInspector ports.PortInspector,
) *DoctorProjectService {
	return &DoctorProjectService{manifests: manifests, registry: registry, files: files, commands: commands, ports: portInspector}
}

// DoctorProjectRequest identifies a registered project or requests current-directory resolution.
type DoctorProjectRequest struct {
	Project string
}

// DoctorProjectResult is the structured, presentation-independent doctor result.
type DoctorProjectResult struct {
	Project core.RegisteredProject
	Status  core.DiagnosticStatus
	Summary core.DiagnosticSummary
	Checks  []core.Diagnostic
}

// DoctorFailedError indicates that a complete report contains failed checks.
type DoctorFailedError struct{ Failed int }

func (e *DoctorFailedError) Error() string {
	return fmt.Sprintf("doctor found %d failed check(s)", e.Failed)
}

// DoctorProject performs read-only manifest, registry, requirement, Compose, and port checks.
func (s *DoctorProjectService) DoctorProject(ctx context.Context, request DoctorProjectRequest) (DoctorProjectResult, error) {
	registered, err := s.registry.List(ctx)
	if err != nil {
		return DoctorProjectResult{}, err
	}
	project, isRegistered, err := s.resolveDoctorProject(ctx, request.Project, registered)
	if err != nil {
		return DoctorProjectResult{}, err
	}

	checks := make([]core.Diagnostic, 0)
	manifest, manifestErr := s.manifests.Load(ctx, project.Path)
	if manifestErr != nil {
		checks = append(checks, diagnostic("manifest", "strict-validation", core.DiagnosticFail, manifestErr.Error()))
	} else {
		project.Name = manifest.Project.Name
		if !isRegistered {
			project.ID = manifest.Project.ID
		}
		checks = append(checks,
			diagnostic("manifest", "schema-version", core.DiagnosticPass, "Manifest schema version 1 is supported"),
			diagnostic("manifest", "project-identity", core.DiagnosticPass, fmt.Sprintf("Project %s is named %q", manifest.Project.ID, manifest.Project.Name)),
			diagnostic("manifest", "path-safety", core.DiagnosticPass, "All manifest paths remain within the project root"),
			diagnostic("manifest", "compose-declarations", core.DiagnosticPass, fmt.Sprintf("Validated %d Compose file(s) and %d environment file(s)", len(manifest.Compose.Files), len(manifest.Compose.EnvFiles))),
			diagnostic("manifest", "port-declarations", core.DiagnosticPass, fmt.Sprintf("Validated %d TCP port declaration(s)", len(manifest.Ports))),
		)
	}

	manifestPath := filepath.Join(project.Path, core.ManifestFilename)
	if s.files.DirectoryExists(project.Path) {
		checks = append(checks, diagnostic("registry", "project-path", core.DiagnosticPass, fmt.Sprintf("Project path exists: %s", project.Path)))
	} else {
		checks = append(checks, diagnostic("registry", "project-path", core.DiagnosticFail, fmt.Sprintf("Registered project path does not exist: %s", project.Path)))
	}
	if s.files.FileExists(manifestPath) {
		checks = append(checks, diagnostic("registry", "manifest-path", core.DiagnosticPass, fmt.Sprintf("Manifest exists: %s", manifestPath)))
	} else {
		checks = append(checks, diagnostic("registry", "manifest-path", core.DiagnosticFail, fmt.Sprintf("Manifest does not exist: %s", manifestPath)))
	}
	if isRegistered {
		if manifestErr == nil && project.ID != manifest.Project.ID {
			checks = append(checks, diagnostic("registry", "project-id", core.DiagnosticFail, fmt.Sprintf("Registry ID %q does not match manifest ID %q", project.ID, manifest.Project.ID)))
		} else if manifestErr == nil {
			checks = append(checks, diagnostic("registry", "project-id", core.DiagnosticPass, "Registry ID matches the manifest ID"))
		}
	} else {
		checks = append(checks, diagnostic("registry", "registration", core.DiagnosticWarning, "Project is not registered; doctor is inspecting the manifest directly"))
	}
	checks = append(checks, s.registryConflictCheck(project, manifest, manifestErr, registered)...)

	if manifestErr == nil {
		checks = append(checks, s.requirementChecks(manifest, project.Path)...)
		checks = append(checks, s.composeChecks(ctx, manifest, project.Path)...)
		checks = append(checks, s.portChecks(ctx, manifest)...)
	}

	summary := core.SummarizeDiagnostics(checks)
	status := core.DiagnosticPass
	if summary.Failed > 0 {
		status = core.DiagnosticFail
	} else if summary.Warnings > 0 {
		status = core.DiagnosticWarning
	}
	result := DoctorProjectResult{Project: project, Status: status, Summary: summary, Checks: checks}
	if summary.Failed > 0 {
		return result, &DoctorFailedError{Failed: summary.Failed}
	}
	return result, nil
}

func (s *DoctorProjectService) resolveDoctorProject(ctx context.Context, requested string, registered []core.RegisteredProject) (core.RegisteredProject, bool, error) {
	if requested != "" {
		id, err := core.ParseProjectID(requested)
		if err != nil {
			return core.RegisteredProject{}, false, err
		}
		for _, project := range registered {
			if project.ID == id {
				return project, true, nil
			}
		}
		return core.RegisteredProject{}, false, fmt.Errorf("project %q is not registered; run `pivot list` or `pivot add <path>`", requested)
	}
	cwd, err := s.files.CurrentDirectory()
	if err != nil {
		return core.RegisteredProject{}, false, err
	}
	canonical, err := s.files.CanonicalDirectory(cwd)
	if err != nil {
		return core.RegisteredProject{}, false, err
	}
	var match *core.RegisteredProject
	for i := range registered {
		if containsPath(registered[i].Path, canonical) && (match == nil || len(registered[i].Path) > len(match.Path)) {
			candidate := registered[i]
			match = &candidate
		}
	}
	if match != nil {
		return *match, true, nil
	}
	root, found, err := s.files.FindManifestRoot(canonical)
	if err != nil {
		return core.RegisteredProject{}, false, err
	}
	if !found {
		return core.RegisteredProject{}, false, errors.New("current directory is not inside a registered project and no .pivot.yaml was found; run `pivot add <path>` or specify a project ID")
	}
	manifest, err := s.manifests.Load(ctx, root)
	if err != nil {
		return core.RegisteredProject{Path: root}, false, nil
	}
	return core.RegisteredProject{ID: manifest.Project.ID, Name: manifest.Project.Name, Path: root}, false, nil
}

func containsPath(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
