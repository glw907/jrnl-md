package journal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	root := t.TempDir()
	s := NewStore(root, "2006-01-02", "03:04 PM", "@")
	return s, root
}

func TestStoreLoadNotFound(t *testing.T) {
	s, _ := newTestStore(t)
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	_, err := s.Load(date)
	if err == nil {
		t.Fatal("Load non-existent: expected error, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Load non-existent: expected ErrNotExist, got %v", err)
	}
}

func TestStoreSaveAndLoad(t *testing.T) {
	s, _ := newTestStore(t)
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	day := Day{
		Date: date,
		Body: "\nWent for a morning run.\n",
	}

	if err := s.Save(day); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := s.Load(date)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !loaded.Date.Equal(date) {
		t.Errorf("Date: got %v, want %v", loaded.Date, date)
	}
	if !strings.Contains(loaded.Body, "morning run") {
		t.Errorf("Body: missing content: %q", loaded.Body)
	}
}

func TestStoreSaveCreatesDirectories(t *testing.T) {
	s, root := newTestStore(t)
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	day := Day{Date: date, Body: "\nTest.\n"}

	if err := s.Save(day); err != nil {
		t.Fatalf("Save: %v", err)
	}

	path := filepath.Join(root, "2026", "04", "06.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file at %s: %v", path, err)
	}
}

func TestStoreDelete(t *testing.T) {
	s, _ := newTestStore(t)
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	day := Day{Date: date, Body: "\nTest.\n"}

	if err := s.Save(day); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := s.Delete(date); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Load(date); !os.IsNotExist(err) {
		t.Errorf("after Delete: expected ErrNotExist, got %v", err)
	}
}

func TestStoreDeleteNotFound(t *testing.T) {
	s, _ := newTestStore(t)
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	err := s.Delete(date)
	if err == nil {
		t.Fatal("Delete non-existent: expected error, got nil")
	}
}

func TestStoreDeleteCleansEmptyDirs(t *testing.T) {
	s, root := newTestStore(t)
	date := time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)
	day := Day{Date: date, Body: "\nTest.\n"}

	if err := s.Save(day); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := s.Delete(date); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	monthDir := filepath.Join(root, "2026", "04")
	if _, err := os.Stat(monthDir); !os.IsNotExist(err) {
		t.Errorf("empty month dir not cleaned: %v", err)
	}
}

func TestStoreAppendTimestampsOn(t *testing.T) {
	s, _ := newTestStore(t)
	if err := s.Append("Hello world."); err != nil {
		t.Fatalf("Append: %v", err)
	}
	today := time.Now()
	day, err := s.Load(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Load after Append: %v", err)
	}
	if !strings.Contains(day.Body, "Hello world.") {
		t.Errorf("Body missing appended text: %q", day.Body)
	}
}

func TestStoreAppendTimestampsOff(t *testing.T) {
	root := t.TempDir()
	s := NewStore(root, "2006-01-02", "", "@")
	if err := s.Append("No timestamp."); err != nil {
		t.Fatalf("Append: %v", err)
	}
	today := time.Now()
	day, err := s.Load(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Load after Append: %v", err)
	}
	if strings.Contains(day.Body, "## ") {
		t.Errorf("timestamps off but found ## heading: %q", day.Body)
	}
	if !strings.Contains(day.Body, "No timestamp.") {
		t.Errorf("Body missing text: %q", day.Body)
	}
}

func TestStoreAppendSecondWrite(t *testing.T) {
	s, _ := newTestStore(t)
	if err := s.Append("First write."); err != nil {
		t.Fatalf("first Append: %v", err)
	}
	if err := s.Append("Second write."); err != nil {
		t.Fatalf("second Append: %v", err)
	}
	today := time.Now()
	day, err := s.Load(time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !strings.Contains(day.Body, "First write.") || !strings.Contains(day.Body, "Second write.") {
		t.Errorf("Body missing content: %q", day.Body)
	}
}

func TestStoreList(t *testing.T) {
	s, _ := newTestStore(t)
	dates := []time.Time{
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
	}
	for _, d := range dates {
		if err := s.Save(Day{Date: d, Body: "\nContent for " + d.Format("2006-01-02") + "\n"}); err != nil {
			t.Fatalf("Save %v: %v", d, err)
		}
	}

	f := Filter{}
	days, err := s.List(f)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(days) != 3 {
		t.Errorf("List: got %d days, want 3", len(days))
	}
}

func TestStoreListOrder(t *testing.T) {
	s, _ := newTestStore(t)
	for _, d := range []time.Time{
		time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC),
	} {
		if err := s.Save(Day{Date: d, Body: "\nContent.\n"}); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	f := Filter{}
	days, err := s.List(f)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i := 1; i < len(days); i++ {
		if !days[i].Date.Before(days[i-1].Date) {
			t.Errorf("List not sorted newest-first: days[%d]=%v, days[%d]=%v",
				i-1, days[i-1].Date, i, days[i].Date)
		}
	}
}

func TestStoreTags(t *testing.T) {
	s, _ := newTestStore(t)
	for d, body := range map[time.Time]string{
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC): "\nMet with @alice about @work.\n",
		time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC): "\n@alice sent a note about @work.\n",
		time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC): "\nRead a book. @reading.\n",
	} {
		if err := s.Save(Day{Date: d, Body: body}); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	tags, err := s.Tags(Filter{})
	if err != nil {
		t.Fatalf("Tags: %v", err)
	}
	if tags["@alice"] != 2 {
		t.Errorf("@alice: got %d, want 2", tags["@alice"])
	}
	if tags["@work"] != 2 {
		t.Errorf("@work: got %d, want 2", tags["@work"])
	}
	if tags["@reading"] != 1 {
		t.Errorf("@reading: got %d, want 1", tags["@reading"])
	}
}
