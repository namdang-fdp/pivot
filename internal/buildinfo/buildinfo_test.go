package buildinfo

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCurrentDefaults(t *testing.T) {
	want := Info{Name: "pivot", Version: "dev", Commit: "unknown", Date: "unknown"}
	if got := Current(); got != want {
		t.Fatalf("Current() = %#v, want %#v", got, want)
	}
}

func TestInfoJSONFieldNames(t *testing.T) {
	data, err := json.Marshal(Current())
	if err != nil {
		t.Fatalf("marshal build info: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal build info: %v", err)
	}
	wantKeys := map[string]bool{"name": true, "version": true, "commit": true, "date": true}
	gotKeys := make(map[string]bool, len(got))
	for key := range got {
		gotKeys[key] = true
	}
	if !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("JSON keys = %#v, want %#v", gotKeys, wantKeys)
	}
}
