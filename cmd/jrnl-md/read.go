package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/display"
	"github.com/glw907/jrnl-md/internal/export"
	"github.com/glw907/jrnl-md/internal/journal"
)

func readEntries(fj *journal.FolderJournal, cfg config.Config, f *flags, tagArgs []string) error {
	entries := fj.AllEntries()

	flt, err := buildFilter(f, tagArgs)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries = flt.Apply(entries)

	if f.tags {
		return showTags(entries)
	}

	fmt.Fprintf(os.Stderr, "%d entries found\n", len(entries))

	if len(entries) == 0 {
		return nil
	}

	if f.export != "" {
		var output string
		var err error
		switch strings.ToLower(f.export) {
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
			return fmt.Errorf("unknown export format %q (supported: json, md, txt, xml, yaml)", f.export)
		}
		if err != nil {
			return fmt.Errorf("exporting as %s: %w", f.export, err)
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
		if f.short {
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
