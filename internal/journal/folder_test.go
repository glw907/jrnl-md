package journal

import (
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
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	entries := fj.AllEntries()
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
	fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Test entry.", false)

	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
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
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	fj.AddEntry(time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local), "Afternoon entry.", false)

	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	path := filepath.Join(dir, "2026", "03", "29.md")
	data, err := os.ReadFile(path)
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
	fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Test.", false)

	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
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
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	entries := fj.AllEntries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	fj.DeleteEntries([]Entry{entries[1]})

	remaining := fj.AllEntries()
	if len(remaining) != 2 {
		t.Fatalf("expected 2 entries after delete, got %d", len(remaining))
	}
	if remaining[0].Body != "First." {
		t.Errorf("entry 0 body = %q, want %q", remaining[0].Body, "First.")
	}
	if remaining[1].Body != "Third." {
		t.Errorf("entry 1 body = %q, want %q", remaining[1].Body, "Third.")
	}

	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
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
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	fj.DeleteEntries(nil)

	entries := fj.AllEntries()
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
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	entries := fj.AllEntries()
	fj.DeleteEntries([]Entry{entries[0]})

	remaining := fj.AllEntries()
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
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	entries := fj.AllEntries()
	newTime := time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local)
	fj.ChangeEntryTimes(entries, newTime)

	updated := fj.AllEntries()
	if len(updated) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(updated))
	}
	if !updated[0].Date.Equal(newTime) {
		t.Errorf("date = %v, want %v", updated[0].Date, newTime)
	}

	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
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
	fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Secret entry.", false)

	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
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
	if err := fj2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	entries := fj2.AllEntries()
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
	fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Convert me.", false)
	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load so LoadedPaths is populated, then encrypt and remove old plain files.
	fj = NewFolderJournal(dir, testOpts)
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	oldFiles := fj.LoadedPaths()

	fj.MarkAllModified()
	fj.SetEncryption(true, "pass123")
	if err := fj.Save(); err != nil {
		t.Fatalf("encrypted Save failed: %v", err)
	}

	for _, f := range oldFiles {
		os.Remove(f)
	}

	encPath := filepath.Join(dir, "2026", "03", "29.md.age")
	if _, err := os.Stat(encPath); err != nil {
		t.Fatalf("encrypted file missing: %v", err)
	}

	encOpts := testOpts
	encOpts.Encrypt = true
	encOpts.Passphrase = "pass123"
	fj2 := NewFolderJournal(dir, encOpts)
	if err := fj2.Load(); err != nil {
		t.Fatalf("encrypted Load failed: %v", err)
	}

	entries := fj2.AllEntries()
	if len(entries) != 1 || entries[0].Body != "Convert me." {
		t.Errorf("unexpected entries after encrypt: %v", entries)
	}

	encFiles := fj2.LoadedPaths()
	fj2.MarkAllModified()
	fj2.SetEncryption(false, "")
	if err := fj2.Save(); err != nil {
		t.Fatalf("decrypted Save failed: %v", err)
	}
	for _, f := range encFiles {
		os.Remove(f)
	}

	fj3 := NewFolderJournal(dir, testOpts)
	if err := fj3.Load(); err != nil {
		t.Fatalf("plaintext Load failed: %v", err)
	}
	entries = fj3.AllEntries()
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
	fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Secret.", false)
	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	opts.Passphrase = "wrong"
	fj2 := NewFolderJournal(dir, opts)
	if err := fj2.Load(); err == nil {
		t.Fatal("expected error with wrong passphrase")
	}
}

func TestLoadDay(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nToday's entry.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	// Also write another day that LoadDay should NOT load.
	content2 := "# 2026-03-28 Saturday\n\n## [10:00 AM]\n\nYesterday.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local), content2, "md")

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.LoadDay(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("LoadDay failed: %v", err)
	}

	entries := fj.AllEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Body != "Today's entry." {
		t.Errorf("body = %q", entries[0].Body)
	}
}

