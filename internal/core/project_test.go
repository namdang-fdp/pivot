package core

import "testing"

func TestParseProjectID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "letters", value: "centeros"},
		{name: "digits and hyphens", value: "project-2-api"},
		{name: "uppercase", value: "CenterOS", wantErr: true},
		{name: "underscore", value: "center_os", wantErr: true},
		{name: "leading hyphen", value: "-centeros", wantErr: true},
		{name: "trailing hyphen", value: "centeros-", wantErr: true},
		{name: "double hyphen", value: "center--os", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseProjectID(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseProjectID(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestSlugifyProjectID(t *testing.T) {
	t.Parallel()
	got, err := SlugifyProjectID("CenterOS API_v2")
	if err != nil {
		t.Fatalf("SlugifyProjectID: %v", err)
	}
	if got != "centeros-api-v2" {
		t.Fatalf("SlugifyProjectID = %q, want centeros-api-v2", got)
	}
}
