package main

import (
	"fmt"
	"os"
)

func main() {
	cmd := newRootCmd()
	cmd.SetArgs(preprocessArgs(os.Args[1:]))
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
