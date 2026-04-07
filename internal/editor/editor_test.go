package editor

import (
	"path/filepath"
	"testing"
)

func TestBuildArgs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.md")
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

func TestBuildArgsNvimWrapper(t *testing.T) {
	path := "/tmp/test.md"
	args := buildArgs("nvim-journal", path, 5)
	if len(args) < 3 {
		t.Fatalf("buildArgs nvim-journal: expected at least 3 args, got %d: %v", len(args), args)
	}
	if args[1] != "+5" {
		t.Errorf("buildArgs nvim-journal: expected +5, got %q", args[1])
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

func TestEnsureEntryPoint(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "content ending with single newline",
			in:   "# 2026-04-06 Monday\n\nContent.\n",
			want: "# 2026-04-06 Monday\n\nContent.\n\n\n",
		},
		{
			name: "content already correct",
			in:   "# 2026-04-06 Monday\n\nContent.\n\n\n",
			want: "# 2026-04-06 Monday\n\nContent.\n\n\n",
		},
		{
			name: "excess trailing newlines",
			in:   "# 2026-04-06 Monday\n\nContent.\n\n\n\n",
			want: "# 2026-04-06 Monday\n\nContent.\n\n\n",
		},
		{
			name: "new file without timestamp",
			in:   "# 2026-04-06 Monday\n",
			want: "# 2026-04-06 Monday\n\n\n",
		},
		{
			name: "new file with timestamp",
			in:   "# 2026-04-06 Monday\n\n## 09:00 AM\n",
			want: "# 2026-04-06 Monday\n\n## 09:00 AM\n\n\n",
		},
		{
			name: "no trailing newline",
			in:   "# 2026-04-06 Monday",
			want: "# 2026-04-06 Monday\n\n\n",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := ensureEntryPoint(tt.in)
			if got != tt.want {
				t.Errorf("ensureEntryPoint:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}
