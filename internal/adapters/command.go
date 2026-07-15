package adapters

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/namdang-fdp/pivot/internal/ports"
)

// HostCommandInspector performs read-only command availability checks.
type HostCommandInspector struct{}

// NewHostCommandInspector creates a host command inspector.
func NewHostCommandInspector() *HostCommandInspector { return &HostCommandInspector{} }

// Find resolves a command through the current PATH without executing it.
func (*HostCommandInspector) Find(name string) (ports.CommandObservation, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return ports.CommandObservation{}, fmt.Errorf("command %q was not found in PATH", name)
	}
	return ports.CommandObservation{Path: path}, nil
}

// CheckCompose verifies the configured Compose CLI with a read-only version call.
func (*HostCommandInspector) CheckCompose(ctx context.Context) error {
	checkContext, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	command := exec.CommandContext(checkContext, "docker", "compose", "version")
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Run(); err != nil {
		return fmt.Errorf("`docker compose version` failed: %w", err)
	}
	return nil
}
