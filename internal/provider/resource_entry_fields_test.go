package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccEntryResourceIgnoreChangesPreservesRemoteFieldLifecycle(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	config := func(managed string) string {
		return fmt.Sprintf(`
resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "environment"
  entry_id        = "entry"
  content_type_id = "article"
  fields = {
    managed = jsonencode({ "en-US" = %q })
  }

  lifecycle {
    ignore_changes = [fields["external"]]
  }
}
`, managed)
	}

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: config("one"),
			Check: func(*terraform.State) error {
				update, _ := requireEntryUpdateThenPublish(t, recorder.snapshot())
				require.JSONEq(t, `{"en-US":"one"}`, string(update.fields["managed"]))
				require.NotContains(t, update.fields, "external")

				return nil
			},
		},
		{
			PreConfig: func() {
				entry := getTestEntry(t, server)
				response, putErr := server.Handler().PutEntry(t.Context(), &cm.EntryRequest{
					Fields: cm.NewOptEntryFields(cm.EntryFields{
						"managed":  jx.Raw(`{"en-US":"one"}`),
						"external": jx.Raw(`{"en-US":"remote"}`),
					}),
					Metadata: entry.Metadata,
				}, cm.PutEntryParams{
					SpaceID: "space", EnvironmentID: "environment", EntryID: "entry", XContentfulVersion: entry.Sys.Version,
				})
				require.NoError(t, putErr)
				require.IsType(t, &cm.EntryStatusCode{}, response)
				recorder.reset()
			},
			Config: config("one"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: func(*terraform.State) error {
				requireNoEntryMutations(t, recorder)
				entry := getTestEntry(t, server)
				require.True(t, entry.Sys.PublishedVersion.IsSet())
				require.Less(t, entry.Sys.PublishedVersion.Or(0), entry.Sys.Version)

				return nil
			},
		},
		{
			PreConfig: recorder.reset,
			Config:    config("two"),
			Check: func(*terraform.State) error {
				update, _ := requireEntryUpdateThenPublish(t, recorder.snapshot())
				require.JSONEq(t, `{"en-US":"two"}`, string(update.fields["managed"]))
				require.JSONEq(t, `{"en-US":"remote"}`, string(update.fields["external"]))

				return nil
			},
		},
		{
			PreConfig: func() {
				entry := getTestEntry(t, server)
				response, putErr := server.Handler().PutEntry(t.Context(), &cm.EntryRequest{
					Fields: cm.NewOptEntryFields(cm.EntryFields{
						"managed": jx.Raw(`{"en-US":"two"}`),
					}),
					Metadata: entry.Metadata,
				}, cm.PutEntryParams{
					SpaceID: "space", EnvironmentID: "environment", EntryID: "entry", XContentfulVersion: entry.Sys.Version,
				})
				require.NoError(t, putErr)
				require.IsType(t, &cm.EntryStatusCode{}, response)
				recorder.reset()
			},
			Config: config("three"),
			Check: func(*terraform.State) error {
				update, _ := requireEntryUpdateThenPublish(t, recorder.snapshot())
				require.JSONEq(t, `{"en-US":"three"}`, string(update.fields["managed"]))
				require.NotContains(t, update.fields, "external")

				return nil
			},
		},
	}})
}

