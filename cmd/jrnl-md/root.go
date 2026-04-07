package main

import (
	"fmt"
	"time"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/dateparse"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/spf13/cobra"
)

type rootFlags struct {
	configFile string
}

func newRootCmd() *cobra.Command {
	var f rootFlags

	cmd := &cobra.Command{
		Use:          "jrnl-md",
		Short:        "A markdown journaling CLI",
		SilenceUsage: true,
	}

	cmd.PersistentFlags().StringVar(&f.configFile, "config-file", "", "path to config file (default: ~/.config/jrnl-md/config.toml)")

	cmd.AddCommand(newWriteCmd(&f))
	cmd.AddCommand(newEditCmd(&f))
	cmd.AddCommand(newListCmd(&f))
	cmd.AddCommand(newTagsCmd(&f))
	cmd.AddCommand(newCompletionCmd())

	cmd.Version = "2.0.0"

	return cmd
}

func loadConfig(f *rootFlags) (config.Config, error) {
	path := f.configFile
	if path == "" {
		path = config.DefaultPath()
	}
	cfg, err := config.Load(path)
	if err != nil {
		return config.Config{}, fmt.Errorf("loading config: %w", err)
	}
	return cfg, nil
}

// dateFlags holds the date filter flags shared by list and tags subcommands.
type dateFlags struct {
	from           string
	to             string
	on             string
	year           int
	month          int
	day            int
	todayInHistory bool
}

// registerDateFlags adds date filter flags to a command.
func registerDateFlags(cmd *cobra.Command, f *dateFlags) {
	cmd.Flags().StringVar(&f.from, "from", "", "days from date")
	cmd.Flags().StringVar(&f.to, "to", "", "days up to date")
	cmd.Flags().StringVar(&f.on, "on", "", "single day")
	cmd.Flags().IntVar(&f.year, "year", 0, "all days in a year")
	cmd.Flags().IntVar(&f.month, "month", 0, "all days in a month")
	cmd.Flags().IntVar(&f.day, "day", 0, "all entries on a day-of-month")
	cmd.Flags().BoolVar(&f.todayInHistory, "today-in-history", false, "today's date in prior years")
}

// parseDateFilter builds the date-related portion of a journal.Filter.
func parseDateFilter(f *dateFlags, now time.Time) (journal.Filter, error) {
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

// hasDateFilter reports whether any date filter flag is set.
func (f *dateFlags) hasFilter() bool {
	return f.from != "" || f.to != "" || f.on != "" ||
		f.year != 0 || f.month != 0 || f.day != 0 || f.todayInHistory
}
