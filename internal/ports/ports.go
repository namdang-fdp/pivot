package ports

import (
	"context"

	"github.com/namdang-fdp/pivot/internal/core"
)

// ManifestRepository reads and creates strict project manifests.
type ManifestRepository interface {
	Load(context.Context, string) (core.Manifest, error)
	Create(context.Context, string, core.Manifest) (string, error)
}

// ProjectRegistry stores explicitly registered projects.
type ProjectRegistry interface {
	List(context.Context) ([]core.RegisteredProject, error)
	Update(context.Context, func([]core.RegisteredProject) ([]core.RegisteredProject, error)) error
}

// Filesystem provides the path observations needed by Slice 1 use cases.
type Filesystem interface {
	CurrentDirectory() (string, error)
	CanonicalDirectory(string) (string, error)
	DirectoryExists(string) bool
	FileExists(string) bool
	FindManifestRoot(string) (string, bool, error)
}

// CommandObservation describes whether a host command resolves through PATH.
type CommandObservation struct {
	Path string
}

// CommandInspector performs bounded, read-only host command checks.
type CommandInspector interface {
	Find(string) (CommandObservation, error)
	CheckCompose(context.Context) error
}

// PortObservation describes local TCP availability and best-effort ownership data.
type PortObservation struct {
	Available bool
	PID       int
	Command   string
}

// PortInspector observes local TCP ports without modifying their owners.
type PortInspector interface {
	InspectTCP(context.Context, int) (PortObservation, error)
}
