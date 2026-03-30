package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/glw907/jrnl-md/internal/atomicfile"
	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/crypto"
	"github.com/glw907/jrnl-md/internal/dateparse"
	"github.com/glw907/jrnl-md/internal/display"
	"github.com/glw907/jrnl-md/internal/editor"
	"github.com/glw907/jrnl-md/internal/export"
	"github.com/glw907/jrnl-md/internal/journal"
	"github.com/glw907/jrnl-md/internal/prompt"
)

var (
	version = "0.1.0"

	flagN          int
	flagShort      bool
	flagStarred    bool
	flagEdit       bool
	flagDelete     bool
	flagEncrypt    bool
	flagDecrypt    bool
	flagChangeTime string
	flagFrom       string
	flagTo         string
	flagOn         string
	flagContains   string
	flagExport     string
	flagList       bool
	flagTags       bool
	flagVersion    bool
	flagConfigFile string
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "jrnl-md [journal:] [text...]",
		Short:        "A markdown-native journal for the command line",
		Long:         "jrnl-md is a journaling CLI that stores entries as markdown day files.",
		Args:         cobra.ArbitraryArgs,
		RunE:         runRoot,
		SilenceUsage: true,
	}

	cmd.Flags().IntVarP(&flagN, "num", "n", 0, "Show last N entries")
	cmd.Flags().BoolVarP(&flagShort, "short", "s", false, "Show short entry list")
	cmd.Flags().BoolVar(&flagStarred, "starred", false, "Show only starred entries")
	cmd.Flags().BoolVar(&flagEdit, "edit", false, "Open entries in editor")
	cmd.Flags().BoolVar(&flagDelete, "delete", false, "Delete entries")
	cmd.Flags().BoolVar(&flagEncrypt, "encrypt", false, "Encrypt the journal")
	cmd.Flags().BoolVar(&flagDecrypt, "decrypt", false, "Decrypt the journal")
	cmd.Flags().StringVar(&flagChangeTime, "change-time", "", "Change entry time")
	cmd.Flag("change-time").NoOptDefVal = "now"
	cmd.Flags().StringVar(&flagFrom, "from", "", "Show entries from date")
	cmd.Flags().StringVar(&flagTo, "to", "", "Show entries to date")
	cmd.Flags().StringVar(&flagOn, "on", "", "Show entries on date")
	cmd.Flags().StringVar(&flagContains, "contains", "", "Filter entries containing text")
	cmd.Flags().StringVar(&flagExport, "export", "", "Export format (json, md, txt, xml, yaml)")
	cmd.Flags().BoolVar(&flagList, "list", false, "List configured journals")
	cmd.Flags().BoolVar(&flagTags, "tags", false, "List all tags")
	cmd.Flags().BoolVarP(&flagVersion, "version", "v", false, "Show version")
	cmd.Flags().StringVar(&flagConfigFile, "config", "", "Config file path")

	cmd.AddCommand(newCompletionCmd())

	return cmd
}

