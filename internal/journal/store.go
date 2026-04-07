package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/jrnl-md/internal/fsutil"
)

// Store manages day files under a root directory.
type Store struct {
	root    string
	dateFmt string
	timeFmt string // empty string means timestamps disabled
	tagSyms string
}

// NewStore creates a Store backed by root.
func NewStore(root, dateFmt, timeFmt, tagSyms string) *Store {
	return &Store{
		root:    root,
		dateFmt: dateFmt,
		timeFmt: timeFmt,
		tagSyms: tagSyms,
	}
}

// DayPath returns the file path for a given date.
func (s *Store) DayPath(date time.Time) string {
	return filepath.Join(s.root,
		date.Format("2006"),
		date.Format("01"),
		date.Format("2006-01-02")+".md",
	)
}

// Load reads and parses the day file for date.
// Returns os.ErrNotExist if the file does not exist.
func (s *Store) Load(date time.Time) (Day, error) {
	path := s.DayPath(date)
	data, err := os.ReadFile(path)
	if err != nil {
		return Day{}, err
	}
	return parseDay(date, string(data))
}

// Save writes day to its day file using atomic write.
// Creates parent directories as needed.
func (s *Store) Save(day Day) error {
	path := s.DayPath(day.Date)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	content := formatDay(day)
	return fsutil.AtomicWrite(path, []byte(content), 0644)
}

// Delete removes the day file for date. Returns an error if the file
// does not exist. Removes parent directories if they become empty.
func (s *Store) Delete(date time.Time) error {
	path := s.DayPath(date)
	if err := os.Remove(path); err != nil {
		return err
	}
	monthDir := filepath.Dir(path)
	yearDir := filepath.Dir(monthDir)
	removeIfEmpty(monthDir)
	removeIfEmpty(yearDir)
	return nil
}

func removeIfEmpty(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return
	}
	os.Remove(dir)
}

// Append appends body to today's day file. Creates the file with a
// day heading if it doesn't exist. Adds a timestamp heading if
// timeFmt is non-empty.
func (s *Store) Append(body string) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var dayBody string
	existing, err := s.Load(today)
	if err == nil {
		dayBody = existing.Body
	}

	var sb strings.Builder
	sb.WriteString(dayBody)
	hasContent := strings.TrimSpace(dayBody) != ""

	if s.timeFmt != "" {
		sb.WriteString("\n")
		sb.WriteString("## ")
		sb.WriteString(now.Format(s.timeFmt))
		sb.WriteString("\n\n")
	} else if hasContent {
		sb.WriteString("\n")
	}

	sb.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		sb.WriteString("\n")
	}

	day := Day{Date: today, Body: sb.String()}
	return s.Save(day)
}

// List returns all days matching f, sorted newest-first.
// Uses directory structure to skip irrelevant years/months.
func (s *Store) List(f Filter) ([]Day, error) {
	var days []Day

	years, err := readDirNames(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing journal root: %w", err)
	}

	for _, yearStr := range years {
		y, err := strconv.Atoi(yearStr)
		if err != nil {
			continue
		}
		if f.Year != 0 && y != f.Year {
			continue
		}
		if f.Start != nil && y < f.Start.Year() {
			continue
		}
		if f.End != nil && y > f.End.Year() {
			continue
		}

		yearDir := filepath.Join(s.root, yearStr)
		months, err := readDirNames(yearDir)
		if err != nil {
			continue
		}

		for _, monthStr := range months {
			m, err := strconv.Atoi(monthStr)
			if err != nil {
				continue
			}
			if f.Month != 0 && m != f.Month {
				continue
			}

			monthDir := filepath.Join(yearDir, monthStr)
			entries, err := os.ReadDir(monthDir)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				name := entry.Name()
				if !strings.HasSuffix(name, ".md") {
					continue
				}
				base := strings.TrimSuffix(name, ".md")
				date, err := time.Parse("2006-01-02", base)
				if err != nil {
					continue
				}
				day, err := s.Load(date)
				if err != nil {
					continue
				}
				if f.Match(day, s.tagSyms) {
					days = append(days, day)
				}
			}
		}
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].Date.After(days[j].Date)
	})

	if f.N > 0 && len(days) > f.N {
		days = days[:f.N]
	}

	return days, nil
}

// Tags returns tag frequencies across all days matching f.
func (s *Store) Tags(f Filter) (map[string]int, error) {
	days, err := s.List(f)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, day := range days {
		tags := extractTags(day.Body, s.tagSyms)
		seen := make(map[string]bool)
		for _, tag := range tags {
			if !seen[tag] {
				counts[tag]++
				seen[tag] = true
			}
		}
	}
	return counts, nil
}

func readDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