func TestLoadDayMissingFile(t *testing.T) {
	dir := t.TempDir()

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.LoadDay(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("LoadDay on missing file should succeed, got: %v", err)
	}

	entries := fj.AllEntries()
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestLoadDayEncrypted(t *testing.T) {
	dir := t.TempDir()

	opts := testOpts
	opts.Encrypt = true
	opts.Passphrase = "testpass"

	fj := NewFolderJournal(dir, opts)
	fj.AddEntry(time.Date(2026, 3, 29, 9, 0, 0, 0, time.Local), "Secret.", false)
	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	fj2 := NewFolderJournal(dir, opts)
	if err := fj2.LoadDay(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("LoadDay encrypted failed: %v", err)
	}

	entries := fj2.AllEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Body != "Secret." {
		t.Errorf("body = %q", entries[0].Body)
	}
}

func TestLoadDayAddEntrySaveRoundTrip(t *testing.T) {
	dir := t.TempDir()

	// Write an existing day file with one entry.
	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMorning entry.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	// LoadDay, add a second entry, save.
	fj := NewFolderJournal(dir, testOpts)
	if err := fj.LoadDay(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("LoadDay failed: %v", err)
	}

	fj.AddEntry(time.Date(2026, 3, 29, 14, 30, 0, 0, time.Local), "Afternoon entry.", false)

	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Reload with full Load and verify both entries are present.
	fj2 := NewFolderJournal(dir, testOpts)
	if err := fj2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	entries := fj2.AllEntries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Body != "Morning entry." {
		t.Errorf("entry 0 body = %q, want %q", entries[0].Body, "Morning entry.")
	}
	if entries[1].Body != "Afternoon entry." {
		t.Errorf("entry 1 body = %q, want %q", entries[1].Body, "Afternoon entry.")
	}
}

func TestChangeEntryTimesCrossDay(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nMoving entry.\n\n## [10:00 AM]\n\nStaying entry.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	entries := fj.AllEntries()
	newTime := time.Date(2026, 3, 28, 15, 0, 0, 0, time.Local)
	fj.ChangeEntryTimes([]Entry{entries[0]}, newTime)

	all := fj.AllEntries()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}

	if all[0].Date.Day() != 28 {
		t.Errorf("moved entry day = %d, want 28", all[0].Date.Day())
	}
	if all[0].Body != "Moving entry." {
		t.Errorf("moved entry body = %q", all[0].Body)
	}

	if all[1].Date.Day() != 29 {
		t.Errorf("stayed entry day = %d, want 29", all[1].Date.Day())
	}

	if err := fj.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
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

func TestLoadedPathsAfterLoad(t *testing.T) {
	dir := t.TempDir()

	content1 := "# 2026-03-28 Saturday\n\n## [09:00 AM]\n\nDay one.\n"
	content2 := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nDay two.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 28, 0, 0, 0, 0, time.Local), content1, "md")
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content2, "md")

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	paths := fj.LoadedPaths()
	if len(paths) != 2 {
		t.Fatalf("expected 2 loaded paths, got %d", len(paths))
	}
}

func TestLoadedPathsAfterLoadDay(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nToday.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.LoadDay(time.Date(2026, 3, 29, 14, 0, 0, 0, time.Local)); err != nil {
		t.Fatalf("LoadDay failed: %v", err)
	}

	paths := fj.LoadedPaths()
	if len(paths) != 1 {
		t.Fatalf("expected 1 loaded path, got %d", len(paths))
	}
}

func TestLoadedPathsMissingDir(t *testing.T) {
	dir := t.TempDir()

	fj := NewFolderJournal(filepath.Join(dir, "nonexistent"), testOpts)
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	paths := fj.LoadedPaths()
	if len(paths) != 0 {
		t.Fatalf("expected 0 loaded paths, got %d", len(paths))
	}
}

func TestReplaceEntries(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nOriginal entry.\n\n## [10:00 AM]\n\nStay unchanged.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	entries := fj.AllEntries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Replace only the first entry (09:00 AM) with a modified version.
	newEntry := Entry{
		Date:    entries[0].Date,
		Body:    "Replaced entry.",
		Starred: false,
	}
	fj.ReplaceEntries([]Entry{entries[0]}, []Entry{newEntry})

	result := fj.AllEntries()
	if len(result) != 2 {
		t.Fatalf("expected 2 entries after replace, got %d", len(result))
	}

	bodies := map[string]bool{}
	for _, e := range result {
		bodies[e.Body] = true
	}
	if !bodies["Replaced entry."] {
		t.Error("expected 'Replaced entry.' in journal")
	}
	if bodies["Original entry."] {
		t.Error("expected 'Original entry.' to be gone")
	}
	if !bodies["Stay unchanged."] {
		t.Error("expected 'Stay unchanged.' to remain")
	}
}

func TestReplaceEntriesDeleteAll(t *testing.T) {
	dir := t.TempDir()

	content := "# 2026-03-29 Sunday\n\n## [09:00 AM]\n\nDelete me.\n"
	writeDayFile(t, dir, time.Date(2026, 3, 29, 0, 0, 0, 0, time.Local), content, "md")

	fj := NewFolderJournal(dir, testOpts)
	if err := fj.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	entries := fj.AllEntries()
	fj.ReplaceEntries(entries, nil)

	result := fj.AllEntries()
	if len(result) != 0 {
		t.Fatalf("expected 0 entries after replace with nil, got %d", len(result))
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

	entries := fj.AllEntries()
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

	// Pre-populate a day file with an entry at the same timestamp
	writeDayFile(t, dir, date, "# 2026-01-10 Saturday\n\n## [09:00 AM]\n\nExisting entry.\n", "md")

	fj := NewFolderJournal(dir, testOpts)
	// ImportEntry must call LoadDay internally — do NOT call fj.Load() first

	added, err := fj.ImportEntry(Entry{Date: date, Body: "Duplicate entry."})
	if err != nil {
		t.Fatalf("ImportEntry failed: %v", err)
	}
	if added {
		t.Error("expected added=false for duplicate timestamp")
	}

	// Confirm original is untouched
	if err := fj.Load(); err != nil {
		t.Fatal(err)
	}
	entries := fj.AllEntries()
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

	entries := fj.AllEntries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
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
