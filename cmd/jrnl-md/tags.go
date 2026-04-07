package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/spf13/cobra"
)

func newTagsCmd(rf *rootFlags) *cobra.Command {
	var f dateFlags

	cmd := &cobra.Command{
		Use:   "tags [date-filter-flags]",
		Short: "List all tags with frequency counts",
		Long: `List all tags with frequency counts, sorted descending.

Accepts the same date filter flags as list: --from, --to, --on, --year,
--month, --day, --today-in-history.`,
		Example: `  jrnl-md tags
  jrnl-md tags --year 2025
  jrnl-md tags --from "last month"`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTags(cmd, args, rf, &f)
		},
	}

	registerDateFlags(cmd, &f)

	return cmd
}

func runTags(cmd *cobra.Command, args []string, rf *rootFlags, f *dateFlags) error {
	cfg, err := loadConfig(rf)
	if err != nil {
		return err
	}

	now := time.Now()
	flt, err := parseDateFilter(f, now)
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
