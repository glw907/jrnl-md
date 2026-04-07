package main

import (
	"fmt"
	"time"

	"github.com/glw907/jrnl-md/internal/dateparse"
	"github.com/glw907/jrnl-md/internal/display"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/spf13/cobra"
)

type listFlags struct {
	n              int
	all            bool
	short          bool
	from           string
	to             string
	on             string
	year           int
	month          int
	day            int
	todayInHistory bool
	and            bool
	not            []string
	contains       string
}

func newListCmd(rf *rootFlags) *cobra.Command {
	var f listFlags

	cmd := &cobra.Command{
		Use:          "list [@tag...]",
		Short:        "List journal days matching filters",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, args, rf, &f)
		},
	}

	cmd.Flags().IntVarP(&f.n, "n", "n", 0, "last N days (default: default_list_count)")
	cmd.Flags().BoolVar(&f.all, "all", false, "show all days")
	cmd.Flags().BoolVar(&f.short, "short", false, "one-line summary per day")
	cmd.Flags().StringVar(&f.from, "from", "", "days from date")
	cmd.Flags().StringVar(&f.to, "to", "", "days up to date")
	cmd.Flags().StringVar(&f.on, "on", "", "single day")
	cmd.Flags().IntVar(&f.year, "year", 0, "all days in a year")
	cmd.Flags().IntVar(&f.month, "month", 0, "all days in a month (across years)")
	cmd.Flags().IntVar(&f.day, "day", 0, "all entries on a day-of-month")
	cmd.Flags().BoolVar(&f.todayInHistory, "today-in-history", false, "today's date in prior years")
	cmd.Flags().BoolVar(&f.and, "and", false, "require all specified tags")
	cmd.Flags().StringArrayVar(&f.not, "not", nil, "exclude days with tag")
	cmd.Flags().StringVar(&f.contains, "contains", "", "days containing text (case-insensitive)")

	return cmd
}

func runList(cmd *cobra.Command, args []string, rf *rootFlags, f *listFlags) error {
	cfg, err := loadConfig(rf)
	if err != nil {
		return err
	}

	now := time.Now()
	filter, err := buildFilter(args, f, cfg.General.DefaultListCount, now)
	if err != nil {
		return err
	}

	timeFmt := ""
	if cfg.General.Timestamps {
		timeFmt = cfg.Format.Time
	}
	s := journal.NewStore(cfg.JournalPath(), cfg.Format.Date, timeFmt, cfg.Format.TagSymbols)

	days, err := s.List(filter)
	if err != nil {
		return fmt.Errorf("listing days: %w", err)
	}

	if len(days) == 0 {
		return nil
	}

	if f.short {
		termWidth := display.TerminalWidth()
		for _, day := range days {
			fmt.Println(display.ShortLine(day, termWidth, timeFmt))
		}
		return nil
	}

	opts := display.FormatOpts{
		Linewrap:   cfg.General.Linewrap,
		DateColor:  cfg.Colors.Date,
		BodyColor:  cfg.Colors.Body,
		TagsColor:  cfg.Colors.Tags,
		TagSymbols: cfg.Format.TagSymbols,
	}
	fmt.Print(display.FormatDays(days, opts))
	return nil
}

func buildFilter(args []string, f *listFlags, defaultN int, now time.Time) (journal.Filter, error) {
	var flt journal.Filter

	for _, arg := range args {
		flt.Tags = append(flt.Tags, arg)
	}

	flt.AndTags = f.and
	flt.NotTags = f.not
	flt.Contains = f.contains
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

	if f.all {
		flt.N = 0
	} else if f.n > 0 {
		flt.N = f.n
	} else if !hasDateFilter(f) && len(flt.Tags) == 0 && flt.Contains == "" {
		flt.N = defaultN
	}

	return flt, nil
}

func hasDateFilter(f *listFlags) bool {
	return f.from != "" || f.to != "" || f.on != "" ||
		f.year != 0 || f.month != 0 || f.day != 0 || f.todayInHistory
}
