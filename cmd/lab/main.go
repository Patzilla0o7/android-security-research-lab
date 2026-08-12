package main

import (
	"os"

	"github.com/Patzilla0o7/android-security-research-lab/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
