package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/namdang-fdp/pivot/internal/buildinfo"
)

func execute(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := NewRootCommand()
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

func TestVersion(t *testing.T) {
	stdout, stderr, err := execute(t, "version")
	if err != nil {
		t.Fatalf("execute version: %v", err)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	want := "pivot dev\ncommit: unknown\nbuilt: unknown\n"
	if stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestVersionJSON(t *testing.T) {
	stdout, stderr, err := execute(t, "version", "--json")
	if err != nil {
		t.Fatalf("execute JSON version: %v", err)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	var got buildinfo.Info
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout is not valid JSON: %v", err)
	}
	want := buildinfo.Info{Name: "pivot", Version: "dev", Commit: "unknown", Date: "unknown"}
	if got != want {
		t.Fatalf("version JSON = %#v, want %#v", got, want)
	}
}

func TestRootShowsHelp(t *testing.T) {
	stdout, stderr, err := execute(t)
	if err != nil {
		t.Fatalf("execute root: %v", err)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	for _, text := range []string{"Usage:", "pivot [flags]", "version"} {
		if !strings.Contains(stdout, text) {
			t.Errorf("root help does not contain %q:\n%s", text, stdout)
		}
	}
	if strings.Contains(stdout, "completion") {
		t.Fatalf("root help exposes an out-of-scope completion command:\n%s", stdout)
	}
}
