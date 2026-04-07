package main

import (
	"fmt"
	"strings"

	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/spf13/cobra"
)

func newWriteCmd(rf *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write <text>",
		Short: "Append text to today's day file",
		Long: `Append text to today's day file.

Creates the file with a day heading if it does not exist. With timestamps
enabled (the default), each write gets a ## time heading before the body.
With timestamps disabled, consecutive writes are separated by a blank line.

The text argument is everything after "write" on the command line.`,
		Example: `  jrnl-md write Went for a morning run. Feeling good.
  jrnl-md write Met with @sarah about the project timeline.`,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWrite(cmd, args, rf)
		},
	}
	return cmd
}

func runWrite(cmd *cobra.Command, args []string, rf *rootFlags) error {
	cfg, err := loadConfig(rf)
	if err != nil {
		return err
	}

	timeFmt := ""
	if cfg.General.Timestamps {
		timeFmt = cfg.Format.Time
	}

	s := journal.NewStore(cfg.JournalPath(), cfg.Format.Date, timeFmt, cfg.Format.TagSymbols)

	body := strings.Join(args, " ")
	if err := s.Append(body); err != nil {
		return fmt.Errorf("writing entry: %w", err)
	}
	return nil
}