func runRoot(cmd *cobra.Command, args []string) error {
	if flagVersion {
		fmt.Printf("jrnl-md %s\n", version)
		return nil
	}

	configPath := flagConfigFile
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

	if flagList {
		return listJournals(cfg)
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

	if flagEncrypt {
		return encryptJournal(path, journalName, cfg, configPath)
	}
	if flagDecrypt {
		return decryptJournal(path, journalName, cfg, configPath)
	}

	var passphrase string
	if encrypted {
		passphrase, err = promptPassphrase(fmt.Sprintf("Passphrase for journal %q: ", journalName))
		if err != nil {
			return err
		}
	}

	fj := journal.NewFolderJournal(path, journalOptions(cfg, encrypted, passphrase))
	if err := fj.Load(); err != nil {
		return fmt.Errorf("loading journal: %w", err)
	}

	if flagDelete {
		return deleteEntries(fj, cfg, tagArgs)
	}

	if flagChangeTime != "" {
		return changeTime(fj, cfg, tagArgs)
	}

	if len(text) > 0 {
		return writeInline(fj, text, cfg)
	}

	if flagEdit || (len(args) == 0 && !hasFilterFlags(cmd)) {
		return editEntry(fj, cfg, encrypted, passphrase)
	}

	return readEntries(fj, cfg, tagArgs)
}

func writeInline(fj *journal.FolderJournal, text []string, cfg config.Config) error {
	body := strings.Join(text, " ")
	starred := strings.HasSuffix(body, "*") || strings.HasPrefix(body, "*")
	if starred {
		body = strings.Trim(body, "* ")
	}

	fj.AddEntry(time.Now(), body, starred)

	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Entry added.")
	return nil
}

func editEntry(fj *journal.FolderJournal, cfg config.Config, encrypted bool, passphrase string) error {
	if cfg.General.Editor == "" {
		msg := "no editor configured"
		if configPath, err := config.DefaultPath(); err == nil {
			msg = fmt.Sprintf("no editor configured. Set editor in %s", configPath)
		}
		return errors.New(msg)
	}

	now := time.Now()

	var tmpl string
	if cfg.General.Template != "" {
		tmplPath, err := expandPath(cfg.General.Template)
		if err != nil {
			return fmt.Errorf("expanding template path: %w", err)
		}
		data, err := os.ReadFile(tmplPath)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", tmplPath, err)
		}
		tmpl = string(data)
	}

	if encrypted {
		return editEncrypted(fj, cfg, now, passphrase, tmpl)
	}

	path := fj.DayFilePath(now)
	lineCount, err := editor.PrepareDayFile(path, now, cfg.Format.Date, cfg.Format.Time, tmpl)
	if err != nil {
		return fmt.Errorf("preparing day file: %w", err)
	}

	if err := editor.Launch(cfg.General.Editor, path, lineCount); err != nil {
		return fmt.Errorf("launching editor: %w", err)
	}

	return nil
}

func editEncrypted(fj *journal.FolderJournal, cfg config.Config, now time.Time, passphrase, tmpl string) error {
	encPath := fj.DayFilePath(now)

	var existing string
	data, err := os.ReadFile(encPath)
	if err == nil {
		plain, err := crypto.Decrypt(data, passphrase)
		if err != nil {
			return fmt.Errorf("decrypting %s: %w", encPath, err)
		}
		existing = string(plain)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading %s: %w", encPath, err)
	}

	if existing == "" {
		existing = fmt.Sprintf("# %s %s\n", now.Format(cfg.Format.Date), now.Format("Monday"))
	}
	existing += fmt.Sprintf("\n## [%s]\n\n", now.Format(cfg.Format.Time))
	if tmpl != "" {
		existing += tmpl
		if !strings.HasSuffix(tmpl, "\n") {
			existing += "\n"
		}
	}

	tmpFile, err := os.CreateTemp("", "jrnl-md-*.md")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer secureRemove(tmpPath)

	if _, err := tmpFile.WriteString(existing); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	lineCount := strings.Count(existing, "\n") + 1
	if err := editor.Launch(cfg.General.Editor, tmpPath, lineCount); err != nil {
		return fmt.Errorf("launching editor: %w", err)
	}

	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("reading edited file: %w", err)
	}

	dir := filepath.Dir(encPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	enc, err := crypto.Encrypt(edited, passphrase)
	if err != nil {
		return fmt.Errorf("encrypting: %w", err)
	}

	if err := atomicfile.WriteFile(encPath, enc, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", encPath, err)
	}

	return nil
}

func secureRemove(path string) {
	os.Remove(path)
}

