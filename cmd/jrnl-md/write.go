package main

import (
	"fmt"
	"strings"

	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/spf13/cobra"
)

func newWriteCmd(rf *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "write <text>",
		Short:        "Append text to today's journal entry",
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
