package journal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var testOpts = Options{
	DateFmt:    "2006-01-02",
	TimeFmt:    "03:04 PM",
	TagSymbols: "@",
	FileExt:    "md",
}

func writeDayFile(t *testing.T, base string, date time.Time, content string, ext string) {
	t.Helper()
	dir := filepath.Join(base, date.Format("2006"), date.Format("01"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, date.Format("02")+"."+ext)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestFolderOpen(t *testing.T) {
	dir := t.TempDir()
	content1 := "# 2026-03-28 Saturday\n\n## [09:00 AM]\n\nYesterday's entry.\n"
	content2 := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nToday's entry.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local), content1, "md")
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content2, "md")

	fj := NewFolderJournal(dir, testOpts)
	entries, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatalf("Entries failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Body != "Yesterday's entry." {
		t.Errorf("entry 0 body = %q", entries[0].Body)
	}
	if entries[1].Body != "Today's entry." {
		t.Errorf("entry 1 body = %q", entries[1].Body)
	}
}

func TestFolderWrite(t *testing.T) {
	dir := t.TempDir()
	fj := NewFolderJournal(dir, testOpts)
	if err := fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Test entry.", false); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}
	path := filepath.Join(dir, "2026", "03", "29.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "# 2026-03-29") {
		t.Errorf("missing day title, got: %q", content[:30])
	}
	if !strings.Contains(content, "Test entry.") {
		t.Error("missing entry body")
	}
	if !strings.Contains(content, "## [09:00 AM]") {
		t.Error("missing entry heading")
	}
}

func TestFolderAddEntryAppendsToExistingDay(t *testing.T) {
	dir := t.TempDir()
	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMorning entry.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.AddEntry(time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local), "Afternoon entry.", false); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "2026", "03", "29.md"))
	if err != nil {
		t.Fatal(err)
	}
	result := string(data)
	if !strings.Contains(result, "Morning entry.") {
		t.Error("lost morning entry")
	}
	if !strings.Contains(result, "Afternoon entry.") {
		t.Error("missing afternoon entry")
	}
	if !strings.Contains(result, "## [02:30 PM]") {
		t.Error("missing afternoon heading")
	}
}

func TestFolderCustomExtension(t *testing.T) {
	dir := t.TempDir()
	opts := testOpts
	opts.FileExt = "txt"
	fj := NewFolderJournal(dir, opts)
	if err := fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Test.", false); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}
	path := filepath.Join(dir, "2026", "03", "29.txt")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected .txt file, got error: %v", err)
	}
}

func TestDeleteEntries(t *testing.T) {
	dir := t.TempDir()
	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nFirst.\n\n## [10:00 AM]\n\nSecond.\n\n## [11:00 AM]\n\nThird.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	entries, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if err := fj.DeleteEntry(entries[1]); err != nil {
		t.Fatal(err)
	}
	remaining, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 2 {
		t.Fatalf("expected 2 entries after delete, got %d", len(remaining))
	}
	if remaining[0].Body != "First." {
		t.Errorf("entry 0 body = %q, want %q", remaining[0].Body, "First.")
	}
	if remaining[1].Body != "Third." {
		t.Errorf("entry 1 body = %q, want %q", remaining[1].Body, "Third.")
	}
	data, err := os.ReadFile(filepath.Join(dir, "2026", "03", "29.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Second.") {
		t.Error("deleted entry still in file")
	}
}

func TestDeleteEntriesEmpty(t *testing.T) {
	dir := t.TempDir()
	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nKeep me.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	entries, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestDeleteEntriesAcrossDays(t *testing.T) {
	dir := t.TempDir()
	content1 := "# 2026-03-28 Saturday\n\n## [09:00 AM]\n\nDay one.\n"
	content2 := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nDay two.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local), content1, "md")
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content2, "md")

	fj := NewFolderJournal(dir, testOpts)
	entries, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if err := fj.DeleteEntry(entries[0]); err != nil {
		t.Fatal(err)
	}
	remaining, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(remaining))
	}
	if remaining[0].Body != "Day two." {
		t.Errorf("wrong entry remained: %q", remaining[0].Body)
	}
}

