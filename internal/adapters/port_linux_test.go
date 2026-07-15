package adapters

import (
	"context"
	"net"
	"strconv"
	"testing"
)

func TestLinuxPortInspectorAvailable(t *testing.T) {
	t.Parallel()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("release port: %v", err)
	}
	observation, err := NewLinuxPortInspector().InspectTCP(context.Background(), port)
	if err != nil {
		t.Fatalf("InspectTCP: %v", err)
	}
	if !observation.Available {
		t.Fatalf("port %s reported occupied", strconv.Itoa(port))
	}
}

func TestLinuxPortInspectorOccupiedWithoutOwnerEnrichment(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()
	t.Setenv("PATH", t.TempDir())
	port := listener.Addr().(*net.TCPAddr).Port
	observation, err := NewLinuxPortInspector().InspectTCP(context.Background(), port)
	if err != nil {
		t.Fatalf("InspectTCP: %v", err)
	}
	if observation.Available || observation.PID != 0 || observation.Command != "" {
		t.Fatalf("observation = %#v", observation)
	}
}