func TestAccEntryResourceContentTypeDefaultLifecycle(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	defaulting := &entryCreateDefaultingAdapter{delegate: server, errorSink: fixture.errorSink}
	recorder.delegate = defaulting
	configIgnoringDefault := func(managed string) string {
		return fmt.Sprintf(`
resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "environment"
  entry_id        = "entry"
  content_type_id = "article"
  fields = {
    managed = jsonencode({ "en-US" = %q })
  }

  lifecycle {
    ignore_changes = [fields["defaulted"]]
  }
}
`, managed)
	}
	configOmittingDefault := func(managed string) string {
		return fmt.Sprintf(`
resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "environment"
  entry_id        = "entry"
  content_type_id = "article"
  fields = {
    managed = jsonencode({ "en-US" = %q })
  }
}
`, managed)
	}

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: configIgnoringDefault("one"),
			Check: func(*terraform.State) error {
				update, _ := requireEntryUpdateThenPublish(t, recorder.snapshot())
				require.JSONEq(t, `{"en-US":"one"}`, string(update.fields["managed"]))
				require.NotContains(t, update.fields, "defaulted", "the HTTP adapter must apply the default after recording Terraform's request")

				entry := getTestEntry(t, server)
				require.JSONEq(t, `{"en-US":"content-type default"}`, string(entry.Fields.Value["defaulted"]))
				require.True(t, entry.Sys.PublishedVersion.IsSet())
				require.Less(t, entry.Sys.PublishedVersion.Or(0), entry.Sys.Version)

				return nil
			},
		},
		{
			PreConfig: recorder.reset,
			Config:    configIgnoringDefault("two"),
			Check: func(*terraform.State) error {
				update, _ := requireEntryUpdateThenPublish(t, recorder.snapshot())
				require.JSONEq(t, `{"en-US":"two"}`, string(update.fields["managed"]))
				require.JSONEq(t, `{"en-US":"content-type default"}`, string(update.fields["defaulted"]))

				return nil
			},
		},
		{
			PreConfig: recorder.reset,
			Config:    configOmittingDefault("two"),
			Check: func(*terraform.State) error {
				update, _ := requireEntryUpdateThenPublish(t, recorder.snapshot())
				require.JSONEq(t, `{"en-US":"two"}`, string(update.fields["managed"]))
				require.NotContains(t, update.fields, "defaulted")

				entry := getTestEntry(t, server)
				require.NotContains(t, entry.Fields.Value, "defaulted")
				require.True(t, entry.Sys.PublishedVersion.IsSet())
				require.Less(t, entry.Sys.PublishedVersion.Or(0), entry.Sys.Version)

				return nil
			},
		},
		{
			PreConfig: recorder.reset,
			Config:    configOmittingDefault("two"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: func(*terraform.State) error {
				requireNoEntryMutations(t, recorder)

				return nil
			},
		},
	}})
}

func TestAccEntryResourceNullFieldLifecycle(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	config := func(managedValue string, includeTerraformNull, includeAdditionalRawNull bool) string {
		terraformNull := ""
		if includeTerraformNull {
			terraformNull = "terraform_null = null"
		}

		additionalRawNull := ""
		if includeAdditionalRawNull {
			additionalRawNull = "additional_raw_null = jsonencode(null)"
		}

		return fmt.Sprintf(`
resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "environment"
  entry_id        = "entry"
  content_type_id = "article"
  fields = {
    managed  = %s
    raw_null = jsonencode(null)
    %s
    %s
  }
}
`, managedValue, terraformNull, additionalRawNull)
	}
	managed := `jsonencode({ "en-US" = "one" })`

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: config(managed, true, false),
			Check: func(*terraform.State) error {
				update, _ := requireEntryUpdateThenPublish(t, recorder.snapshot())
				require.JSONEq(t, `{"en-US":"one"}`, string(update.fields["managed"]))
				require.NotContains(t, update.fields, "terraform_null")
				require.JSONEq(t, `null`, string(update.fields["raw_null"]))
				require.NotContains(t, update.fields, "additional_raw_null")

				return nil
			},
		},
		{
			PreConfig: recorder.reset,
			Config:    config(managed, false, false),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionUpdate),
				plancheck.ExpectKnownValue("contentful_entry.test", tfjsonpath.New("published_version"), knownvalue.NotNull()),
			}},
			Check: func(*terraform.State) error {
				requireNoEntryMutations(t, recorder, "removing an omitted Terraform null must not write or publish an Entry")

				return nil
			},
		},
		{
			PreConfig: recorder.reset,
			Config:    config(managed, false, false),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: func(*terraform.State) error {
				requireNoEntryMutations(t, recorder)

				return nil
			},
		},
		{
			PreConfig: recorder.reset,
			Config:    config(managed, false, true),
			Check: func(*terraform.State) error {
				update, _ := requireEntryUpdateThenPublish(t, recorder.snapshot())
				require.JSONEq(t, `{"en-US":"one"}`, string(update.fields["managed"]))
				require.JSONEq(t, `null`, string(update.fields["raw_null"]))
				require.JSONEq(t, `null`, string(update.fields["additional_raw_null"]))

				return nil
			},
		},
		{
			PreConfig: recorder.reset,
			Config:    config("null", false, true),
			Check: func(*terraform.State) error {
				update, _ := requireEntryUpdateThenPublish(t, recorder.snapshot())
				require.NotContains(t, update.fields, "managed")
				require.JSONEq(t, `null`, string(update.fields["raw_null"]))
				require.JSONEq(t, `null`, string(update.fields["additional_raw_null"]))

				entry := getTestEntry(t, server)
				require.NotContains(t, entry.Fields.Value, "managed")
				require.NotContains(t, entry.Fields.Value, "raw_null")
				require.NotContains(t, entry.Fields.Value, "additional_raw_null")
				require.True(t, entry.Sys.PublishedVersion.IsSet())
				require.Less(t, entry.Sys.PublishedVersion.Or(0), entry.Sys.Version)

				return nil
			},
		},
		{
			PreConfig: recorder.reset,
			Config:    config("null", false, true),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: func(*terraform.State) error {
				requireNoEntryMutations(t, recorder)

				return nil
			},
		},
	}})
}

