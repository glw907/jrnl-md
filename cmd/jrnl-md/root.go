package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/journal"
)

var version = "0.1.0"

type flags struct {
	n          int
	short      bool
	starred    bool
	and        bool
	not        []string
	notStarred bool
	notTagged  bool
	edit       bool
	delete     bool
	encrypt    bool
	decrypt    bool
	changeTime string
	from       string
	to         string
	on         string
	contains   string
	export     string
	format     string
	file       string
	list       bool
	tags       bool
	version    bool
	configFile string
}

func newRootCmd() *cobra.Command {
	var f flags

	cmd := &cobra.Command{
		Use:          "jrnl-md [journal:] [text...]",
		Short:        "A markdown-native journal for the command line",
		Long:         "jrnl-md is a journaling CLI that stores entries as markdown day files.",
		Args:         cobra.ArbitraryArgs,
		RunE:         func(cmd *cobra.Command, args []string) error { return runRoot(cmd, args, &f) },
		SilenceUsage: true,
	}

	cmd.Flags().IntVarP(&f.n, "num", "n", 0, "Show last N entries")
	cmd.Flags().BoolVarP(&f.short, "short", "s", false, "Show short entry list")
	cmd.Flags().BoolVar(&f.starred, "starred", false, "Show only starred entries")
	cmd.Flags().BoolVar(&f.and, "and", false, "Require all specified tags (AND logic)")
	cmd.Flags().StringArrayVar(&f.not, "not", nil, "Exclude entries containing tag")
	cmd.Flags().BoolVar(&f.notStarred, "not-starred", false, "Exclude starred entries")
	cmd.Flags().BoolVar(&f.notTagged, "not-tagged", false, "Exclude entries that have any tags")
	cmd.Flags().BoolVar(&f.edit, "edit", false, "Open entries in editor")
	cmd.Flags().BoolVar(&f.delete, "delete", false, "Delete entries")
	cmd.Flags().BoolVar(&f.encrypt, "encrypt", false, "Encrypt the journal")
	cmd.Flags().BoolVar(&f.decrypt, "decrypt", false, "Decrypt the journal")
	cmd.Flags().StringVar(&f.changeTime, "change-time", "", "Change entry time")
	cmd.Flag("change-time").NoOptDefVal = "now"
	cmd.Flags().StringVar(&f.from, "from", "", "Show entries from date")
	cmd.Flags().StringVar(&f.to, "to", "", "Show entries to date")
	cmd.Flags().StringVar(&f.on, "on", "", "Show entries on date")
	cmd.Flags().StringVar(&f.contains, "contains", "", "Filter entries containing text")
	cmd.Flags().StringVar(&f.export, "export", "", "Export format (json, md, txt, xml, yaml)")
	cmd.Flags().StringVar(&f.format, "format", "", "Export format (alias for --export)")
	cmd.Flags().StringVar(&f.file, "file", "", "Write export output to file instead of stdout")
	cmd.Flags().BoolVar(&f.list, "list", false, "List configured journals")
	cmd.Flags().BoolVar(&f.tags, "tags", false, "List all tags")
	cmd.Flags().BoolVarP(&f.version, "version", "v", false, "Show version")
	cmd.Flags().StringVar(&f.configFile, "config", "", "Config file path")

	cmd.AddCommand(newCompletionCmd())

	return cmd
}

