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
	tagParser *TagParser
}

func NewFolderJournal(path string, opts Options) *FolderJournal {
	return &FolderJournal{
		path:      path,
		opts:      opts,
		tagParser: NewTagParser(opts.TagSymbols),
	}
}

// AddEntry adds a new entry to the given date's day file and saves immediately.
func (fj *FolderJournal) AddEntry(date time.Time, body string, starred bool) error {
	d, err := fj.loadDayFile(date)
	if err != nil {
		return err
	}

	d.addEntry(body, starred, date)
	d.entries[len(d.entries)-1].Tags = fj.tagParser.Parse(body)

	return fj.writeDay(d)
}

// DeleteEntry removes a single entry (matched by timestamp and body) from
// its day file and saves immediately. Deletes the file if no entries remain.
func (fj *FolderJournal) DeleteEntry(e Entry) error {
	return fj.DeleteEntries([]Entry{e})
}

// UpdateEntry replaces old with updated in the day file. If the entry moves to a
// different calendar day, it is removed from the old day and added to the new.
func (fj *FolderJournal) UpdateEntry(old, updated Entry) error {
	oldKey := dateKeyFromTime(old.Date)
	newKey := dateKeyFromTime(updated.Date)

	if oldKey == newKey {
		d, err := fj.loadDayFile(old.Date)
		if err != nil {
			return err
		}
		for i, e := range d.entries {
			if e.Date.Equal(old.Date) && e.Body == old.Body {
				d.entries[i] = updated
				d.entries[i].Tags = fj.tagParser.Parse(updated.Body)
				break
			}
		}
		return fj.writeDay(d)
	}

	if err := fj.DeleteEntry(old); err != nil {
		return err
	}
	return fj.AddEntry(updated.Date, updated.Body, updated.Starred)
}

func groupByDay(entries []Entry) map[dateKey][]Entry {
	grouped := make(map[dateKey][]Entry)
	for _, e := range entries {
		key := dateKeyFromTime(e.Date)
		grouped[key] = append(grouped[key], e)
	}
	return grouped
}

// DeleteEntries removes multiple entries, grouped by day file for efficiency.
// Each affected day file is loaded and written at most once.
func (fj *FolderJournal) DeleteEntries(entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	for _, batch := range groupByDay(entries) {
		d, err := fj.loadDayFile(batch[0].Date)
		if err != nil {
			return err
		}

		// Build match counts: how many copies of each (date,body) to remove.
		type matchKey struct {
			unix int64
			body string
		}
		counts := make(map[matchKey]int, len(batch))
		for _, target := range batch {
			counts[matchKey{target.Date.Unix(), target.Body}]++
		}

		kept := d.entries[:0]
		for _, existing := range d.entries {
			mk := matchKey{existing.Date.Unix(), existing.Body}
			if counts[mk] > 0 {
				counts[mk]--
				continue
			}
			kept = append(kept, existing)
		}
		d.entries = kept

		if err := fj.writeDay(d); err != nil {
			return err
		}
	}
	return nil
}

