package journal

import (
	"strings"
	"testing"
	"time"
)

func TestParseDay(t *testing.T) {
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	input := "# 2026-04-06 Monday\n\nWent for a morning run.\n\nSecond paragraph.\n"

	day, err := parseDay(date, input)
	if err != nil {
		t.Fatalf("parseDay: %v", err)
	}
	if !day.Date.Equal(date) {
		t.Errorf("Date: got %v, want %v", day.Date, date)
	}
	if !strings.Contains(day.Body, "morning run") {
		t.Errorf("Body missing expected content: %q", day.Body)
	}
	if strings.HasPrefix(strings.TrimSpace(day.Body), "#") {
		t.Errorf("Body should not start with heading: %q", day.Body)
	}
}

func TestParseDayEmptyBody(t *testing.T) {
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	input := "# 2026-04-06 Monday\n"

	day, err := parseDay(date, input)
	if err != nil {
		t.Fatalf("parseDay: %v", err)
	}
	if strings.TrimSpace(day.Body) != "" {
		t.Errorf("Body should be empty: %q", day.Body)
	}
}

func TestFormatDay(t *testing.T) {
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	day := Day{
		Date: date,
		Body: "\nWent for a morning run.\n",
	}

	out := formatDay(day)
	if !strings.HasPrefix(out, "# 2026-04-06 Monday") {
		t.Errorf("formatDay: missing heading: %q", out)
	}
	if !strings.Contains(out, "morning run") {
		t.Errorf("formatDay: missing body: %q", out)
	}
}

func TestFormatDayRoundtrip(t *testing.T) {
	date := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	original := Day{
		Date: date,
		Body: "\n## 09:00 AM\n\nMorning thoughts.\n\n## 02:30 PM\n\nAfternoon work.\n",
	}

	formatted := formatDay(original)
	parsed, err := parseDay(date, formatted)
	if err != nil {
		t.Fatalf("parseDay after formatDay: %v", err)
	}
	if parsed.Body != original.Body {
		t.Errorf("roundtrip body mismatch:\ngot:  %q\nwant: %q", parsed.Body, original.Body)
	}
}

func TestHeadingLine(t *testing.T) {
	cases := []struct {
		date time.Time
		want string
	}{
		{time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC), "# 2026-04-06 Monday"},
		{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "# 2026-01-01 Thursday"},
		{time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "# 2025-12-31 Wednesday"},
	}
	for _, tt := range cases {
		t.Run(tt.want, func(t *testing.T) {
			got := headingLine(tt.date)
			if got != tt.want {
				t.Errorf("headingLine: got %q, want %q", got, tt.want)
			}
		})
	}
}
