package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/cysp/terraform-provider-contentful/internal/migrate"
)

const checkFailedExitCode = 2

func main() {
	var write bool

	var check bool

	flag.BoolVar(&write, "write", false, "rewrite files in place")
	flag.BoolVar(&check, "check", false, "exit non-zero when migrations would change files")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] [path ...]\n\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Migrates Terraform configuration for terraform-provider-contentful schema changes.")
		fmt.Fprintln(flag.CommandLine.Output(), "When no paths are provided, the current directory is scanned recursively.")
		fmt.Fprintln(flag.CommandLine.Output(), "\nFlags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}

	report, err := migrate.Run(context.Background(), migrate.Options{
		Paths: paths,
		Write: write,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, change := range report.Changes {
		action := "would update"
		if write {
			action = "updated"
		}

		fmt.Fprintf(os.Stdout, "%s: %s: %s\n", action, change.Path, change.Description)
	}

	for _, skip := range report.Skips {
		fmt.Fprintf(os.Stderr, "skipped: %s: %s: %s\n", skip.Path, skip.Address, skip.Reason)
	}

	if check && !write && len(report.Changes) > 0 {
		os.Exit(checkFailedExitCode)
	}
}
