package journal

import (
	"testing"
)

func TestCleanupDayContent(t *testing.T) {
	t.Run("strips trailing empty entry heading", func(t *testing.T) {
		input := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nReal entry.\n\n## [02:30 PM]\n\n"
		got := CleanupDayContent(input)
		want := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nReal entry.\n"
		if got != want {
			t.Errorf("got:\n%q\nwant:\n%q", got, want)
		}
	})

	t.Run("preserves non-empty trailing entry", func(t *testing.T) {
		input := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nReal entry.\n\n## [02:30 PM]\n\nAlso real.\n"
		got := CleanupDayContent(input)
		if got != input {
			t.Errorf("should not change content with no empty headings:\ngot: %q\nwant: %q", got, input)
		}
	})

	t.Run("normalizes blank lines before ## headings", func(t *testing.T) {
		input := "# 2026-03-29 Sunday\n\n\n\n## [09:00 AM]\n\nEntry.\n\n\n\n\n## [02:30 PM]\n\nSecond.\n"
		got := CleanupDayContent(input)
		want := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nEntry.\n\n## [02:30 PM]\n\nSecond.\n"
		if got != want {
			t.Errorf("got:\n%q\nwant:\n%q", got, want)
		}
	})

	t.Run("trims trailing whitespace from lines", func(t *testing.T) {
		input := "# 2026-03-29 Sunday  \n\n## [09:00 AM]  \n\nEntry with spaces.  \n"
		got := CleanupDayContent(input)
		want := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nEntry with spaces.\n"
		if got != want {
			t.Errorf("got:\n%q\nwant:\n%q", got, want)
		}
	})

	t.Run("ensures single trailing newline", func(t *testing.T) {
		input := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nEntry.\n\n\n\n"
		got := CleanupDayContent(input)
		want := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nEntry.\n"
		if got != want {
			t.Errorf("got:\n%q\nwant:\n%q", got, want)
		}
	})

	t.Run("no change returns identical content", func(t *testing.T) {
		input := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nClean content.\n"
		got := CleanupDayContent(input)
		if got != input {
			t.Errorf("expected no change, got:\n%q", got)
		}
	})
}
