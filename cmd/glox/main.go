package main

import (
	"fmt"
	"os"

	"github.com/tangzero/glox"
)

// exit codes from https://man.freebsd.org/cgi/man.cgi?query=sysexits
const (
	ExitUsage    = 64
	ExitSoftware = 70
)

func main() {
	if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "Usage: glox [script]")
		os.Exit(ExitUsage)
	}

	if len(os.Args) == 2 {
		script := os.Args[1]
		if err := glox.RunFile(script); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(ExitSoftware)
		}
		return // success
	}

	if err := glox.RunPrompt(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitSoftware)
	}
}