func TestChangeEntryTimesSameDay(t *testing.T) {
	dir := t.TempDir()
	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMorning.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	entries, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	newTime := time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local)
	updated := entries[0]
	updated.Date = newTime
	if err := fj.UpdateEntry(entries[0], updated); err != nil {
		t.Fatal(err)
	}
	result, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if !result[0].Date.Equal(newTime) {
		t.Errorf("date = %v, want %v", result[0].Date, newTime)
	}
	data, err := os.ReadFile(filepath.Join(dir, "2026", "03", "29.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "## [02:00 PM]") {
		t.Errorf("expected new time heading, got: %s", string(data))
	}
}

func TestEncryptedRoundTrip(t *testing.T) {
	dir := t.TempDir()
	opts := testOpts
	opts.Encrypt = true
	opts.Passphrase = "testpass"

	fj := NewFolderJournal(dir, opts)
	if err := fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Secret entry.", false); err != nil {
		t.Fatalf("AddEntry failed: %v", err)
	}
	path := filepath.Join(dir, "2026", "03", "29.md.age")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected .md.age file: %v", err)
	}
	if strings.Contains(string(data), "Secret entry.") {
		t.Error("encrypted file contains plaintext")
	}
	plainPath := filepath.Join(dir, "2026", "03", "29.md")
	if _, err := os.Stat(plainPath); err == nil {
		t.Error("plaintext file should not exist")
	}
	fj2 := NewFolderJournal(dir, opts)
	entries, err := fj2.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Body != "Secret entry." {
		t.Errorf("body = %q, want %q", entries[0].Body, "Secret entry.")
	}
}

func TestEncryptDecryptConversion(t *testing.T) {
	dir := t.TempDir()
	fj := NewFolderJournal(dir, testOpts)
	if err := fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Convert me.", false); err != nil {
		t.Fatal(err)
	}
	_, err := fj.ReencryptAll(true, "pass123")
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	encPath := filepath.Join(dir, "2026", "03", "29.md.age")
	if _, err := os.Stat(encPath); err != nil {
		t.Fatalf("encrypted file missing: %v", err)
	}
	entries, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Body != "Convert me." {
		t.Errorf("unexpected entries after encrypt: %v", entries)
	}
	_, err = fj.ReencryptAll(false, "")
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	fj3 := NewFolderJournal(dir, testOpts)
	entries, err = fj3.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Body != "Convert me." {
		t.Errorf("unexpected entries after decrypt: %v", entries)
	}
}

func TestEncryptedWrongPassphrase(t *testing.T) {
	dir := t.TempDir()
	opts := testOpts
	opts.Encrypt = true
	opts.Passphrase = "correct"

	fj := NewFolderJournal(dir, opts)
	if err := fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Secret.", false); err != nil {
		t.Fatal(err)
	}
	opts.Passphrase = "wrong"
	fj2 := NewFolderJournal(dir, opts)
	_, err := fj2.Entries(&Filter{})
	if err == nil {
		t.Fatal("expected error with wrong passphrase")
	}
}

func TestLoadDayFileSingle(t *testing.T) {
	dir := t.TempDir()
	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nToday's entry.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")
	content2 := "# 2026-03-28 Saturday\n\n## [10:00 AM]\n\nYesterday.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local), content2, "md")

	fj := NewFolderJournal(dir, testOpts)
	d, err := fj.loadDayFile(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("loadDayFile failed: %v", err)
	}
	if len(d.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(d.entries))
	}
	if d.entries[0].Body != "Today's entry." {
		t.Errorf("body = %q", d.entries[0].Body)
	}
}

func TestLoadDayFileMissing(t *testing.T) {
	dir := t.TempDir()
	fj := NewFolderJournal(dir, testOpts)
	d, err := fj.loadDayFile(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("loadDayFile on missing file should succeed, got: %v", err)
	}
	if len(d.entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(d.entries))
	}
}

func TestLoadDayFileEncrypted(t *testing.T) {
	dir := t.TempDir()
	opts := testOpts
	opts.Encrypt = true
	opts.Passphrase = "testpass"
	fj := NewFolderJournal(dir, opts)
	if err := fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Secret.", false); err != nil {
		t.Fatal(err)
	}
	fj2 := NewFolderJournal(dir, opts)
	d, err := fj2.loadDayFile(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("loadDayFile encrypted failed: %v", err)
	}
	if len(d.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(d.entries))
	}
	if d.entries[0].Body != "Secret." {
		t.Errorf("body = %q", d.entries[0].Body)
	}
}

