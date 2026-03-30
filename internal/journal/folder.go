package journal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/jrnl-md/internal/atomicfile"
	"github.com/glw907/jrnl-md/internal/crypto"
)

type dateKey struct {
	year  int
	month time.Month
	day   int
}

func dateKeyFromTime(t time.Time) dateKey {
	return dateKey{t.Year(), t.Month(), t.Day()}
}

// Options configures journal format and storage behavior.
type Options struct {
	DateFmt    string
	TimeFmt    string
	TagSymbols string
	FileExt    string
	Encrypt    bool
	Passphrase string
}

// FolderJournal manages a journal stored as day files in a YYYY/MM/DD
// directory hierarchy.
type FolderJournal struct {
	path      string
	opts      Options
	days      map[dateKey]*day
	tagParser *TagParser
}

func NewFolderJournal(path string, opts Options) *FolderJournal {
	return &FolderJournal{
		path:      path,
		opts:      opts,
		days:      make(map[dateKey]*day),
		tagParser: NewTagParser(opts.TagSymbols),
	}
}

// Load reads all day files from disk. If the journal directory does not
// exist, Load succeeds with an empty journal.
func (fj *FolderJournal) Load() error {
	plainExt := "." + fj.opts.FileExt
	encExt := plainExt + ".age"

	return filepath.WalkDir(fj.path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if path == fj.path && errors.Is(err, os.ErrNotExist) {
				return filepath.SkipAll
			}
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		encrypted := strings.HasSuffix(name, encExt)
		if !encrypted && !strings.HasSuffix(name, plainExt) {
			return nil
		}

		rel, err := filepath.Rel(fj.path, path)
		if err != nil {
			return nil
		}
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) != 3 {
			return nil
		}

		year, err := strconv.Atoi(parts[0])
		if err != nil || year < 1000 || year > 9999 {
			return nil
		}
		month, err := strconv.Atoi(parts[1])
		if err != nil || month < 1 || month > 12 {
			return nil
		}

		stem := parts[2]
		if encrypted {
			stem = strings.TrimSuffix(stem, encExt)
		} else {
			stem = strings.TrimSuffix(stem, plainExt)
		}
		dayNum, err := strconv.Atoi(stem)
		if err != nil || dayNum < 1 || dayNum > 31 {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		if encrypted {
			data, err = crypto.Decrypt(data, fj.opts.Passphrase)
			if err != nil {
				return fmt.Errorf("decrypting %s: %w", path, err)
			}
		}

		parsed, err := parseDay(string(data), fj.opts.DateFmt, fj.opts.TimeFmt)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		for i := range parsed.entries {
			parsed.entries[i].Tags = fj.tagParser.Parse(parsed.entries[i].Body)
		}

		key := dateKey{year, time.Month(month), dayNum}
		fj.days[key] = &parsed

		return nil
	})
}

// LoadDay reads and parses only the day file for the given date. If the
// file does not exist, LoadDay succeeds with no entries for that day.
func (fj *FolderJournal) LoadDay(date time.Time) error {
	plainExt := "." + fj.opts.FileExt
	encExt := plainExt + ".age"

	base := filepath.Join(
		fj.path,
		fmt.Sprintf("%04d", date.Year()),
		fmt.Sprintf("%02d", int(date.Month())),
		fmt.Sprintf("%02d", date.Day()),
	)

	var path string
	var encrypted bool

	if fj.opts.Encrypt {
		path = base + encExt
		encrypted = true
	} else {
		path = base + plainExt
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	if encrypted {
		data, err = crypto.Decrypt(data, fj.opts.Passphrase)
		if err != nil {
			return fmt.Errorf("decrypting %s: %w", path, err)
		}
	}

	parsed, err := parseDay(string(data), fj.opts.DateFmt, fj.opts.TimeFmt)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	for i := range parsed.entries {
		parsed.entries[i].Tags = fj.tagParser.Parse(parsed.entries[i].Body)
	}

	key := dateKeyFromTime(date)
	fj.days[key] = &parsed

	return nil
}

// Save writes all modified day files to disk atomically.
func (fj *FolderJournal) Save() error {
	for key, d := range fj.days {
		if !d.modified {
			continue
		}

		ext := fj.opts.FileExt
		perm := os.FileMode(0644)
		if fj.opts.Encrypt {
			ext = fj.opts.FileExt + ".age"
			perm = 0600
		}

		path := filepath.Join(fj.path,
			fmt.Sprintf("%04d", key.year),
			fmt.Sprintf("%02d", int(key.month)),
			fmt.Sprintf("%02d.%s", key.day, ext),
		)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}

		content := d.Format(fj.opts.DateFmt, fj.opts.TimeFmt)
		data := []byte(content)

		if fj.opts.Encrypt {
			var err error
			data, err = crypto.Encrypt(data, fj.opts.Passphrase)
			if err != nil {
				return fmt.Errorf("encrypting %s: %w", path, err)
			}
		}

		if err := atomicfile.WriteFile(path, data, perm); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		d.modified = false
	}
	return nil
}

