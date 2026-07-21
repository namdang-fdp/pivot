package application

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/namdang-fdp/pivot/internal/adapters"
	"github.com/namdang-fdp/pivot/internal/core"
	"github.com/namdang-fdp/pivot/internal/ports"
)

type memoryRegistry struct {
	mu       sync.Mutex
	projects []core.RegisteredProject
}

func (r *memoryRegistry) List(context.Context) ([]core.RegisteredProject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]core.RegisteredProject(nil), r.projects...), nil
}

func (r *memoryRegistry) Update(_ context.Context, update func([]core.RegisteredProject) ([]core.RegisteredProject, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	projects, err := update(append([]core.RegisteredProject(nil), r.projects...))
	if err != nil {
		return err
	}
	r.projects = append([]core.RegisteredProject(nil), projects...)
	return nil
}

type fakeCommands struct {
	present  map[string]bool
	checkErr error
}

func (f fakeCommands) Find(name string) (ports.CommandObservation, error) {
	if !f.present[name] {
		return ports.CommandObservation{}, fmt.Errorf("command %q was not found in PATH", name)
	}
	return ports.CommandObservation{Path: "/test/bin/" + name}, nil
}

func (f fakeCommands) CheckCompose(context.Context) error { return f.checkErr }

type fakePortInspector struct {
	observations map[int]ports.PortObservation
	errors       map[int]error
}

func (f fakePortInspector) InspectTCP(_ context.Context, port int) (ports.PortObservation, error) {
	return f.observations[port], f.errors[port]
}

type fixedFilesystem struct {
	*adapters.HostFilesystem
	cwd string
}

func (f fixedFilesystem) CurrentDirectory() (string, error) { return f.cwd, nil }

func TestInitProjectGeneratesStrictManifest(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "CenterOS API")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	manifestRepository := adapters.NewYAMLManifestRepository()
	service := NewInitProjectService(manifestRepository, adapters.NewHostFilesystem())
	result, err := service.InitProject(context.Background(), InitProjectRequest{Path: root})
	if err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	if result.ManifestPath != filepath.Join(root, core.ManifestFilename) {
		t.Fatalf("manifest path = %q", result.ManifestPath)
	}
	manifest, err := manifestRepository.Load(context.Background(), root)
	if err != nil {
		t.Fatalf("load generated manifest: %v", err)
	}
	if manifest.Project.ID != "centeros-api" || manifest.Project.Name != "CenterOS API" {
		t.Fatalf("generated project = %#v", manifest.Project)
	}
	if len(manifest.Requirements.Commands) != 0 || len(manifest.Requirements.Files) != 0 || len(manifest.Ports) != 0 {
		t.Fatalf("generated manifest is not minimal: %#v", manifest)
	}
}

func TestAddProjectDuplicateRules(t *testing.T) {
	t.Parallel()
	manifestRepository := adapters.NewYAMLManifestRepository()
	registry := &memoryRegistry{}
	service := NewAddProjectService(manifestRepository, registry, adapters.NewHostFilesystem())
	first := createProject(t, manifestRepository, "centeros", "CenterOS", core.Manifest{})

	added, err := service.AddProject(context.Background(), first)
	if err != nil || added.AlreadyRegistered {
		t.Fatalf("first AddProject = %#v, %v", added, err)
	}
	idempotent, err := service.AddProject(context.Background(), first)
	if err != nil || !idempotent.AlreadyRegistered {
		t.Fatalf("idempotent AddProject = %#v, %v", idempotent, err)
	}

	second := createProject(t, manifestRepository, "centeros", "Other", core.Manifest{})
	if _, err := service.AddProject(context.Background(), second); err == nil || !strings.Contains(err.Error(), "already registered at") {
		t.Fatalf("duplicate ID error = %v", err)
	}
	registry.projects = []core.RegisteredProject{{ID: "different", Name: "Different", Path: first}}
	if _, err := service.AddProject(context.Background(), first); err == nil || !strings.Contains(err.Error(), "already registered as") {
		t.Fatalf("duplicate path error = %v", err)
	}
}

