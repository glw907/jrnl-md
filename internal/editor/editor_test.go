package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveMissingEditor(t *testing.T) {
	os.Unsetenv("VISUAL")
	os.Unsetenv("EDITOR")
	e := Resolve("")
	_ = e
}

func TestResolveEditorFromArg(t *testing.T) {
	os.Unsetenv("VISUAL")
	os.Unsetenv("EDITOR")
	e := Resolve("micro")
	if e != "micro" {
		t.Errorf("Resolve: got %q, want micro", e)
	}
}

func TestResolveEditorVisual(t *testing.T) {
	os.Setenv("VISUAL", "vim")
	os.Setenv("EDITOR", "nano")
	defer os.Unsetenv("VISUAL")
	defer os.Unsetenv("EDITOR")

	e := Resolve("")
	if e != "vim" {
		t.Errorf("Resolve with VISUAL=vim: got %q, want vim", e)
	}
}

func TestResolveEditorFallsBackToEditor(t *testing.T) {
	os.Unsetenv("VISUAL")
	os.Setenv("EDITOR", "nano")
	defer os.Unsetenv("EDITOR")

	e := Resolve("")
	if e != "nano" {
		t.Errorf("Resolve with EDITOR=nano, no VISUAL: got %q, want nano", e)
	}
}

func TestBuildArgs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(path, []byte("# Test\n\nContent.\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	args := buildArgs("nvim", path, 3)
	if len(args) == 0 {
		t.Fatal("buildArgs: returned empty args")
	}
	if args[0] != "nvim" {
		t.Errorf("buildArgs: first arg should be editor: %q", args[0])
	}
	found := false
	for _, a := range args {
		if a == path {
			found = true
		}
	}
	if !found {
		t.Errorf("buildArgs: path not in args: %v", args)
	}
}

func TestBuildArgsVim(t *testing.T) {
	path := "/tmp/test.md"
	args := buildArgs("vim", path, 3)
	if len(args) < 3 {
		t.Fatalf("buildArgs vim: expected at least 3 args, got %d: %v", len(args), args)
	}
	if args[1] != "+3" {
		t.Errorf("buildArgs vim: expected +3, got %q", args[1])
	}
}

func TestBuildArgsMicro(t *testing.T) {
	path := "/tmp/test.md"
	args := buildArgs("micro", path, 5)
	if len(args) < 2 {
		t.Fatalf("buildArgs micro: expected at least 2 args, got %d: %v", len(args), args)
	}
	if args[1] != path+":5" {
		t.Errorf("buildArgs micro: expected %q:5, got %q", path, args[1])
	}
}

func TestCursorLine(t *testing.T) {
	cases := []struct {
		content  string
		wantLine int
	}{
		{"# 2026-04-06 Monday\n\nContent.\n", 3},
		{"# 2026-04-06 Monday\n\n## 09:00 AM\n\nContent.\n", 3},
		{"# 2026-04-06 Monday\n", 2},
	}
	for _, tt := range cases {
		got := cursorLine(tt.content)
		if got != tt.wantLine {
			t.Errorf("cursorLine(%q): got %d, want %d", tt.content, got, tt.wantLine)
		}
	}
}
