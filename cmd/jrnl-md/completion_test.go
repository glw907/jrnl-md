package main

import (
	"bytes"
	"testing"
)

func TestCompletionSubcommands(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			cmd := newRootCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetArgs([]string{"completion", shell})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("completion %s failed: %v", shell, err)
			}

			if buf.Len() == 0 {
				t.Errorf("completion %s produced no output", shell)
			}
		})
	}
}