func TestListProjectsOrdersAndReportsAvailability(t *testing.T) {
	t.Parallel()
	availableRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(availableRoot, core.ManifestFilename), []byte("fixture\n"), 0o600); err != nil {
		t.Fatalf("write manifest fixture: %v", err)
	}
	missingManifestRoot := t.TempDir()
	registry := &memoryRegistry{projects: []core.RegisteredProject{
		{ID: "zeta", Name: "Zeta", Path: missingManifestRoot},
		{ID: "alpha", Name: "Alpha", Path: availableRoot},
	}}
	service := NewListProjectsService(registry, adapters.NewHostFilesystem())

	projects, err := service.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 2 || projects[0].ID != "alpha" || projects[1].ID != "zeta" {
		t.Fatalf("projects = %#v", projects)
	}
	if !projects[0].Available || projects[1].Available {
		t.Fatalf("availability = %#v", projects)
	}
}

func TestDoctorProjectPassesCompleteInspection(t *testing.T) {
	t.Parallel()
	manifestRepository := adapters.NewYAMLManifestRepository()
	manifest := doctorManifest()
	root := createProject(t, manifestRepository, "centeros", "CenterOS", manifest)
	writeDoctorFiles(t, root)
	registry := &memoryRegistry{projects: []core.RegisteredProject{{ID: "centeros", Name: "CenterOS", Path: root}}}
	service := NewDoctorProjectService(
		manifestRepository, registry, adapters.NewHostFilesystem(),
		fakeCommands{present: map[string]bool{"mise": true, "docker": true}},
		fakePortInspector{observations: map[int]ports.PortObservation{5432: {Available: true}}, errors: map[int]error{}},
	)
	result, err := service.DoctorProject(context.Background(), DoctorProjectRequest{Project: "centeros"})
	if err != nil {
		t.Fatalf("DoctorProject: %v", err)
	}
	if result.Status != core.DiagnosticPass || result.Summary.Failed != 0 || result.Summary.Warnings != 0 {
		t.Fatalf("result = %#v", result)
	}
	if result.Summary.Passed != len(result.Checks) {
		t.Fatalf("summary = %#v, checks = %d", result.Summary, len(result.Checks))
	}
}

func TestDoctorProjectReportsMissingRequirementsComposeAndOccupiedPort(t *testing.T) {
	t.Parallel()
	manifestRepository := adapters.NewYAMLManifestRepository()
	root := createProject(t, manifestRepository, "centeros", "CenterOS", doctorManifest())
	registry := &memoryRegistry{projects: []core.RegisteredProject{{ID: "centeros", Name: "CenterOS", Path: root}}}
	service := NewDoctorProjectService(
		manifestRepository, registry, adapters.NewHostFilesystem(),
		fakeCommands{present: map[string]bool{}},
		fakePortInspector{observations: map[int]ports.PortObservation{5432: {Available: false}}, errors: map[int]error{}},
	)
	result, err := service.DoctorProject(context.Background(), DoctorProjectRequest{Project: "centeros"})
	var failed *DoctorFailedError
	if !errors.As(err, &failed) {
		t.Fatalf("DoctorProject error = %v, want DoctorFailedError", err)
	}
	if result.Status != core.DiagnosticFail || result.Summary.Failed < 5 {
		t.Fatalf("result = %#v", result)
	}
	messages := diagnosticMessages(result.Checks)
	for _, expected := range []string{"command \"mise\"", "Required file is missing", "Compose file is missing", "Docker Compose is unavailable", "owner unknown"} {
		if !strings.Contains(messages, expected) {
			t.Errorf("diagnostics do not contain %q:\n%s", expected, messages)
		}
	}
}