func readEntries(fj *journal.FolderJournal, cfg config.Config, tagArgs []string) error {
	entries := fj.AllEntries()

	f, err := buildFilter(tagArgs)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries = f.Apply(entries)

	if flagTags {
		return showTags(entries)
	}

	fmt.Fprintf(os.Stderr, "%d entries found\n", len(entries))

	if len(entries) == 0 {
		return nil
	}

	if flagExport != "" {
		var output string
		var err error
		switch strings.ToLower(flagExport) {
		case "json":
			output, err = export.JSON(entries, cfg)
		case "md", "markdown":
			output, err = export.Markdown(entries, cfg)
		case "txt", "text":
			output, err = export.Text(entries, cfg)
		case "xml":
			output, err = export.XML(entries, cfg)
		case "yaml":
			output, err = export.YAML(entries, cfg)
		default:
			return fmt.Errorf("unknown export format %q (supported: json, md, txt, xml, yaml)", flagExport)
		}
		if err != nil {
			return fmt.Errorf("exporting as %s: %w", flagExport, err)
		}
		fmt.Print(output)
		return nil
	}

	linewrap := cfg.General.Linewrap
	if linewrap == 0 {
		linewrap = display.TerminalWidth()
	}

	indent := ""
	if cfg.General.IndentCharacter != "" {
		indent = cfg.General.IndentCharacter + " "
	}

	dateColor := display.ColorFunc(cfg.Colors.Date)
	bodyColor := display.ColorFunc(cfg.Colors.Body)

	for _, e := range entries {
		if flagShort {
			fmt.Println(e.FormatShort(cfg.Format.Date, cfg.Format.Time))
		} else {
			dateStr := dateColor(e.Date.Format(cfg.Format.Date + " " + cfg.Format.Time))
			fmt.Println(dateStr)
			fmt.Println()

			body := e.Body
			if linewrap > 0 && indent != "" {
				body = display.WrapText(body, linewrap-len(indent))
			} else if linewrap > 0 {
				body = display.WrapText(body, linewrap)
			}
			if indent != "" {
				body = display.IndentBody(body, indent)
			}
			fmt.Println(bodyColor(body))
		}
	}

	return nil
}

func buildFilter(tagArgs []string) (journal.Filter, error) {
	var f journal.Filter
	f.N = flagN
	f.Starred = flagStarred

	if len(tagArgs) > 0 {
		f.Tags = tagArgs
	}

	if flagContains != "" {
		f.Contains = flagContains
	}

	if flagOn != "" {
		start, err := dateparse.Parse(flagOn)
		if err != nil {
			return f, fmt.Errorf("parsing --on date: %w", err)
		}
		startOfDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local)
		endOfDay := time.Date(start.Year(), start.Month(), start.Day(), 23, 59, 59, 0, time.Local)
		f.StartDate = &startOfDay
		f.EndDate = &endOfDay
	}

	if flagFrom != "" {
		start, err := dateparse.Parse(flagFrom)
		if err != nil {
			return f, fmt.Errorf("parsing --from date: %w", err)
		}
		f.StartDate = &start
	}

	if flagTo != "" {
		end, err := dateparse.ParseInclusive(flagTo)
		if err != nil {
			return f, fmt.Errorf("parsing --to date: %w", err)
		}
		f.EndDate = &end
	}

	return f, nil
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

func listJournals(cfg config.Config) error {
	fmt.Println("Journals:")
	for name, j := range cfg.Journals {
		fmt.Printf("  %s -> %s\n", name, j.Path)
	}
	return nil
}

func showTags(entries []journal.Entry) error {
	counts := export.TagCounts(entries)

	if len(counts) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	tags := make([]string, 0, len(counts))
	for tag := range counts {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	for _, tag := range tags {
		fmt.Printf("%-20s : %d\n", tag, counts[tag])
	}
	return nil
}

func hasFilterFlags(cmd *cobra.Command) bool {
	return flagN > 0 || flagShort || flagStarred || flagDelete || flagEncrypt || flagDecrypt || flagChangeTime != "" || flagFrom != "" || flagTo != "" || flagOn != "" || flagContains != "" || flagTags || flagExport != ""
}

func deleteEntries(fj *journal.FolderJournal, cfg config.Config, tagArgs []string) error {
	entries := fj.AllEntries()

	f, err := buildFilter(tagArgs)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries = f.Apply(entries)

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries to delete, because the search returned no results.")
		return nil
	}

	if len(entries) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry found.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries found.\n", len(entries))
	}

	var toDelete []journal.Entry
	for _, e := range entries {
		msg := fmt.Sprintf("Delete entry '%s'?", e.FormatShort(cfg.Format.Date, cfg.Format.Time))
		if prompt.YesNo(os.Stdin, os.Stderr, msg) {
			toDelete = append(toDelete, e)
		}
	}

	if len(toDelete) == 0 {
		return nil
	}

	fj.DeleteEntries(toDelete)

	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	if len(toDelete) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry deleted.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries deleted.\n", len(toDelete))
	}

	return nil
}

