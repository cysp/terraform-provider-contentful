package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccEntryListResource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	spaceID := "0p38pssr0fi3"
	environmentID := "test"

	server.SetEntry(spaceID, environmentID, "author", "33lT7CnYPNJGdls6nONU3t", cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"name": jx.Raw(`"Author 1"`),
		}),
	})

	server.SetEntry(spaceID, environmentID, "author", "author2", cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"name": jx.Raw(`"Author 2"`),
		}),
	})

	server.SetEntry(spaceID, environmentID, "post", "post1", cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{
			"name": jx.Raw(`"Post 1"`),
		}),
	})

	configVariables := config.Variables{
		"space_id":       config.StringVariable(spaceID),
		"environment_id": config.StringVariable(environmentID),
		"content_type":   config.StringVariable("author"),
		"query": config.MapVariable(map[string]config.Variable{
			"fields.name[ne]": config.StringVariable("nonexistent"),
		}),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: `
				provider "contentful" {}

				variable "space_id" {
					type = string
				}

				variable "environment_id" {
					type = string
				}

				variable "content_type" {
					type = string
				}

				variable "query" {
					type = map(string)
				}

				list "contentful_entry" "entries" {
					provider = contentful

					config {
						space_id       = var.space_id
						environment_id = var.environment_id

						content_type = var.content_type
						query        = var.query
					}
				}
				`,
				ConfigVariables: configVariables,
				Query:           true,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("contentful_entry.entries", 1),
					querycheck.ExpectIdentity("contentful_entry.entries", map[string]knownvalue.Check{
						"space_id":       knownvalue.StringExact("0p38pssr0fi3"),
						"environment_id": knownvalue.StringExact("test"),
						"entry_id":       knownvalue.StringExact("33lT7CnYPNJGdls6nONU3t"),
					}),
				},
			},
		},
	})
}

func TestAccEntryListResourceNotFoundEnvironment(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: `
				provider "contentful" {}

				variable "space_id" {
					type = string
				}

				variable "environment_id" {
					type = string
				}

				list "contentful_entry" "entries" {
					provider = contentful

					config {
						space_id       = var.space_id
						environment_id = var.environment_id
					}
				}
				`,
				ConfigVariables: configVariables,
				Query:           true,
				ExpectError:     regexp.MustCompile("Failed to list entries"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("contentful_entry.entries", 0),
				},
			},
		},
	})
}