func TestAccEntryResourceUpdateRejectsResponseOnlyField(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryAdditionalUpdateFieldAdapter{delegate: server, errorSink: fixture.errorSink}
	recorder.delegate = fault
	config := managedEntryConfig

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{Config: config("one")},
		{
			PreConfig: func() {
				recorder.reset()
				fault.shot.arm()
			},
			Config:      config("two"),
			ExpectError: regexp.MustCompile(`Unexpected entry fields`),
		},
		{
			PreConfig: func() {
				requests := recorder.snapshot()
				require.Len(t, requests, 1, "the contradictory update response must prevent publication")
				requireEntryUpdate(t, requests[0])
				require.Equal(t, "2", requests[0].version)
				require.JSONEq(t, `{"en-US":"two"}`, string(requests[0].fields["managed"]))
				require.NotContains(t, requests[0].fields, "response-only")
				recorder.reset()
			},
			Config: config("two"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: func(*terraform.State) error {
				requireNoEntryMutations(t, recorder)

				return nil
			},
		},
	}})
}

func TestAccEntryResourceUpdatePublishRejectsResponseOnlyField(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryAdditionalPublishFieldAdapter{delegate: server, errorSink: fixture.errorSink}
	recorder.delegate = fault
	config := managedEntryConfig

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{Config: config("one")},
		{
			PreConfig: func() {
				recorder.reset()
				fault.shot.arm()
			},
			Config:      config("two"),
			ExpectError: regexp.MustCompile(`Unexpected entry fields`),
		},
		{
			PreConfig: func() {
				requireEntryUpdateThenPublish(t, recorder.snapshot())
				recorder.reset()
			},
			Config: config("two"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: func(*terraform.State) error {
				requireNoEntryMutations(t, recorder)

				return nil
			},
		},
	}})
}

func TestAccEntryResourceAnomalousCreatePublishRejectsResponseOnlyField(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	anomalousVersion := 3
	tupleFault := &entryPublishTupleAdapter{delegate: server, version: &anomalousVersion, errorSink: fixture.errorSink}
	fieldFault := &entryAdditionalPublishFieldAdapter{delegate: tupleFault, errorSink: fixture.errorSink}
	recorder.delegate = fieldFault

	fieldFault.shot.arm()
	tupleFault.shot.arm()

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{
			Config:      managedEntryConfig("one"),
			ExpectError: regexp.MustCompile(`Unexpected entry fields`),
		},
	}})

	_, publish := requireEntryUpdateThenPublish(t, recorder.snapshot())
	require.Equal(t, "1", publish.version)
}

func TestAccEntryResourceMetadataReadReorderingDoesNotDrift(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	reorderedRead := &entryReorderedMetadataReadAdapter{delegate: server, errorSink: fixture.errorSink}
	recorder.delegate = reorderedRead
	config := `
resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "environment"
  entry_id        = "entry"
  content_type_id = "article"
  fields          = { managed = jsonencode({ "en-US" = "one" }) }
  metadata        = { tags = ["first", "second"] }
}
`

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{Config: config},
		{
			PreConfig: recorder.reset,
			Config:    config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("contentful_entry.test", "metadata.tags.#", "2"),
				resource.TestCheckResourceAttr("contentful_entry.test", "metadata.tags.0", "first"),
				resource.TestCheckResourceAttr("contentful_entry.test", "metadata.tags.1", "second"),
				func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "response reordering must not write or publish an Entry")

					return nil
				},
			),
		},
	}})
}

func TestAccEntryResourceMetadataConfigReorderingIsRepresentationOnly(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	config := func(first, second string) string {
		return fmt.Sprintf(`
resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "environment"
  entry_id        = "entry"
  content_type_id = "article"
  fields          = { managed = jsonencode({ "en-US" = "one" }) }
  metadata        = { tags = [%q, %q] }
}
`, first, second)
	}

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{Config: config("first", "second")},
		{
			PreConfig: recorder.reset,
			Config:    config("second", "first"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionUpdate),
			}},
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("contentful_entry.test", "metadata.tags.0", "second"),
				resource.TestCheckResourceAttr("contentful_entry.test", "metadata.tags.1", "first"),
				func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "reordering metadata must not write or publish an Entry")

					entry := getTestEntry(t, server)
					metadata, ok := entry.Metadata.Get()
					require.True(t, ok)
					require.Equal(t, "first", metadata.Tags[0].Sys.ID)
					require.Equal(t, "second", metadata.Tags[1].Sys.ID)

					return nil
				},
			),
		},
	}})
}
