package main

import (
	"strings"

	"github.com/glw907/jrnl-md/internal/config"
)

// preprocessArgs converts -N numeric shorthand (e.g. -3) to -n N
// before cobra parses the flags.
func preprocessArgs(args []string) []string {
	var result []string
	for _, arg := range args {
		if len(arg) >= 2 && arg[0] == '-' && arg[1] >= '0' && arg[1] <= '9' {
			allDigits := true
			for _, c := range arg[1:] {
				if c < '0' || c > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				result = append(result, "-n", arg[1:])
				continue
			}
		}
		result = append(result, arg)
	}
	return result
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