// AddEntries adds multiple entries, grouped by day file for efficiency.
// Each affected day file is loaded and written at most once.
func (fj *FolderJournal) AddEntries(entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	for _, batch := range groupByDay(entries) {
		d, err := fj.loadDayFile(batch[0].Date)
		if err != nil {
			return err
		}
		for _, e := range batch {
			d.addEntry(e.Body, e.Starred, e.Date)
			d.entries[len(d.entries)-1].Tags = fj.tagParser.Parse(e.Body)
		}
		if err := fj.writeDay(d); err != nil {
			return err
		}
	}
	return nil
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

// Encrypted reports whether the journal uses encryption.
func (fj *FolderJournal) Encrypted() bool { return fj.opts.Encrypt }

// ReencryptAll walks all day files and re-writes them with the new encryption
// setting. Writes all new files before deleting any old files (safe on failure).
// Returns the number of files processed.
func (fj *FolderJournal) ReencryptAll(toEncrypt bool, newPassphrase string) (int, error) {
	files, err := fj.listDayFiles(nil, nil)
	if err != nil {
		return 0, err
	}

	newExt := "." + fj.opts.FileExt
	if toEncrypt {
		newExt = "." + fj.opts.FileExt + ".age"
	}

	type rewrite struct{ oldPath, newPath string }
	var rewrites []rewrite

	for _, fi := range files {
		data, err := os.ReadFile(fi.path)
		if err != nil {
			return 0, fmt.Errorf("reading %s: %w", fi.path, err)
		}

		if fj.opts.Encrypt {
			data, err = crypto.Decrypt(data, fj.opts.Passphrase)
			if err != nil {
				return 0, fmt.Errorf("decrypting %s: %w", fi.path, err)
			}
		}

		perm := os.FileMode(0644)
		if toEncrypt {
			data, err = crypto.Encrypt(data, newPassphrase)
			if err != nil {
				return 0, fmt.Errorf("encrypting: %w", err)
			}
			perm = 0600
		}

		newPath := filepath.Join(
			fj.path,
			fmt.Sprintf("%04d", fi.date.Year()),
			fmt.Sprintf("%02d", int(fi.date.Month())),
			fmt.Sprintf("%02d%s", fi.date.Day(), newExt),
		)

		dir := filepath.Dir(newPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return 0, err
		}

		if err := atomicfile.WriteFile(newPath, data, perm); err != nil {
			return 0, fmt.Errorf("writing %s: %w", newPath, err)
		}

		if fi.path != newPath {
			rewrites = append(rewrites, rewrite{fi.path, newPath})
		}
	}

	for _, r := range rewrites {
		os.Remove(r.oldPath)
	}

	fj.opts.Encrypt = toEncrypt
	fj.opts.Passphrase = newPassphrase

	return len(files), nil
}

// dayFileInfo holds a resolved path and parsed date for a single day file.
type dayFileInfo struct {
	path string
	date time.Time
}

// listDayFiles walks the YYYY/MM/ directory structure and returns a sorted
// slice of day files whose dates fall within [start, end] (inclusive). Either
// bound may be nil to indicate no limit.
func (fj *FolderJournal) listDayFiles(start, end *time.Time) ([]dayFileInfo, error) {
	plainExt := "." + fj.opts.FileExt
	encExt := plainExt + ".age"

	yearDirs, err := os.ReadDir(fj.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var files []dayFileInfo

	for _, yd := range yearDirs {
		if !yd.IsDir() {
			continue
		}
		year, err := strconv.Atoi(yd.Name())
		if err != nil || year < 1000 || year > 9999 {
			continue
		}
		if start != nil && year < start.Year() {
			continue
		}
		if end != nil && year > end.Year() {
			continue
		}

		monthDirs, err := os.ReadDir(filepath.Join(fj.path, yd.Name()))
		if err != nil {
			continue
		}

		for _, md := range monthDirs {
			if !md.IsDir() {
				continue
			}
			month, err := strconv.Atoi(md.Name())
			if err != nil || month < 1 || month > 12 {
				continue
			}
			if start != nil && year == start.Year() && time.Month(month) < start.Month() {
				continue
			}
			if end != nil && year == end.Year() && time.Month(month) > end.Month() {
				continue
			}

			dayEntries, err := os.ReadDir(filepath.Join(fj.path, yd.Name(), md.Name()))
			if err != nil {
				continue
			}

			for _, df := range dayEntries {
				if df.IsDir() {
					continue
				}

				name := df.Name()
				encrypted := strings.HasSuffix(name, encExt)
				if !encrypted && !strings.HasSuffix(name, plainExt) {
					continue
				}

				stem := name
				if encrypted {
					stem = strings.TrimSuffix(stem, encExt)
				} else {
					stem = strings.TrimSuffix(stem, plainExt)
				}
				dayNum, err := strconv.Atoi(stem)
				if err != nil || dayNum < 1 || dayNum > 31 {
					continue
				}

				date := time.Date(year, time.Month(month), dayNum, 0, 0, 0, 0, time.Local)

				if start != nil {
					startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
					if date.Before(startDay) {
						continue
					}
				}
				if end != nil {
					endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
					if date.After(endDay) {
						continue
					}
				}

				path := filepath.Join(fj.path, yd.Name(), md.Name(), name)
				files = append(files, dayFileInfo{path: path, date: date})
			}
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].date.Before(files[j].date)
	})

	return files, nil
}

// loadDayFile reads and parses a single day file. Returns an empty day
// if the file does not exist. Does not store state in fj.
func (fj *FolderJournal) loadDayFile(date time.Time) (*day, error) {
	path := fj.DayFilePath(date)

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &day{
			date: time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	if fj.opts.Encrypt {
		data, err = crypto.Decrypt(data, fj.opts.Passphrase)
		if err != nil {
			return nil, fmt.Errorf("decrypting %s: %w", path, err)
		}
	}

	parsed, err := parseDay(string(data), fj.opts.DateFmt, fj.opts.TimeFmt)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	for i := range parsed.entries {
		parsed.entries[i].Tags = fj.tagParser.Parse(parsed.entries[i].Body)
	}

	return &parsed, nil
}

// writeDay serializes a day to disk. If the day has no entries, the file
// is deleted and empty parent directories are cleaned up.
func (fj *FolderJournal) writeDay(d *day) error {
	if len(d.entries) == 0 {
		return fj.removeDayFile(d.date)
	}

	path := fj.DayFilePath(d.date)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	content := d.Format(fj.opts.DateFmt, fj.opts.TimeFmt)
	data := []byte(content)
	perm := os.FileMode(0644)

	if fj.opts.Encrypt {
		var err error
		data, err = crypto.Encrypt(data, fj.opts.Passphrase)
		if err != nil {
			return fmt.Errorf("encrypting: %w", err)
		}
		perm = 0600
	}

	return atomicfile.WriteFile(path, data, perm)
}

// removeDayFile deletes the day file for the given date and cleans up
// empty parent directories.
func (fj *FolderJournal) removeDayFile(date time.Time) error {
	path := fj.DayFilePath(date)
	err := os.Remove(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	fj.cleanEmptyParents(filepath.Dir(path))
	return nil
}

// cleanEmptyParents removes empty directories up to (but not including)
// the journal root.
func (fj *FolderJournal) cleanEmptyParents(dir string) {
	for dir != fj.path {
		if filepath.Dir(dir) == dir {
			break
		}
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		os.Remove(dir)
		dir = filepath.Dir(dir)
	}
}

// Entries loads day files matching the filter's date range, applies content
// filters, and returns matching entries sorted by date.
func (fj *FolderJournal) Entries(f *Filter) ([]Entry, error) {
	start, end := f.DateRange()
	files, err := fj.listDayFiles(start, end)
	if err != nil {
		return nil, err
	}

	var all []Entry
	for _, fi := range files {
		d, err := fj.loadDayFile(fi.date)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fi.path, err)
		}
		all = append(all, d.entries...)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Date.Before(all[j].Date)
	})

	return f.Apply(all), nil
}

// DayEntries returns all entries for a single calendar day.
func (fj *FolderJournal) DayEntries(date time.Time) ([]Entry, error) {
	d, err := fj.loadDayFile(date)
	if err != nil {
		return nil, err
	}
	return d.entries, nil
}

// ImportEntry adds e to the journal if no entry with the same timestamp exists.
// Returns true if added, false if duplicate. Saves immediately.
func (fj *FolderJournal) ImportEntry(e Entry) (bool, error) {
	d, err := fj.loadDayFile(e.Date)
	if err != nil {
		return false, err
	}

	for _, existing := range d.entries {
		if existing.Date.Equal(e.Date) {
			return false, nil
		}
	}

	d.addEntry(e.Body, e.Starred, e.Date)
	d.entries[len(d.entries)-1].Tags = fj.tagParser.Parse(e.Body)

	if err := fj.writeDay(d); err != nil {
		return false, err
	}

	return true, nil
}