func TestAddEntryRoundTrip(t *testing.T) {
	dir := t.TempDir()
	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMorning entry.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.AddEntry(time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local), "Afternoon entry.", false); err != nil {
		t.Fatal(err)
	}
	fj2 := NewFolderJournal(dir, testOpts)
	entries, err := fj2.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Body != "Morning entry." {
		t.Errorf("entry 0 body = %q", entries[0].Body)
	}
	if entries[1].Body != "Afternoon entry." {
		t.Errorf("entry 1 body = %q", entries[1].Body)
	}
}

func TestChangeEntryTimesCrossDay(t *testing.T) {
	dir := t.TempDir()
	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMoving entry.\n\n## [10:00 AM]\n\nStaying entry.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	entries, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	newTime := time.Date(2026, 3, 28, 15, 0, 0, 0, time.Local)
	updated := entries[0]
	updated.Date = newTime
	if err := fj.UpdateEntry(entries[0], updated); err != nil {
		t.Fatal(err)
	}
	all, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all[0].Date.Day() != 28 || all[0].Body != "Moving entry." {
		t.Errorf("moved entry wrong: day=%d body=%q", all[0].Date.Day(), all[0].Body)
	}
	if all[1].Date.Day() != 29 {
		t.Errorf("stayed entry day = %d, want 29", all[1].Date.Day())
	}
	newPath := filepath.Join(dir, "2026", "03", "28.md")
	data, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("new day file not created: %v", err)
	}
	if !strings.Contains(string(data), "Moving entry.") {
		t.Error("entry not in new day file")
	}
}

func TestEncrypted(t *testing.T) {
	dir := t.TempDir()

	fj := NewFolderJournal(dir, testOpts)
	if fj.Encrypted() {
		t.Error("expected Encrypted() = false for plain journal")
	}

	encOpts := testOpts
	encOpts.Encrypt = true
	encOpts.Passphrase = "test"
	fj2 := NewFolderJournal(dir, encOpts)
	if !fj2.Encrypted() {
		t.Error("expected Encrypted() = true for encrypted journal")
	}
}

func TestImportEntry_AddsNewEntry(t *testing.T) {
	dir := t.TempDir()
	fj := NewFolderJournal(dir, testOpts)
	date := time.Date(2026, 1, 10, 9, 0, 0, 0, time.Local)
	added, err := fj.ImportEntry(Entry{Date: date, Body: "New import."})
	if err != nil {
		t.Fatalf("ImportEntry failed: %v", err)
	}
	if !added {
		t.Error("expected added=true for new entry")
	}
	entries, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Body != "New import." {
		t.Errorf("unexpected body: %q", entries[0].Body)
	}
}

func TestImportEntry_SkipsDuplicate(t *testing.T) {
	dir := t.TempDir()
	date := time.Date(2026, 1, 10, 9, 0, 0, 0, time.Local)
	writeDayFile(t, dir, date, "# 2026-01-10 Saturday\n\n## [09:00 AM]\n\nExisting entry.\n", "md")

	fj := NewFolderJournal(dir, testOpts)
	added, err := fj.ImportEntry(Entry{Date: date, Body: "Duplicate entry."})
	if err != nil {
		t.Fatalf("ImportEntry failed: %v", err)
	}
	if added {
		t.Error("expected added=false for duplicate timestamp")
	}
	entries, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Body != "Existing entry." {
		t.Errorf("expected original body, got %q", entries[0].Body)
	}
}

func TestImportEntry_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	fj := NewFolderJournal(dir, testOpts)
	d1 := time.Date(2026, 1, 10, 9, 0, 0, 0, time.Local)
	d2 := time.Date(2026, 1, 11, 15, 0, 0, 0, time.Local)
	added1, err := fj.ImportEntry(Entry{Date: d1, Body: "Day one entry."})
	if err != nil {
		t.Fatalf("ImportEntry d1: %v", err)
	}
	added2, err := fj.ImportEntry(Entry{Date: d2, Body: "Day two entry."})
	if err != nil {
		t.Fatalf("ImportEntry d2: %v", err)
	}
	if !added1 || !added2 {
		t.Errorf("expected both added, got added1=%v added2=%v", added1, added2)
	}
	entries, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestLoadDayFile(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nHello world.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)

	t.Run("existing file", func(t *testing.T) {
		d, err := fj.loadDayFile(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local))
		if err != nil {
			t.Fatal(err)
		}
		if len(d.entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(d.entries))
		}
		if d.entries[0].Body != "Hello world." {
			t.Errorf("body = %q", d.entries[0].Body)
		}
	})

	t.Run("missing file returns empty day", func(t *testing.T) {
		d, err := fj.loadDayFile(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
		if err != nil {
			t.Fatal(err)
		}
		if len(d.entries) != 0 {
			t.Fatalf("expected 0 entries, got %d", len(d.entries))
		}
	})
}

