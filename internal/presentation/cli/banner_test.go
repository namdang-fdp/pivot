package cli

import (
	"strings"
	"testing"
)

func TestRootHelpShowsPivotBranding(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "root"},
		{name: "help flag", args: []string{"--help"}},
		{name: "help command", args: []string{"help"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, err := execute(t, test.args...)
			if err != nil {
				t.Fatalf("execute %v: %v", test.args, err)
			}
			if stderr != "" {
				t.Fatalf("stderr = %q, want empty", stderr)
			}

			bannerAt := strings.Index(stdout, pivotBanner)
			taglineAt := strings.Index(stdout, pivotTagline)
			descriptionAt := strings.Index(stdout, "Pivot safely orchestrates complete local development contexts.")
			if bannerAt < 0 || taglineAt <= bannerAt || descriptionAt <= taglineAt {
				t.Fatalf("root branding is missing or out of order:\n%s", stdout)
			}
		})
	}
}

func assertNoPivotBanner(t *testing.T, output string) {
	t.Helper()
	if strings.Contains(output, pivotBanner) || strings.Contains(output, pivotTagline) {
		t.Fatalf("command output contains root branding:\n%s", output)
	}
}
