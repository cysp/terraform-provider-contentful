package provider_test

import (
	"regexp"
	"strconv"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccEntryResourceFailedPublishDoesNotRetryWithoutRefresh(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryRejectedPublishAdapter{delegate: server}
	recorder.delegate = fault
	config := managedEntryConfig

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{
			{Config: config("one"), Check: resource.TestCheckResourceAttr("contentful_entry.test", "published_version", "1")},
			{
				PreConfig: func() {
					recorder.reset()
					fault.shot.arm()
				},
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`(?s)Failed to publish entry.*publication was\s+not\s+confirmed`),
			},
			{
				PreConfig: func() {
					update, publish := requireEntryUpdateThenPublish(t, recorder.snapshot())
					require.Equal(t, "2", update.version)
					require.JSONEq(t, `{"en-US":"two"}`, string(update.fields["managed"]))
					require.Equal(t, "3", publish.version)
					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(state *terraform.State) error {
					requireNoEntryMutations(t, recorder, "an unchanged apply must not retry publication")
					entry := getTestEntry(t, server)
					require.Equal(t, 3, entry.Sys.Version)
					require.Equal(t, 1, entry.Sys.PublishedVersion.Or(0))
					require.NoError(t, resource.TestCheckResourceAttr("contentful_entry.test", "published_version", "1")(state))

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourceChangedConfigAuthorsAndPublishesNewDraftWithoutRefresh(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryRejectedPublishAdapter{delegate: server}
	recorder.delegate = fault
	config := managedEntryConfig

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{
			{Config: config("one")},
			{
				PreConfig: func() {
					recorder.reset()
					fault.shot.arm()
				},
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`Failed to publish entry`),
			},
			{
				PreConfig: recorder.reset,
				Config:    config("three"),
				Check: func(state *terraform.State) error {
					update, publish := requireEntryUpdateThenPublish(t, recorder.snapshot())
					require.Equal(t, "3", update.version, "the new draft must be fenced by the exact current draft version")
					require.JSONEq(t, `{"en-US":"three"}`, string(update.fields["managed"]))
					require.Equal(t, "4", publish.version, "only the replacement draft may be published")

					entry := getTestEntry(t, server)
					require.Equal(t, 4, entry.Sys.PublishedVersion.Or(0))
					require.JSONEq(t, `{"en-US":"three"}`, string(entry.Fields.Value["managed"]))
					require.NoError(t, resource.TestCheckResourceAttr("contentful_entry.test", "published_version", "4")(state))

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourceAmbiguousDraftWriteIsNotClaimedAfterRefresh(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	failure := &entryCommittedUpdateFailureAdapter{delegate: server}
	recorder.delegate = failure
	config := managedEntryConfig

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{Config: config("one")},
		{
			PreConfig: func() {
				recorder.reset()
				failure.failAfterCommitOnce()
			},
			Config:      config("two"),
			ExpectError: regexp.MustCompile(`Failed to update entry`),
		},
		{
			PreConfig: func() {
				requests := recorder.snapshot()
				require.Len(t, requests, 1, "an ambiguous draft write must not be transparently replayed")
				requireEntryUpdate(t, requests[0])

				writtenFromVersion, err := strconv.Atoi(requests[0].version)
				require.NoError(t, err)

				entry := getTestEntry(t, server)
				require.Equal(t, writtenFromVersion+1, entry.Sys.Version)
				require.Less(t, entry.Sys.PublishedVersion.Or(0), entry.Sys.Version)
				require.JSONEq(t, `{"en-US":"two"}`, string(entry.Fields.Value["managed"]))

				recorder.reset()
			},
			Config: config("two"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: func(state *terraform.State) error {
				requireNoEntryMutations(t, recorder, "refreshing matching content must not authorize publication")

				entry := getTestEntry(t, server)
				require.Less(t, entry.Sys.PublishedVersion.Or(0), entry.Sys.Version)
				require.NoError(t, resource.TestCheckResourceAttr("contentful_entry.test", "published_version", strconv.Itoa(entry.Sys.PublishedVersion.Or(0)))(state))

				return nil
			},
		},
	}})
}

func TestAccEntryResourceCreatePublishFailureRequiresReplacement(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryRejectedPublishAdapter{delegate: server}
	recorder.delegate = fault
	config := managedEntryConfig("one")

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					fault.shot.arm()
				},
				Config:      config,
				ExpectError: regexp.MustCompile(`Failed to publish entry`),
			},
			{
				PreConfig: func() {
					update, publish := requireEntryUpdateThenPublish(t, recorder.snapshot())
					require.False(t, update.versionPresent)
					require.Empty(t, update.version)
					require.True(t, publish.versionPresent)
					require.Equal(t, "1", publish.version)
					recorder.reset()
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionReplace),
				}},
				Check: func(state *terraform.State) error {
					update, publish := requireEntryUpdateThenPublish(t, recorder.snapshot())
					require.False(t, update.versionPresent)
					require.Empty(t, update.version)
					require.True(t, publish.versionPresent)
					require.Equal(t, "1", publish.version)
					require.NoError(t, resource.TestCheckResourceAttr("contentful_entry.test", "published_version", "1")(state))

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourceFailedPublishDoesNotRetryAfterRefresh(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryRejectedPublishAdapter{delegate: server}
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
			ExpectError: regexp.MustCompile(`Failed to publish entry`),
		},
		{
			PreConfig: recorder.reset,
			Config:    config("two"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: func(state *terraform.State) error {
				requireNoEntryMutations(t, recorder, "refreshing the provider-authored draft must not retry publication")
				require.NoError(t, resource.TestCheckResourceAttr("contentful_entry.test", "published_version", "1")(state))

				return nil
			},
		},
	}})
}

func TestAccEntryResourceHigherPostPublishCurrentVersionIsAdoptedWithWarning(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryHigherPostPublishCurrentVersionAdapter{delegate: server, server: server, errorSink: fixture.errorSink}
	recorder.delegate = fault
	config := managedEntryConfig

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{Config: config("one")},
			{
				PreConfig: func() {
					recorder.reset()
					fault.shot.arm()
				},
				Config: config("two"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_entry.test", "published_version", "3"),
					resource.TestCheckResourceAttr("contentful_entry.test", "fields.managed", `{"en-US":"two"}`),
				),
			},
			{
				PreConfig: recorder.reset,
				Config:    config("three"),
				Check: func(*terraform.State) error {
					update, publish := requireEntryUpdateThenPublish(t, recorder.snapshot())
					require.Equal(t, "6", update.version)
					require.JSONEq(t, `{"en-US":"three"}`, string(update.fields["managed"]))
					require.Equal(t, "7", publish.version)

					entry := getTestEntry(t, server)
					require.Equal(t, 8, entry.Sys.Version)
					require.Equal(t, 7, entry.Sys.PublishedVersion.Or(0))

					return nil
				},
			},
			{
				PreConfig: recorder.reset,
				Config:    config("three"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder)

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourceTypedPublishContradictionDoesNotAuthorizeRetry(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	responseVersion := 3
	publishedVersion := 2
	tupleFault := &entryPublishTupleAdapter{
		delegate: server, version: &responseVersion, publishedVersion: &publishedVersion, errorSink: fixture.errorSink,
	}
	recorder.delegate = tupleFault
	config := managedEntryConfig
	options := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: options,
		Steps: []resource.TestStep{
			{Config: config("one")},
			{
				PreConfig: func() {
					recorder.reset()
					tupleFault.shot.arm()
				},
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`Unexpected entry publication response`),
			},
			{
				PreConfig: recorder.reset,
				Config:    config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder)

					return nil
				},
			},
			{
				PreConfig: func() {
					options.Plan.NoRefresh = false

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
		},
	})
}

func TestAccEntryResourceExternalAdvanceDoesNotAuthorizePublication(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryRejectedPublishAdapter{delegate: server}
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
			ExpectError: regexp.MustCompile(`Failed to publish entry`),
		},
		{
			PreConfig: func() {
				entry := getTestEntry(t, server)
				response, putErr := server.Handler().PutEntry(t.Context(), &cm.EntryRequest{Fields: entry.Fields, Metadata: entry.Metadata}, cm.PutEntryParams{
					SpaceID: "space", EnvironmentID: "environment", EntryID: "entry", XContentfulVersion: cm.NewOptInt(entry.Sys.Version),
				})
				require.NoError(t, putErr)
				require.IsType(t, &cm.EntryStatusCode{}, response)
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

func TestAccEntryResourceExternalUnpublishDoesNotAuthorizePublication(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryRejectedPublishAdapter{delegate: server}
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
			ExpectError: regexp.MustCompile(`Failed to publish entry`),
		},
		{
			PreConfig: func() {
				pending := getTestEntry(t, server)
				response, err := server.Handler().UnpublishEntry(t.Context(), cm.UnpublishEntryParams{
					SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
				})
				require.NoError(t, err)

				unpublished, ok := response.(*cm.Entry)
				require.True(t, ok)
				require.Equal(t, pending.Sys.Version+1, unpublished.Sys.Version)
				require.False(t, unpublished.Sys.PublishedVersion.IsSet())
				recorder.reset()
			},
			Config: config("two"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr("contentful_entry.test", "published_version"),
				func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "external unpublish must revoke stale publication authority")

					return nil
				},
			),
		},
	}})
}

