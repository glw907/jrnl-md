package e2e

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	tests := []struct {
		name string
		flag string
	}{
		{"long flag", "--version"},
		{"short flag", "-v"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			stdout, stderr := run(t, env, tc.flag)
			combined := stdout + stderr
			if !strings.Contains(combined, "jrnl-md") {
				t.Errorf("expected output to contain %q, got stdout=%q stderr=%q", "jrnl-md", stdout, stderr)
			}
		})
	}
}