func TestWriteDay(t *testing.T) {
	dir := t.TempDir()
	fj := NewFolderJournal(dir, testOpts)

	t.Run("writes day file", func(t *testing.T) {
		d := &day{
			date:    time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local),
			entries: []Entry{{Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), Body: "Written."}},
		}
		if err := fj.writeDay(d); err != nil {
			t.Fatal(err)
		}

		path := filepath.Join(dir, "2026", "03", "29.md")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "Written.") {
			t.Errorf("file missing entry body, got: %s", data)
		}
	})

	t.Run("empty day deletes file and parent dirs", func(t *testing.T) {
		d := &day{
			date:    time.Date(2026, 5, 10, 0, 0, 0, 0, time.Local),
			entries: []Entry{{Date: time.Date(2026, 5, 10, 9, 0, 0, 0, time.Local), Body: "Temp."}},
		}
		if err := fj.writeDay(d); err != nil {
			t.Fatal(err)
		}

		path := filepath.Join(dir, "2026", "05", "10.md")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("file should exist: %v", err)
		}

		d.entries = nil
		if err := fj.writeDay(d); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Error("file should be deleted")
		}
		monthDir := filepath.Join(dir, "2026", "05")
		if _, err := os.Stat(monthDir); !errors.Is(err, os.ErrNotExist) {
			t.Error("empty month dir should be removed")
		}
	})
}

func TestListDayFiles(t *testing.T) {
	dir := t.TempDir()

	dates := []time.Time{
		time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local),
		time.Date(2026, 3, 20, 0, 0, 0, 0, time.Local),
		time.Date(2026, 4, 5, 0, 0, 0, 0, time.Local),
	}
	for _, d := range dates {
		content := fmt.Sprintf("# %s %s\n\n## [09:00 AM]\n\nEntry.\n",
			d.Format("2006-01-02"), d.Format("Monday"))
		writeDayFile(t, dir, d, content, "md")
	}

	fj := NewFolderJournal(dir, testOpts)

	t.Run("no date range returns all", func(t *testing.T) {
		files, err := fj.listDayFiles(nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 3 {
			t.Fatalf("expected 3 files, got %d", len(files))
		}
		if files[0].date.Day() != 15 || files[2].date.Month() != 4 {
			t.Errorf("wrong order: %v", files)
		}
	})

	t.Run("start date filters earlier files", func(t *testing.T) {
		start := time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local)
		files, err := fj.listDayFiles(&start, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 2 {
			t.Fatalf("expected 2 files, got %d", len(files))
		}
	})

	t.Run("end date filters later files", func(t *testing.T) {
		end := time.Date(2026, 3, 31, 23, 59, 59, 0, time.Local)
		files, err := fj.listDayFiles(nil, &end)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 2 {
			t.Fatalf("expected 2 files, got %d", len(files))
		}
	})

	t.Run("exact date", func(t *testing.T) {
		start := time.Date(2026, 3, 20, 0, 0, 0, 0, time.Local)
		end := time.Date(2026, 3, 20, 23, 59, 59, 0, time.Local)
		files, err := fj.listDayFiles(&start, &end)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files))
		}
	})

	t.Run("missing directory returns empty", func(t *testing.T) {
		fj2 := NewFolderJournal(filepath.Join(dir, "nonexistent"), testOpts)
		files, err := fj2.listDayFiles(nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != 0 {
			t.Fatalf("expected 0 files, got %d", len(files))
		}
	})
}