// AllEntries returns all entries across all days, sorted by date.
func (fj *FolderJournal) AllEntries() []Entry {
	n := 0
	for _, d := range fj.days {
		n += len(d.entries)
	}
	entries := make([]Entry, 0, n)
	for _, d := range fj.days {
		entries = append(entries, d.entries...)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.Before(entries[j].Date)
	})
	return entries
}

// AddEntry adds a new entry to the appropriate day, creating it if needed.
func (fj *FolderJournal) AddEntry(date time.Time, body string, starred bool) {
	key := dateKeyFromTime(date)
	d, ok := fj.days[key]
	if !ok {
		d = &day{
			date: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local),
		}
		fj.days[key] = d
	}

	d.addEntry(body, starred, date)
	d.entries[len(d.entries)-1].Tags = fj.tagParser.Parse(body)
}

// DayFilePath returns the expected file path for a given date.
func (fj *FolderJournal) DayFilePath(date time.Time) string {
	ext := fj.opts.FileExt
	if fj.opts.Encrypt {
		ext = fj.opts.FileExt + ".age"
	}
	return filepath.Join(
		fj.path,
		date.Format("2006"),
		date.Format("01"),
		date.Format("02")+"."+ext,
	)
}

// DeleteEntries removes entries matching by timestamp and marks affected
// days as modified.
func (fj *FolderJournal) DeleteEntries(entries []Entry) {
	if len(entries) == 0 {
		return
	}

	toDelete := make(map[time.Time]bool, len(entries))
	for _, e := range entries {
		toDelete[e.Date] = true
	}

	for _, d := range fj.days {
		var kept []Entry
		changed := false
		for _, e := range d.entries {
			if toDelete[e.Date] {
				changed = true
			} else {
				kept = append(kept, e)
			}
		}
		if changed {
			d.entries = kept
			d.modified = true
		}
	}
}

// ChangeEntryTimes updates the timestamp of matching entries. If an entry
// moves to a different calendar day, it is relocated to the target day.
func (fj *FolderJournal) ChangeEntryTimes(entries []Entry, newTime time.Time) {
	if len(entries) == 0 {
		return
	}

	toChange := make(map[time.Time]bool, len(entries))
	for _, e := range entries {
		toChange[e.Date] = true
	}

	newKey := dateKeyFromTime(newTime)

	keys := make([]dateKey, 0, len(fj.days))
	for k := range fj.days {
		keys = append(keys, k)
	}

	for _, k := range keys {
		d := fj.days[k]
		var kept []Entry
		changed := false
		for _, e := range d.entries {
			if !toChange[e.Date] {
				kept = append(kept, e)
				continue
			}
			changed = true
			oldKey := dateKeyFromTime(e.Date)
			e.Date = newTime
			if oldKey == newKey {
				kept = append(kept, e)
			} else {
				target, ok := fj.days[newKey]
				if !ok {
					target = &day{
						date: time.Date(newTime.Year(), newTime.Month(), newTime.Day(), 0, 0, 0, 0, time.Local),
					}
					fj.days[newKey] = target
				}
				target.entries = append(target.entries, e)
				target.modified = true
			}
		}
		if changed {
			d.entries = kept
			d.modified = true
		}
	}
}

// SetEncryption changes the encryption settings for future saves.
func (fj *FolderJournal) SetEncryption(encrypt bool, passphrase string) {
	fj.opts.Encrypt = encrypt
	fj.opts.Passphrase = passphrase
}

// MarkAllModified marks every loaded day as modified so Save writes all files.
func (fj *FolderJournal) MarkAllModified() {
	for _, d := range fj.days {
		d.modified = true
	}
}

// DayFiles returns all day file paths on disk (plain and encrypted).
func (fj *FolderJournal) DayFiles() ([]string, error) {
	plainExt := "." + fj.opts.FileExt
	encExt := plainExt + ".age"
	var paths []string

	err := filepath.WalkDir(fj.path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, encExt) || strings.HasSuffix(name, plainExt) {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}
