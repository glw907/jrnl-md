package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/glw907/jrnl-md/internal/atomicfile"
	"github.com/glw907/jrnl-md/internal/config"
	"github.com/glw907/jrnl-md/internal/display"
	"github.com/glw907/jrnl-md/internal/export"
	"github.com/glw907/jrnl-md/internal/journal"
)

func readEntries(fj *journal.FolderJournal, cfg config.Config, f *flags, tagArgs []string) error {
	flt, err := buildFilter(f, tagArgs, cfg)
	if err != nil {
		return fmt.Errorf("building filter: %w", err)
	}
	entries, err := fj.Entries(&flt)
	if err != nil {
		return fmt.Errorf("loading journal: %w", err)
	}

	if f.tags {
		return showTags(entries)
	}

	fmt.Fprintf(os.Stderr, "%d entries found\n", len(entries))

	if len(entries) == 0 {
		return nil
	}

	if f.export != "" {
		format := strings.ToLower(f.export)
		switch format {
		case "pretty":
			// fall through to default display below
		case "short":
			f.short = true
		case "tags":
			return showTags(entries)
		case "dates":
			return showDates(entries, cfg.Format.Date)
		default:
			var output string
			var err error
			switch format {
			case export.FormatJSON:
				output, err = export.JSON(entries, cfg)
			case export.FormatMarkdown, "markdown":
				output, err = export.Markdown(entries, cfg)
			case export.FormatText, "text":
				output, err = export.Text(entries, cfg)
			case export.FormatXML:
				output, err = export.XML(entries, cfg)
			case export.FormatYAML:
				output, err = export.YAML(entries, cfg)
			default:
				return fmt.Errorf("unknown export format %q (supported: pretty, short, tags, dates, %s, %s, %s, %s, %s)",
					f.export, export.FormatJSON, export.FormatMarkdown,
					export.FormatText, export.FormatXML, export.FormatYAML)
			}
			if err != nil {
				return fmt.Errorf("exporting as %s: %w", format, err)
			}
			if f.file != "" {
				if err := atomicfile.WriteFile(f.file, []byte(output), 0o600); err != nil {
					return fmt.Errorf("writing export to %s: %w", f.file, err)
				}
				return nil
			}
			fmt.Print(output)
			return nil
		}
	}

	linewrap := cfg.General.Linewrap
	if linewrap == 0 {
		linewrap = display.TerminalWidth()
	}

	indent := ""
	if cfg.General.IndentCharacter != "" {
		indent = cfg.General.IndentCharacter + " "
	}

	dateColorFn := display.ColorFunc(cfg.Colors.Date)
	dateColor := func(a ...any) string {
		if dateColorFn != nil {
			return dateColorFn(a...)
		}
		s := ""
		for _, v := range a {
			s += fmt.Sprint(v)
		}
		return s
	}
	bodyColorFn := display.ColorFunc(cfg.Colors.Body)
	bodyColor := func(a ...any) string {
		if bodyColorFn != nil {
			return bodyColorFn(a...)
		}
		s := ""
		for _, v := range a {
			s += fmt.Sprint(v)
		}
		return s
	}

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
			if cfg.General.Highlight {
				tagColorFn := display.ColorFunc(cfg.Colors.Tags)
				if tagColorFn == nil {
					tagColorFn = display.ColorFunc("cyan")
				}
				body = display.HighlightTags(body, cfg.Format.TagSymbols, tagColorFn)
			}
			fmt.Println(bodyColor(body))
		}
	}

	return nil
}

func showDates(entries []journal.Entry, dateFmt string) error {
	counts := make(map[string]int)
	var order []string
	for _, e := range entries {
		d := e.Date.Format(dateFmt)
		if counts[d] == 0 {
			order = append(order, d)
		}
		counts[d]++
	}
	for _, d := range order {
		n := counts[d]
		if n == 1 {
			fmt.Printf("%s: 1 entry\n", d)
		} else {
			fmt.Printf("%s: %d entries\n", d, n)
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

	type tagCount struct {
		tag   string
		count int
	}
	tc := make([]tagCount, 0, len(counts))
	for tag, n := range counts {
		tc = append(tc, tagCount{tag, n})
	}
	sort.Slice(tc, func(i, j int) bool {
		if tc[i].count != tc[j].count {
			return tc[i].count > tc[j].count
		}
		return tc[i].tag < tc[j].tag
	})

	for _, item := range tc {
		fmt.Printf("%-20s : %d\n", item.tag, item.count)
	}
	return nil
}