func TestEntries(t *testing.T) {
	dir := t.TempDir()

	content1 := "# 2026-03-01 Sunday\n\n## [09:00 AM]\n\nMarch first @work.\n\n## [02:00 PM] *\n\nAfternoon.\n"
	content2 := "# 2026-03-15 Sunday\n\n## [10:00 AM]\n\nMarch fifteenth @personal.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local), content1, "md")
	writeDayFile(t, dir, time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local), content2, "md")

	fj := NewFolderJournal(dir, testOpts)

	t.Run("no filter returns all sorted", func(t *testing.T) {
		f := &Filter{}
		entries, err := fj.Entries(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(entries))
		}
		if entries[0].Body != "March first @work." {
			t.Errorf("entry 0 body = %q", entries[0].Body)
		}
	})

	t.Run("date range filter", func(t *testing.T) {
		start := time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local)
		f := &Filter{StartDate: &start}
		entries, err := fj.Entries(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
	})

	t.Run("content filter", func(t *testing.T) {
		f := &Filter{Starred: true}
		entries, err := fj.Entries(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 starred entry, got %d", len(entries))
		}
	})

	t.Run("N limit", func(t *testing.T) {
		f := &Filter{N: 2}
		entries, err := fj.Entries(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}
		if entries[0].Body != "Afternoon." {
			t.Errorf("expected second entry, got %q", entries[0].Body)
		}
	})

	t.Run("missing dir returns empty", func(t *testing.T) {
		fj2 := NewFolderJournal(filepath.Join(dir, "nonexistent"), testOpts)
		f := &Filter{}
		entries, err := fj2.Entries(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("expected 0, got %d", len(entries))
		}
	})
}

func TestDayEntries(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nFirst.\n\n## [10:00 AM]\n\nSecond.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)

	entries, err := fj.DayEntries(time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	entries, err = fj.DayEntries(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestAddEntryImmediate(t *testing.T) {
	dir := t.TempDir()
	fj := NewFolderJournal(dir, testOpts)

	date := time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local)
	if err := fj.AddEntry(date, "Immediate save.", false); err != nil {
		t.Fatal(err)
	}

	// Verify file exists on disk without calling Save
	path := filepath.Join(dir, "2026", "03", "29.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if !strings.Contains(string(data), "Immediate save.") {
		t.Errorf("entry not in file: %s", data)
	}

	// Verify appending to existing day
	if err := fj.AddEntry(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local), "Second.", false); err != nil {
		t.Fatal(err)
	}

	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "Immediate save.") || !strings.Contains(content, "Second.") {
		t.Errorf("both entries should be present: %s", content)
	}
}

func TestDeleteEntrySingle(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nFirst.\n\n## [10:00 AM]\n\nSecond.\n\n## [11:00 AM]\n\nThird.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	f := &Filter{}
	entries, err := fj.Entries(f)
	if err != nil {
		t.Fatal(err)
	}

	if err := fj.DeleteEntry(entries[1]); err != nil {
		t.Fatal(err)
	}

	entries, err = fj.Entries(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Body != "First." || entries[1].Body != "Third." {
		t.Errorf("wrong entries remained: %q, %q", entries[0].Body, entries[1].Body)
	}
}

func TestDeleteEntryRemovesEmptyFile(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nOnly entry.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	f := &Filter{}
	entries, err := fj.Entries(f)
	if err != nil {
		t.Fatal(err)
	}

	if err := fj.DeleteEntry(entries[0]); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "2026", "03", "29.md")
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Error("file should be deleted when last entry is removed")
	}
}

func TestUpdateEntry(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nOriginal.\n\n## [10:00 AM]\n\nKeep me.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	f := &Filter{}
	entries, err := fj.Entries(f)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("same day update", func(t *testing.T) {
		updated := entries[0]
		updated.Body = "Modified."
		if err := fj.UpdateEntry(entries[0], updated); err != nil {
			t.Fatal(err)
		}

		result, err := fj.Entries(f)
		if err != nil {
			t.Fatal(err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(result))
		}
		if result[0].Body != "Modified." {
			t.Errorf("entry not updated: %q", result[0].Body)
		}
		if result[1].Body != "Keep me." {
			t.Errorf("other entry changed: %q", result[1].Body)
		}
	})
}

