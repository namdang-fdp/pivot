package adapters

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/namdang-fdp/pivot/internal/ports"
)

var (
	ssPIDPattern     = regexp.MustCompile(`pid=([0-9]+)`)
	ssCommandPattern = regexp.MustCompile(`users:\(\(\"([^\"]+)\"`)
)

// LinuxPortInspector observes TCP listeners without modifying them.
type LinuxPortInspector struct{}

// NewLinuxPortInspector creates a Linux TCP port inspector.
func NewLinuxPortInspector() *LinuxPortInspector { return &LinuxPortInspector{} }

// InspectTCP probes availability and attempts read-only owner enrichment with ss.
func (*LinuxPortInspector) InspectTCP(ctx context.Context, port int) (ports.PortObservation, error) {
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", fmt.Sprintf(":%d", port))
	if err == nil {
		if closeErr := listener.Close(); closeErr != nil {
			return ports.PortObservation{}, fmt.Errorf("close TCP port %d probe: %w", port, closeErr)
		}
		return ports.PortObservation{Available: true}, nil
	}
	var opErr *net.OpError
	if !errors.As(err, &opErr) || !strings.Contains(strings.ToLower(err.Error()), "address already in use") {
		return ports.PortObservation{}, fmt.Errorf("inspect TCP port %d: %w", port, err)
	}

	observation := ports.PortObservation{Available: false}
	ssPath, lookupErr := exec.LookPath("ss")
	if lookupErr != nil {
		return observation, nil
	}
	enrichmentContext, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	command := exec.CommandContext(enrichmentContext, ssPath, "-H", "-ltnp", fmt.Sprintf("sport = :%d", port))
	output, runErr := command.Output()
	if runErr != nil {
		return observation, nil
	}
	text := string(output)
	if match := ssPIDPattern.FindStringSubmatch(text); len(match) == 2 {
		observation.PID, _ = strconv.Atoi(match[1])
	}
	if match := ssCommandPattern.FindStringSubmatch(text); len(match) == 2 {
		observation.Command = match[1]
	}
	return observation, nil
}