func TestAccEntryResourceInterveningDraftDoesNotAuthorizePublicationWithoutRefresh(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryRejectedPublishAdapter{delegate: server}
	recorder.delegate = fault
	config := managedEntryConfig
	options := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: options,
		Steps: []resource.TestStep{
			{Config: config("one")},
			{
				PreConfig: func() {
					recorder.reset()
					fault.shot.arm()
				},
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`Failed to publish entry`),
			},
			{
				PreConfig: func() {
					entry := getTestEntry(t, server)
					response, putErr := server.Handler().PutEntry(t.Context(), &cm.EntryRequest{Fields: entry.Fields, Metadata: entry.Metadata}, cm.PutEntryParams{
						SpaceID: "space", EnvironmentID: "environment", EntryID: "entry", XContentfulVersion: cm.NewOptInt(entry.Sys.Version),
					})
					require.NoError(t, putErr)
					require.IsType(t, &cm.EntryStatusCode{}, response)
					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "an unrefreshed external draft must not authorize publication")

					return nil
				},
			},
			{
				PreConfig:          recorder.reset,
				Config:             config("two"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				PreConfig: func() {
					options.Plan.NoRefresh = false

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
		},
	})
}

func TestAccEntryResourceRefreshResolvesAmbiguousPublishSuccess(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryCommittedPublishFailureAdapter{delegate: server}
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
			ExpectError: regexp.MustCompile(`Failed to publish entry`),
		},
		{
			PreConfig: recorder.reset,
			Config:    config("two"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: func(state *terraform.State) error {
				requireNoEntryMutations(t, recorder)
				require.NoError(t, resource.TestCheckResourceAttr("contentful_entry.test", "published_version", "3")(state))

				return nil
			},
		},
	}})
}
