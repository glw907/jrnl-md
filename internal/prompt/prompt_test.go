package prompt

import (
	"bytes"
	"strings"
	"testing"
)

func TestYesNo(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantOut string
	}{
		{"y confirms", "y\n", true, "Continue? [y/N] "},
		{"Y confirms", "Y\n", true, "Continue? [y/N] "},
		{"n denies", "n\n", false, "Continue? [y/N] "},
		{"N denies", "N\n", false, "Continue? [y/N] "},
		{"empty defaults no", "\n", false, "Continue? [y/N] "},
		{"garbage defaults no", "maybe\n", false, "Continue? [y/N] "},
		{"yes defaults no", "yes\n", false, "Continue? [y/N] "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var out bytes.Buffer
			got := YesNo(reader, &out, "Continue?")
			if got != tt.want {
				t.Errorf("YesNo() = %v, want %v", got, tt.want)
			}
			if out.String() != tt.wantOut {
				t.Errorf("output = %q, want %q", out.String(), tt.wantOut)
			}
		})
	}
}
