package main

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