func TestUpdateEntryCrossDay(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMoving.\n\n## [10:00 AM]\n\nStaying.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	f := &Filter{}
	entries, err := fj.Entries(f)
	if err != nil {
		t.Fatal(err)
	}

	newDate := time.Date(2026, 3, 28, 15, 0, 0, 0, time.Local)
	updated := entries[0]
	updated.Date = newDate

	if err := fj.UpdateEntry(entries[0], updated); err != nil {
		t.Fatal(err)
	}

	result, err := fj.Entries(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0].Date.Day() != 28 || result[0].Body != "Moving." {
		t.Errorf("moved entry wrong: day=%d body=%q", result[0].Date.Day(), result[0].Body)
	}
	if result[1].Date.Day() != 29 || result[1].Body != "Staying." {
		t.Errorf("stayed entry wrong: day=%d body=%q", result[1].Date.Day(), result[1].Body)
	}

	if _, err := os.Stat(filepath.Join(dir, "2026", "03", "28.md")); err != nil {
		t.Errorf("new day file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "2026", "03", "29.md")); err != nil {
		t.Errorf("original day file missing: %v", err)
	}
}

func TestReencryptAll(t *testing.T) {
	dir := t.TempDir()

	content1 := "# 2026-03-28 Saturday\n\n## [09:00 AM]\n\nDay one.\n"
	content2 := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nDay two.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local), content1, "md")
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content2, "md")

	fj := NewFolderJournal(dir, testOpts)

	// Encrypt
	count, err := fj.ReencryptAll(true, "testpass")
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2 files, got %d", count)
	}

	// Plain files should be gone, encrypted files should exist
	if _, err := os.Stat(filepath.Join(dir, "2026", "03", "28.md")); !errors.Is(err, os.ErrNotExist) {
		t.Error("plain file should be deleted")
	}
	if _, err := os.Stat(filepath.Join(dir, "2026", "03", "28.md.age")); err != nil {
		t.Errorf("encrypted file missing: %v", err)
	}

	// Verify encrypted content is not plaintext
	data, _ := os.ReadFile(filepath.Join(dir, "2026", "03", "28.md.age"))
	if strings.Contains(string(data), "Day one.") {
		t.Error("encrypted file contains plaintext")
	}

	// Decrypt
	count, err = fj.ReencryptAll(false, "")
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2 files, got %d", count)
	}

	// Encrypted files should be gone, plain files should exist with correct content
	if _, err := os.Stat(filepath.Join(dir, "2026", "03", "28.md.age")); !errors.Is(err, os.ErrNotExist) {
		t.Error("encrypted file should be deleted")
	}
	data, err = os.ReadFile(filepath.Join(dir, "2026", "03", "28.md"))
	if err != nil {
		t.Fatalf("plain file missing: %v", err)
	}
	if !strings.Contains(string(data), "Day one.") {
		t.Errorf("content lost after round-trip: %s", data)
	}
}

func TestDeleteEntriesBatch(t *testing.T) {
	dir := t.TempDir()

	content1 := "# 2026-03-28 Saturday\n\n## [09:00 AM]\n\nDay one.\n"
	content2 := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nFirst.\n\n## [10:00 AM]\n\nSecond.\n\n## [11:00 AM]\n\nThird.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local), content1, "md")
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content2, "md")

	fj := NewFolderJournal(dir, testOpts)
	entries, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	// Delete from both days in one call: "Day one." (Mar 28) and "Second." (Mar 29)
	if err := fj.DeleteEntries([]Entry{entries[0], entries[2]}); err != nil {
		t.Fatal(err)
	}

	result, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result[0].Body != "First." {
		t.Errorf("entry 0 body = %q, want First.", result[0].Body)
	}
	if result[1].Body != "Third." {
		t.Errorf("entry 1 body = %q, want Third.", result[1].Body)
	}
}

func TestAddEntriesBatch(t *testing.T) {
	dir := t.TempDir()
	fj := NewFolderJournal(dir, testOpts)

	entries := []Entry{
		{Date: time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), Body: "First."},
		{Date: time.Date(2026, 3, 29, 10, 0, 0, 0, time.Local), Body: "Second."},
		{Date: time.Date(2026, 3, 28, 9, 0, 0, 0, time.Local), Body: "Other day."},
	}
	if err := fj.AddEntries(entries); err != nil {
		t.Fatal(err)
	}

	result, err := fj.Entries(&Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result))
	}
	if result[0].Body != "Other day." {
		t.Errorf("entry 0 body = %q, want Other day.", result[0].Body)
	}
}
