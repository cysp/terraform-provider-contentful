package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/stretchr/testify/require"
)

const contentTypeListResourceConfig = `
	provider "contentful" {}

	variable "space_id" {
		type = string
	}

	variable "environment_id" {
		type = string
	}

	list "contentful_content_type" "content_types" {
		provider = contentful

		config {
			space_id       = var.space_id
			environment_id = var.environment_id
		}

		include_resource = true
	}
`

func TestAccContentTypeListResource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("master"),
	}

	server.SetContentType("0p38pssr0fi3", "master", "author", cm.ContentTypeRequestData{
		Name:         "Author",
		DisplayField: "name",
		Fields: []cm.ContentTypeRequestDataFieldsItem{
			{
				ID:       "name",
				Name:     "Name",
				Type:     "Symbol",
				Required: cm.NewOptBool(true),
			},
			{
				ID:       "avatar",
				Name:     "Avatar",
				Type:     "Link",
				LinkType: cm.NewOptString("Asset"),
				Required: cm.NewOptBool(false),
			},
			{
				ID:       "blurb",
				Name:     "Blurb",
				Type:     "RichText",
				Required: cm.NewOptBool(false),
			},
		},
	})
	activationResponse, err := server.Handler().ActivateContentType(t.Context(), cm.ActivateContentTypeParams{
		SpaceID:            "0p38pssr0fi3",
		EnvironmentID:      "master",
		ContentTypeID:      "author",
		XContentfulVersion: 1,
	})
	require.NoError(t, err)
	require.IsType(t, &cm.ContentTypeStatusCode{}, activationResponse)

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config:          contentTypeListResourceConfig,
				ConfigVariables: configVariables,
				Query:           true,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("contentful_content_type.content_types", 1),
					querycheck.ExpectIdentity("contentful_content_type.content_types", map[string]knownvalue.Check{
						"space_id":        knownvalue.StringExact("0p38pssr0fi3"),
						"environment_id":  knownvalue.StringExact("master"),
						"content_type_id": knownvalue.StringExact("author"),
					}),
					querycheck.ExpectResourceKnownValues("contentful_content_type.content_types", queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
						"space_id":        knownvalue.StringExact("0p38pssr0fi3"),
						"environment_id":  knownvalue.StringExact("master"),
						"content_type_id": knownvalue.StringExact("author"),
					}), []querycheck.KnownValueCheck{
						{
							Path:       tfjsonpath.New("id"),
							KnownValue: knownvalue.StringExact("0p38pssr0fi3/master/author"),
						},
						{
							Path:       tfjsonpath.New("name"),
							KnownValue: knownvalue.StringExact("Author"),
						},
						{
							Path:       tfjsonpath.New("published_version"),
							KnownValue: knownvalue.NotNull(),
						},
						{
							Path: tfjsonpath.New("fields"),
							KnownValue: knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"id":       knownvalue.StringExact("name"),
									"name":     knownvalue.StringExact("Name"),
									"type":     knownvalue.StringExact("Symbol"),
									"required": knownvalue.Bool(true),
								}),
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"id":        knownvalue.StringExact("avatar"),
									"name":      knownvalue.StringExact("Avatar"),
									"type":      knownvalue.StringExact("Link"),
									"link_type": knownvalue.StringExact("Asset"),
									"required":  knownvalue.Bool(false),
								}),
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"id":       knownvalue.StringExact("blurb"),
									"name":     knownvalue.StringExact("Blurb"),
									"type":     knownvalue.StringExact("RichText"),
									"required": knownvalue.Bool(false),
								}),
							}),
						},
						{
							Path:       tfjsonpath.New("timeouts"),
							KnownValue: knownvalue.Null(),
						},
					}),
				},
			},
		},
	})
}

func TestAccContentTypeListResourceIncludesUnpublished(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	server.SetContentType("0p38pssr0fi3", "master", "unpublished", cm.ContentTypeRequestData{
		Name:         "Unpublished",
		DisplayField: "title",
		Fields: []cm.ContentTypeRequestDataFieldsItem{{
			ID: "title", Name: "Title", Type: "Symbol", Required: cm.NewOptBool(true),
		}},
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{{
			Config: contentTypeListResourceConfig,
			ConfigVariables: config.Variables{
				"space_id":       config.StringVariable("0p38pssr0fi3"),
				"environment_id": config.StringVariable("master"),
			},
			Query: true,
			QueryResultChecks: []querycheck.QueryResultCheck{
				querycheck.ExpectLengthAtLeast("contentful_content_type.content_types", 1),
				querycheck.ExpectIdentity("contentful_content_type.content_types", map[string]knownvalue.Check{
					"space_id":        knownvalue.StringExact("0p38pssr0fi3"),
					"environment_id":  knownvalue.StringExact("master"),
					"content_type_id": knownvalue.StringExact("unpublished"),
				}),
				querycheck.ExpectResourceKnownValues("contentful_content_type.content_types", queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
					"space_id":        knownvalue.StringExact("0p38pssr0fi3"),
					"environment_id":  knownvalue.StringExact("master"),
					"content_type_id": knownvalue.StringExact("unpublished"),
				}), []querycheck.KnownValueCheck{
					{Path: tfjsonpath.New("name"), KnownValue: knownvalue.StringExact("Unpublished")},
					{Path: tfjsonpath.New("published_version"), KnownValue: knownvalue.Null()},
				}),
			},
		}},
	})
}

func TestAccContentTypeListResourceNotFoundEnvironment(t *testing.T) {
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

				list "contentful_content_type" "content_types" {
					provider = contentful

					config {
						space_id       = var.space_id
						environment_id = var.environment_id
					}
				}
				`,
				ConfigVariables: configVariables,
				Query:           true,
				ExpectError:     regexp.MustCompile("Failed to list content types"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("contentful_content_type.content_types", 0),
				},
			},
		},
	})
}
