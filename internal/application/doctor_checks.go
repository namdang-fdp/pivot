package application

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/namdang-fdp/pivot/internal/core"
)

func (s *DoctorProjectService) registryConflictCheck(project core.RegisteredProject, manifest core.Manifest, manifestErr error, registered []core.RegisteredProject) []core.Diagnostic {
	id := project.ID
	if manifestErr == nil {
		id = manifest.Project.ID
	}
	for _, existing := range registered {
		if existing.ID == id && existing.Path != project.Path {
			return []core.Diagnostic{diagnostic("registry", "conflicts", core.DiagnosticFail, fmt.Sprintf("Project ID %q is registered at a different path: %s", id, existing.Path))}
		}
		if existing.Path == project.Path && existing.ID != id {
			return []core.Diagnostic{diagnostic("registry", "conflicts", core.DiagnosticFail, fmt.Sprintf("Project path is registered with different ID %q", existing.ID))}
		}
	}
	return []core.Diagnostic{diagnostic("registry", "conflicts", core.DiagnosticPass, "Registry contains no duplicate ID or path conflict for this project")}
}

func (s *DoctorProjectService) requirementChecks(manifest core.Manifest, root string) []core.Diagnostic {
	checks := make([]core.Diagnostic, 0, len(manifest.Requirements.Commands)+len(manifest.Requirements.Files))
	for _, command := range manifest.Requirements.Commands {
		observation, err := s.commands.Find(command)
		if err != nil {
			checks = append(checks, diagnostic("requirements", "command:"+command, core.DiagnosticFail, err.Error()))
		} else {
			checks = append(checks, diagnostic("requirements", "command:"+command, core.DiagnosticPass, fmt.Sprintf("Command %q is available at %s", command, observation.Path)))
		}
	}
	for _, file := range manifest.Requirements.Files {
		fullPath := filepath.Join(root, filepath.FromSlash(file))
		if s.files.FileExists(fullPath) {
			checks = append(checks, diagnostic("requirements", "file:"+file, core.DiagnosticPass, fmt.Sprintf("Required file exists: %s", file)))
		} else {
			checks = append(checks, diagnostic("requirements", "file:"+file, core.DiagnosticFail, fmt.Sprintf("Required file is missing: %s", file)))
		}
	}
	return checks
}

func (s *DoctorProjectService) composeChecks(ctx context.Context, manifest core.Manifest, root string) []core.Diagnostic {
	checks := make([]core.Diagnostic, 0, len(manifest.Compose.Files)+len(manifest.Compose.EnvFiles)+1)
	for _, file := range manifest.Compose.Files {
		if s.files.FileExists(filepath.Join(root, filepath.FromSlash(file))) {
			checks = append(checks, diagnostic("compose", "file:"+file, core.DiagnosticPass, fmt.Sprintf("Compose file exists: %s", file)))
		} else {
			checks = append(checks, diagnostic("compose", "file:"+file, core.DiagnosticFail, fmt.Sprintf("Compose file is missing: %s", file)))
		}
	}
	for _, file := range manifest.Compose.EnvFiles {
		if s.files.FileExists(filepath.Join(root, filepath.FromSlash(file))) {
			checks = append(checks, diagnostic("compose", "env-file:"+file, core.DiagnosticPass, fmt.Sprintf("Compose environment file exists: %s", file)))
		} else {
			checks = append(checks, diagnostic("compose", "env-file:"+file, core.DiagnosticFail, fmt.Sprintf("Compose environment file is missing: %s", file)))
		}
	}
	if len(manifest.Compose.Files) == 0 {
		return checks
	}
	if _, err := s.commands.Find("docker"); err != nil {
		return append(checks, diagnostic("compose", "docker-compose", core.DiagnosticFail, "Docker Compose is unavailable because command \"docker\" was not found in PATH"))
	}
	if err := s.commands.CheckCompose(ctx); err != nil {
		return append(checks, diagnostic("compose", "docker-compose", core.DiagnosticFail, "`docker compose version` did not succeed: "+err.Error()))
	}
	return append(checks, diagnostic("compose", "docker-compose", core.DiagnosticPass, "`docker compose version` succeeded"))
}

func (s *DoctorProjectService) portChecks(ctx context.Context, manifest core.Manifest) []core.Diagnostic {
	checks := make([]core.Diagnostic, 0, len(manifest.Ports))
	for _, port := range manifest.Ports {
		if !port.Required {
			continue
		}
		observation, err := s.ports.InspectTCP(ctx, port.Port)
		name := fmt.Sprintf("port:%s", port.Name)
		if err != nil {
			checks = append(checks, diagnostic("ports", name, core.DiagnosticFail, err.Error()))
			continue
		}
		if observation.Available {
			checks = append(checks, diagnostic("ports", name, core.DiagnosticPass, fmt.Sprintf("Required TCP port %d is available", port.Port)))
			continue
		}
		owner := "owner unknown"
		if observation.PID > 0 || observation.Command != "" {
			owner = fmt.Sprintf("owner pid=%d command=%q", observation.PID, observation.Command)
		}
		checks = append(checks, diagnostic("ports", name, core.DiagnosticFail, fmt.Sprintf("Required TCP port %d is occupied by an unmanaged process (%s); Pivot will not terminate it", port.Port, owner)))
	}
	return checks
}

func diagnostic(category, name string, status core.DiagnosticStatus, message string) core.Diagnostic {
	return core.Diagnostic{Category: category, Name: name, Status: status, Message: message}
}
