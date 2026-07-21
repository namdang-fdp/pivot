package cli

import "testing"

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple path",
			input: "/tmp/pivot",
			want:  "'/tmp/pivot'",
		},
		{
			name:  "path containing spaces",
			input: "/tmp/Pivot Demo",
			want:  "'/tmp/Pivot Demo'",
		},
		{
			name:  "path containing single quote",
			input: "/tmp/Nam's Project",
			want:  `'/tmp/Nam'"'"'s Project'`,
		},
		{
			name:  "empty value",
			input: "",
			want:  "''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shellQuote(tt.input)
			if got != tt.want {
				t.Fatalf("shellQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
