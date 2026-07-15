package schemas

import (
	"encoding/json"
	"os"
	"testing"
)

func TestPivotSchemaContract(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("pivot.schema.json")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var schema struct {
		Properties map[string]struct {
			Const any `json:"const"`
		} `json:"properties"`
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	if got := schema.Properties["version"].Const; got != float64(1) {
		t.Fatalf("version const = %#v, want 1", got)
	}
	want := []string{"version", "project", "requirements", "ports"}
	if len(schema.Required) != len(want) {
		t.Fatalf("required = %#v, want %#v", schema.Required, want)
	}
	for i := range want {
		if schema.Required[i] != want[i] {
			t.Fatalf("required = %#v, want %#v", schema.Required, want)
		}
	}
}
