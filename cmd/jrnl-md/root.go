package main

import (
	"fmt"

	"github.com/glw907/jrnl-md/internal/config"
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
