package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/dateparse"
	"github.com/glw907/jrnl-md/internal/journal"
)

// preprocessArgs converts -N numeric shorthand (e.g. -3) to -n N
// before cobra parses the flags.
func preprocessArgs(args []string) []string {
	var result []string
	for _, arg := range args {
		if len(arg) >= 2 && arg[0] == '-' && arg[1] >= '0' && arg[1] <= '9' {
			allDigits := true
			for _, c := range arg[1:] {
				if c < '0' || c > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				result = append(result, "-n", arg[1:])
				continue
			}
		}
		result = append(result, arg)
	}
	return result
}

func parseArgs(args []string, cfg config.Config) (journalName string, text []string, tagArgs []string) {
	if len(args) == 0 {
		return "default", nil, nil
	}

	first := args[0]
	if strings.HasSuffix(first, ":") {
		name := strings.TrimSuffix(first, ":")
		if _, ok := cfg.Journals[name]; ok {
			return name, args[1:], nil
		}
	}

	if len(cfg.Format.TagSymbols) > 0 {
		allTags := true
		var tags []string
		for _, arg := range args {
			if len(arg) > 1 && strings.ContainsRune(cfg.Format.TagSymbols, rune(arg[0])) {
				tags = append(tags, arg)
			} else {
				allTags = false
				break
			}
		}
		if allTags && len(tags) > 0 {
			return "default", nil, tags
		}
	}

	return "default", args, nil
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("determining home directory: %w", err)
		}
		return home + path[1:], nil
	}
	return path, nil
}

func buildFilter(f *flags, tagArgs []string) (journal.Filter, error) {
	var flt journal.Filter
	flt.N = f.n
	flt.Starred = f.starred
	flt.AndTags = f.and
	flt.NotTags = f.not
	flt.NotStarred = f.notStarred
	flt.NotTagged = f.notTagged

	if len(tagArgs) > 0 {
		flt.Tags = tagArgs
	}

	if f.contains != "" {
		flt.Contains = f.contains
	}

	if f.on != "" {
		start, err := dateparse.Parse(f.on)
		if err != nil {
			return flt, fmt.Errorf("parsing --on date: %w", err)
		}
		startOfDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
		endOfDay := time.Date(start.Year(), start.Month(), start.Day(), 23, 59, 59, 0, time.Local)
		flt.StartDate = &startOfDay
		flt.EndDate = &endOfDay
	}

	if f.from != "" {
		start, err := dateparse.Parse(f.from)
		if err != nil {
			return flt, fmt.Errorf("parsing --from date: %w", err)
		}
		flt.StartDate = &start
	}

	if f.to != "" {
		end, err := dateparse.ParseInclusive(f.to)
		if err != nil {
			return flt, fmt.Errorf("parsing --to date: %w", err)
		}
		flt.EndDate = &end
	}

	return flt, nil
}