func TestDoctorProjectReportsDockerComposeCheckFailure(t *testing.T) {
	t.Parallel()
	manifestRepository := adapters.NewYAMLManifestRepository()
	root := createProject(t, manifestRepository, "centeros", "CenterOS", doctorManifest())
	writeDoctorFiles(t, root)
	registry := &memoryRegistry{projects: []core.RegisteredProject{{ID: "centeros", Name: "CenterOS", Path: root}}}
	service := NewDoctorProjectService(
		manifestRepository, registry, adapters.NewHostFilesystem(),
		fakeCommands{present: map[string]bool{"mise": true, "docker": true}, checkErr: errors.New("compose plugin unavailable")},
		fakePortInspector{observations: map[int]ports.PortObservation{5432: {Available: true}}, errors: map[int]error{}},
	)
	result, err := service.DoctorProject(context.Background(), DoctorProjectRequest{Project: "centeros"})
	var failed *DoctorFailedError
	if !errors.As(err, &failed) {
		t.Fatalf("DoctorProject error = %v, want DoctorFailedError", err)
	}
	if !strings.Contains(diagnosticMessages(result.Checks), "docker compose version") {
		t.Fatalf("diagnostics = %#v", result.Checks)
	}
}

func TestDoctorProjectWarningsDoNotFailAndSummaryIsDeterministic(t *testing.T) {
	t.Parallel()
	manifestRepository := adapters.NewYAMLManifestRepository()
	root := createProject(t, manifestRepository, "direct-project", "Direct Project", core.Manifest{})
	files := fixedFilesystem{HostFilesystem: adapters.NewHostFilesystem(), cwd: root}
	service := NewDoctorProjectService(manifestRepository, &memoryRegistry{}, files, fakeCommands{present: map[string]bool{}}, fakePortInspector{})
	first, err := service.DoctorProject(context.Background(), DoctorProjectRequest{})
	if err != nil {
		t.Fatalf("DoctorProject: %v", err)
	}
	second, err := service.DoctorProject(context.Background(), DoctorProjectRequest{})
	if err != nil {
		t.Fatalf("second DoctorProject: %v", err)
	}
	if first.Status != core.DiagnosticWarning || first.Summary.Warnings != 1 || first.Summary.Failed != 0 {
		t.Fatalf("first result = %#v", first)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("doctor result is not deterministic:\nfirst=%#v\nsecond=%#v", first, second)
	}
}

func doctorManifest() core.Manifest {
	return core.Manifest{
		Requirements: core.Requirements{Commands: []string{"mise"}, Files: []string{".env"}},
		Compose:      core.Compose{Files: []string{"ops/compose.yml"}, EnvFiles: []string{".env"}},
		Ports:        []core.Port{{Name: "postgres", Port: 5432, Protocol: "tcp", Required: true}},
	}
}

func createProject(t *testing.T, repository *adapters.YAMLManifestRepository, id core.ProjectID, name string, additions core.Manifest) string {
	t.Helper()
	root := t.TempDir()
	manifest := additions
	manifest.Version = 1
	manifest.Project = core.Project{ID: id, Name: name}
	if manifest.Requirements.Commands == nil {
		manifest.Requirements.Commands = []string{}
	}
	if manifest.Requirements.Files == nil {
		manifest.Requirements.Files = []string{}
	}
	if manifest.Ports == nil {
		manifest.Ports = []core.Port{}
	}
	if _, err := repository.Create(context.Background(), root, manifest); err != nil {
		t.Fatalf("create project manifest: %v", err)
	}
	return root
}

func writeDoctorFiles(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "ops"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, file := range []string{".env", "ops/compose.yml"} {
		if err := os.WriteFile(filepath.Join(root, file), []byte("fixture\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", file, err)
		}
	}
}

func diagnosticMessages(checks []core.Diagnostic) string {
	var messages []string
	for _, check := range checks {
		messages = append(messages, check.Message)
	}
	return strings.Join(messages, "\n")
}