func runRoot(cmd *cobra.Command, args []string, f *flags) error {
	if f.version {
		fmt.Printf("jrnl-md %s\n", version)
		return nil
	}

	configPath := f.configFile
	if configPath == "" {
		var err error
		configPath, err = config.DefaultPath()
		if err != nil {
			return err
		}
	}

	cfg, err := config.Load(configPath)
	if errors.Is(err, os.ErrNotExist) {
		cfg = config.Default()
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("determining home directory: %w", err)
		}
		cfg.Journals["default"] = config.JournalConfig{
			Path: home + "/Documents/Journal/",
		}
		if err := config.Save(cfg, configPath); err != nil {
			return fmt.Errorf("saving default config: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Created config at %s\n", configPath)
	} else if err != nil {
		return fmt.Errorf("loading config %s: %w", configPath, err)
	}

	if f.list {
		return listJournals(cfg)
	}

	if f.format != "" && f.export == "" {
		f.export = f.format
	}

	journalName, text, tagArgs := parseArgs(args, cfg)

	journalCfg, ok := cfg.Journals[journalName]
	if !ok {
		return fmt.Errorf("journal %q not found in config", journalName)
	}

	path, err := expandPath(journalCfg.Path)
	if err != nil {
		return fmt.Errorf("expanding path for journal %q: %w", journalName, err)
	}

	encrypted := journalEncrypted(journalCfg, cfg)

	if f.encrypt {
		return encryptJournal(path, journalName, cfg, configPath)
	}
	if f.decrypt {
		return decryptJournal(path, journalName, cfg, configPath)
	}

	var passphrase string
	if encrypted {
		passphrase, err = promptPassphrase(fmt.Sprintf("Passphrase for journal %q: ", journalName))
		if err != nil {
			return err
		}
	}

	opts := journalOptions(cfg, encrypted, passphrase)
	now := time.Now()

	// Read from stdin when not a terminal and no text args were provided and no filter flags set.
	if len(text) == 0 && !hasFilterFlags(f) && !f.edit && !term.IsTerminal(int(os.Stdin.Fd())) {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		body := strings.TrimSpace(string(data))
		if body != "" {
			text = []string{body}
		}
	}

	if len(text) > 0 {
		fj := journal.NewFolderJournal(path, opts)
		if err := fj.LoadDay(now); err != nil {
			return fmt.Errorf("loading journal: %w", err)
		}
		return writeInline(fj, text, cfg, now)
	}

	if f.edit || (len(args) == 0 && !hasFilterFlags(f)) {
		fj := journal.NewFolderJournal(path, opts)
		if err := fj.LoadDay(now); err != nil {
			return fmt.Errorf("loading journal: %w", err)
		}
		return editEntry(fj, cfg, configPath, passphrase)
	}

	fj := journal.NewFolderJournal(path, opts)
	if err := fj.Load(); err != nil {
		return fmt.Errorf("loading journal: %w", err)
	}

	if f.delete {
		return deleteEntries(fj, cfg, f, tagArgs)
	}
	if f.changeTime != "" {
		return changeTime(fj, cfg, f, tagArgs)
	}
	return readEntries(fj, cfg, f, tagArgs)
}

func listJournals(cfg config.Config) error {
	fmt.Println("Journals:")
	for name, j := range cfg.Journals {
		fmt.Printf("  %s -> %s\n", name, j.Path)
	}
	return nil
}

func hasFilterFlags(f *flags) bool {
	return f.n > 0 || f.short || f.starred || f.delete || f.encrypt || f.decrypt ||
		f.changeTime != "" || f.from != "" || f.to != "" || f.on != "" ||
		f.contains != "" || f.tags || f.export != "" ||
		f.notStarred || f.notTagged || len(f.not) > 0
}

func journalEncrypted(jcfg config.JournalConfig, cfg config.Config) bool {
	if jcfg.Encrypt != nil {
		return *jcfg.Encrypt
	}
	return cfg.General.Encrypt
}

func journalOptions(cfg config.Config, encrypted bool, passphrase string) journal.Options {
	return journal.Options{
		DateFmt:    cfg.Format.Date,
		TimeFmt:    cfg.Format.Time,
		TagSymbols: cfg.Format.TagSymbols,
		FileExt:    cfg.Format.FileExtension,
		Encrypt:    encrypted,
		Passphrase: passphrase,
	}
}

func boolPtr(v bool) *bool { return &v }