func changeTime(fj *journal.FolderJournal, cfg config.Config, tagArgs []string) error {
	newTime, err := dateparse.Parse(flagChangeTime)
	if err != nil {
		return fmt.Errorf("parsing --change-time date: %w", err)
	}

	entries := fj.AllEntries()

	f, err := buildFilter(tagArgs)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries = f.Apply(entries)

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No entries to modify.")
		return nil
	}

	if len(entries) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry found.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries found.\n", len(entries))
	}

	var toChange []journal.Entry
	for _, e := range entries {
		msg := fmt.Sprintf("Change time for '%s'?", e.FormatShort(cfg.Format.Date, cfg.Format.Time))
		if prompt.YesNo(os.Stdin, os.Stderr, msg) {
			toChange = append(toChange, e)
		}
	}

	if len(toChange) == 0 {
		return nil
	}

	fj.ChangeEntryTimes(toChange, newTime)

	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	if len(toChange) == 1 {
		fmt.Fprintf(os.Stderr, "1 entry modified.\n")
	} else {
		fmt.Fprintf(os.Stderr, "%d entries modified.\n", len(toChange))
	}

	return nil
}

func journalEncrypted(jcfg config.JournalConfig, cfg config.Config) bool {
	if jcfg.Encrypt != nil {
		return *jcfg.Encrypt
	}
	return cfg.General.Encrypt
}

func promptPassphrase(msg string) (string, error) {
	fmt.Fprint(os.Stderr, msg)
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}
	return string(pass), nil
}

func promptNewPassphrase() (string, error) {
	pass, err := promptPassphrase("New passphrase: ")
	if err != nil {
		return "", err
	}
	if pass == "" {
		return "", fmt.Errorf("passphrase cannot be empty")
	}
	confirm, err := promptPassphrase("Confirm passphrase: ")
	if err != nil {
		return "", err
	}
	if pass != confirm {
		return "", fmt.Errorf("passphrases do not match")
	}
	return pass, nil
}

func encryptJournal(journalPath, journalName string, cfg config.Config, configPath string) error {
	if journalEncrypted(cfg.Journals[journalName], cfg) {
		return fmt.Errorf("journal %q is already encrypted", journalName)
	}
	passphrase, err := promptNewPassphrase()
	if err != nil {
		return err
	}
	return reencryptJournal(journalPath, journalName, cfg, configPath, false, passphrase, true,
		fmt.Sprintf("Journal %q encrypted", journalName))
}

func decryptJournal(journalPath, journalName string, cfg config.Config, configPath string) error {
	if !journalEncrypted(cfg.Journals[journalName], cfg) {
		return fmt.Errorf("journal %q is not encrypted", journalName)
	}
	passphrase, err := promptPassphrase(fmt.Sprintf("Passphrase for journal %q: ", journalName))
	if err != nil {
		return err
	}
	return reencryptJournal(journalPath, journalName, cfg, configPath, true, passphrase, false,
		fmt.Sprintf("Journal %q decrypted", journalName))
}

func reencryptJournal(journalPath, journalName string, cfg config.Config, configPath string, fromEncrypt bool, passphrase string, toEncrypt bool, successMsg string) error {
	fj := journal.NewFolderJournal(journalPath, journalOptions(cfg, fromEncrypt, passphrase))
	if err := fj.Load(); err != nil {
		return fmt.Errorf("loading journal: %w", err)
	}

	oldFiles, err := fj.DayFiles()
	if err != nil {
		return fmt.Errorf("listing day files: %w", err)
	}

	fj.MarkAllModified()
	fj.SetEncryption(toEncrypt, passphrase)
	if err := fj.Save(); err != nil {
		return fmt.Errorf("saving journal: %w", err)
	}

	for _, f := range oldFiles {
		os.Remove(f)
	}

	jcfg := cfg.Journals[journalName]
	jcfg.Encrypt = boolPtr(toEncrypt)
	cfg.Journals[journalName] = jcfg
	if err := config.Save(cfg, configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s (%d files).\n", successMsg, len(oldFiles))
	return nil
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
