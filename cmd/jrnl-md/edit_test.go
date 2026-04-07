package main

import "testing"

func TestStripEmptyTimestamp(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "trailing empty timestamp",
			in:   "\n## 09:00 AM\n\nContent.\n\n## 10:55 PM\n",
			want: "\n## 09:00 AM\n\nContent.\n",
		},
		{
			name: "trailing empty timestamp with extra newlines",
			in:   "\n## 09:00 AM\n\nContent.\n\n## 10:55 PM\n\n\n",
			want: "\n## 09:00 AM\n\nContent.\n",
		},
		{
			name: "only an empty timestamp",
			in:   "\n## 10:55 PM\n",
			want: "\n",
		},
		{
			name: "timestamp with content after it",
			in:   "\n## 09:00 AM\n\nContent.\n",
			want: "\n## 09:00 AM\n\nContent.\n",
		},
		{
			name: "no timestamp at all",
			in:   "\nSome text.\n",
			want: "\nSome text.\n",
		},
		{
			name: "empty body",
			in:   "\n",
			want: "\n",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := stripEmptyTimestamp(tt.in)
			if got != tt.want {
				t.Errorf("stripEmptyTimestamp:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}
