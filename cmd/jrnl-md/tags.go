package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/glw907/jrnl-md/internal/dateparse"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/spf13/cobra"
)

type tagsFlags struct {
	from           string
	to             string
	on             string
	year           int
	month          int
	day            int
	todayInHistory bool
}

func newTagsCmd(rf *rootFlags) *cobra.Command {
	var f tagsFlags

	cmd := &cobra.Command{
		Use:          "tags",
		Short:        "List all tags with frequency counts",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTags(cmd, args, rf, &f)
		},
	}

	cmd.Flags().StringVar(&f.from, "from", "", "days from date")
	cmd.Flags().StringVar(&f.to, "to", "", "days up to date")
	cmd.Flags().StringVar(&f.on, "on", "", "single day")
	cmd.Flags().IntVar(&f.year, "year", 0, "all days in a year")
	cmd.Flags().IntVar(&f.month, "month", 0, "all days in a month")
	cmd.Flags().IntVar(&f.day, "day", 0, "all entries on a day-of-month")
	cmd.Flags().BoolVar(&f.todayInHistory, "today-in-history", false, "today's date in prior years")

	return cmd
}

func runTags(cmd *cobra.Command, args []string, rf *rootFlags, f *tagsFlags) error {
	cfg, err := loadConfig(rf)
	if err != nil {
		return err
	}

	now := time.Now()
	flt, err := buildTagsFilter(f, now)
	if err != nil {
		return err
	}

	s := journal.NewStore(cfg.JournalPath(), cfg.Format.Date, "", cfg.Format.TagSymbols)
	tagCounts, err := s.Tags(flt)
	if err != nil {
		return fmt.Errorf("collecting tags: %w", err)
	}

	type tagCount struct {
		tag   string
		count int
	}
	var sorted []tagCount
	for tag, count := range tagCounts {
		sorted = append(sorted, tagCount{tag, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].count != sorted[j].count {
			return sorted[i].count > sorted[j].count
		}
		return sorted[i].tag < sorted[j].tag
	})

	for _, tc := range sorted {
		fmt.Printf("%s: %d\n", tc.tag, tc.count)
	}
	return nil
}

func buildTagsFilter(f *tagsFlags, now time.Time) (journal.Filter, error) {
	var flt journal.Filter
	flt.Year = f.year
	flt.Month = f.month
	flt.DayOfMonth = f.day
	flt.TodayInHistory = f.todayInHistory

	if f.from != "" {
		t, err := dateparse.Parse(f.from, now)
		if err != nil {
			return journal.Filter{}, fmt.Errorf("parsing --from: %w", err)
		}
		flt.Start = &t
	}
	if f.to != "" {
		t, err := dateparse.Parse(f.to, now)
		if err != nil {
			return journal.Filter{}, fmt.Errorf("parsing --to: %w", err)
		}
		flt.End = &t
	}
	if f.on != "" {
		t, err := dateparse.Parse(f.on, now)
		if err != nil {
			return journal.Filter{}, fmt.Errorf("parsing --on: %w", err)
		}
		flt.Start = &t
		flt.End = &t
	}
	return flt, nil
}
