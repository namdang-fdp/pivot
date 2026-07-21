package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlice1CLIInitAddListAndDoctor(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	project := filepath.Join(t.TempDir(), "Sample Project")
	if err := os.Mkdir(project, 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	stdout, stderr, err := execute(t, "init", project, "--id", "sample-project", "--name", "Sample Project")
	if err != nil || stderr != "" {
		t.Fatalf("init stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}

	wantInitOutput := "Created manifest: " +
		filepath.Join(project, ".pivot.yaml") +
		"\nNext: pivot add " +
		shellQuote(project) +
		"\n"
	if stdout != wantInitOutput {
		t.Fatalf("init output = %q, want %q", stdout, wantInitOutput)
	}
	assertNoPivotBanner(t, stdout)
	if _, err := os.Stat(filepath.Join(project, ".pivot.yaml")); err != nil {
		t.Fatalf("generated manifest: %v", err)
	}

	stdout, stderr, err = execute(t, "add", project)
	if err != nil || stderr != "" || !strings.Contains(stdout, "Registered project sample-project") {
		t.Fatalf("add stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	assertNoPivotBanner(t, stdout)
	stdout, _, err = execute(t, "add", project)
	if err != nil || !strings.Contains(stdout, "already registered") {
		t.Fatalf("idempotent add stdout=%q err=%v", stdout, err)
	}

	stdout, stderr, err = execute(t, "list")
	if err != nil || stderr != "" {
		t.Fatalf("list stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	for _, value := range []string{"PROJECT", "sample-project", "Sample Project", "true"} {
		if !strings.Contains(stdout, value) {
			t.Errorf("list output lacks %q:\n%s", value, stdout)
		}
	}
	assertNoPivotBanner(t, stdout)

	stdout, stderr, err = execute(t, "list", "--json")
	if err != nil || stderr != "" {
		t.Fatalf("list JSON stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	var listed listJSON
	if err := json.Unmarshal([]byte(stdout), &listed); err != nil {
		t.Fatalf("list JSON: %v", err)
	}
	if len(listed.Projects) != 1 || listed.Projects[0].ID != "sample-project" || !listed.Projects[0].Available {
		t.Fatalf("listed projects = %#v", listed)
	}
	assertNoPivotBanner(t, stdout)

	stdout, stderr, err = execute(t, "doctor", "sample-project")
	if err != nil || stderr != "" || !strings.Contains(stdout, "Summary:") || !strings.Contains(stdout, "0 failed") {
		t.Fatalf("doctor stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	assertNoPivotBanner(t, stdout)
	stdout, stderr, err = execute(t, "doctor", "sample-project", "--json")
	if err != nil || stderr != "" {
		t.Fatalf("doctor JSON stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	var doctor doctorJSON
	if err := json.Unmarshal([]byte(stdout), &doctor); err != nil {
		t.Fatalf("doctor JSON: %v", err)
	}
	if doctor.Project.ID != "sample-project" || doctor.Status != "pass" || doctor.Summary.Failed != 0 {
		t.Fatalf("doctor result = %#v", doctor)
	}
	assertNoPivotBanner(t, stdout)
}

func TestListEmptyRegistry(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	stdout, stderr, err := execute(t, "list")
	if err != nil || stderr != "" || !strings.Contains(stdout, "No projects registered") {
		t.Fatalf("list stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	stdout, _, err = execute(t, "list", "--json")
	if err != nil || stdout != "{\n  \"projects\": []\n}\n" {
		t.Fatalf("empty list JSON = %q, err=%v", stdout, err)
	}
}

func TestDoctorFailureReturnsErrorAndValidJSON(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	project := t.TempDir()
	manifest := `version: 1
project:
  id: failing-project
  name: Failing Project
requirements:
  commands:
    - command-that-does-not-exist-pivot-test
  files:
    - missing.env
ports: []
`
	if err := os.WriteFile(filepath.Join(project, ".pivot.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if _, _, err := execute(t, "add", project); err != nil {
		t.Fatalf("add: %v", err)
	}
	stdout, stderr, err := execute(t, "doctor", "failing-project", "--json")
	if err == nil {
		t.Fatal("doctor error = nil, want failed check error")
	}
	if stderr != "" {
		t.Fatalf("doctor stderr = %q", stderr)
	}
	var result doctorJSON
	if decodeErr := json.Unmarshal([]byte(stdout), &result); decodeErr != nil {
		t.Fatalf("doctor stdout is not valid JSON: %v\n%s", decodeErr, stdout)
	}
	if result.Status != "fail" || result.Summary.Failed != 2 {
		t.Fatalf("doctor result = %#v", result)
	}
}

func TestSlice1CLIInvalidArguments(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, args := range [][]string{
		{"init", "one", "two"},
		{"add", "one", "two"},
		{"list", "unexpected"},
		{"doctor", "one", "two"},
	} {
		stdout, _, err := execute(t, args...)
		if err == nil {
			t.Errorf("execute(%v) error = nil", args)
		}
		assertNoPivotBanner(t, stdout)
	}
}

func TestRootHelpListsOnlyImplementedProductCommands(t *testing.T) {
	stdout, _, err := execute(t, "--help")
	if err != nil {
		t.Fatalf("help: %v", err)
	}
	for _, command := range []string{"add", "doctor", "init", "list", "version"} {
		if !strings.Contains(stdout, command) {
			t.Errorf("help does not contain %q:\n%s", command, stdout)
		}
	}
	for _, excluded := range []string{"switch", "status", "resume", "shutdown"} {
		if strings.Contains(stdout, excluded) {
			t.Errorf("help exposes excluded command %q:\n%s", excluded, stdout)
		}
	}
}
