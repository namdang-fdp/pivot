package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/namdang-fdp/pivot/internal/presentation/cli"
)

func TestIsolatedRegistryWorkflow(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	project := filepath.Join(t.TempDir(), "integration-project")
	if err := os.Mkdir(project, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	execute := func(args ...string) (string, error) {
		t.Helper()
		stdout := new(bytes.Buffer)
		command := cli.NewRootCommand()
		command.SetOut(stdout)
		command.SetErr(new(bytes.Buffer))
		command.SetArgs(args)
		err := command.ExecuteContext(context.Background())
		return stdout.String(), err
	}
	if _, err := execute("init", project, "--id", "integration-project", "--name", "Integration Project"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := execute("add", project); err != nil {
		t.Fatalf("add: %v", err)
	}
	output, err := execute("list", "--json")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var listed struct {
		Projects []struct {
			ID string `json:"id"`
		} `json:"projects"`
	}
	if err := json.Unmarshal([]byte(output), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Projects) != 1 || listed.Projects[0].ID != "integration-project" {
		t.Fatalf("listed = %#v", listed)
	}
	if _, err := os.Stat(filepath.Join(xdg, "pivot", "projects.yaml")); err != nil {
		t.Fatalf("isolated registry not created: %v", err)
	}
}
