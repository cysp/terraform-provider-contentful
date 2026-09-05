package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs --provider-name=terraform-provider-contentful

// set by goreleaser.
var version = "dev"

//nolint:gochecknoglobals // Set by GoReleaser.
var commit = "unknown"

func formatVersion(version, commit string) string {
	return fmt.Sprintf("terraform-provider-contentful %s (commit %s)\n", version, commit)
}

func main() {
	var (
		debug       bool
		showVersion bool
	)

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.BoolVar(&showVersion, "version", false, "print the provider version and exit")
	flag.Parse()

	if showVersion {
		_, err := fmt.Fprint(os.Stdout, formatVersion(version, commit))
		if err != nil {
			log.Fatal(err.Error())
		}

		return
	}

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/cysp/contentful",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.Factory(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
