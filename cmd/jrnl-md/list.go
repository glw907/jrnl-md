package main

import (
	"fmt"
	"time"

	"github.com/glw907/jrnl-md/internal/display"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/spf13/cobra"
)

type listFlags struct {
	dateFlags
	n        int
	all      bool
	short    bool
	and      bool
	not      []string
	contains string
}

func newListCmd(rf *rootFlags) *cobra.Command {
	var f listFlags

	cmd := &cobra.Command{
		Use:   "list [flags] [@tag...]",
		Short: "Display day files matching filters",
		Long: `Display day files matching the given filters.

Defaults to the last default_list_count days (10 by default). Positional
@tag arguments filter to days containing those tags. Date arguments accept
natural language (yesterday, last monday, 3 days ago) as well as YYYY-MM-DD.

Body text is wrapped at the configured linewrap width. Tags are highlighted
in color when a color is configured.`,
		Example: `  jrnl-md list                          # last 10 days
  jrnl-md list -5                        # last 5 days
  jrnl-md list --all                     # all days
  jrnl-md list --short                   # one line per day
  jrnl-md list @work                     # days tagged @work
  jrnl-md list @work @sarah --and        # days tagged both
  jrnl-md list --from "last monday"
  jrnl-md list --on 2026-04-01
  jrnl-md list --year 2025
  jrnl-md list --today-in-history
  jrnl-md list --contains "budget question"`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, args, rf, &f)
		},
	}

	cmd.Flags().IntVarP(&f.n, "n", "n", 0, "last N days (default: default_list_count)")
	cmd.Flags().BoolVar(&f.all, "all", false, "show all days")
	cmd.Flags().BoolVar(&f.short, "short", false, "one-line summary per day")
	cmd.Flags().BoolVar(&f.and, "and", false, "require all specified tags")
	cmd.Flags().StringArrayVar(&f.not, "not", nil, "exclude days with tag")
	cmd.Flags().StringVar(&f.contains, "contains", "", "days containing text (case-insensitive)")
	registerDateFlags(cmd, &f.dateFlags)

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
	flt, err := parseDateFilter(&f.dateFlags, now)
	if err != nil {
		return journal.Filter{}, err
	}

	for _, arg := range args {
		flt.Tags = append(flt.Tags, arg)
	}

	flt.AndTags = f.and
	flt.NotTags = f.not
	flt.Contains = f.contains

	if f.all {
		flt.N = 0
	} else if f.n > 0 {
		flt.N = f.n
	} else if !f.dateFlags.hasFilter() && len(flt.Tags) == 0 && flt.Contains == "" {
		flt.N = defaultN
	}

	return flt, nil
}
