package main

import (
	"testing"
)

func TestPreprocessArgs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"numeric shorthand", []string{"-3"}, []string{"-n", "3"}},
		{"zero", []string{"-0"}, []string{"-n", "0"}},
		{"double digit", []string{"-10"}, []string{"-n", "10"}},
		{"non-numeric flag", []string{"-s"}, []string{"-s"}},
		{"mixed args", []string{"-3", "--starred", "@work"}, []string{"-n", "3", "--starred", "@work"}},
		{"long flag unchanged", []string{"--num", "5"}, []string{"--num", "5"}},
		{"no args", nil, nil},
		{"text args", []string{"hello", "world"}, []string{"hello", "world"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preprocessArgs(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("preprocessArgs(%v) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("preprocessArgs(%v)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}
