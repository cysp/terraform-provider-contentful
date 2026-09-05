package provider_test

import (
	"net/http"
	"regexp"
	"strconv"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccEntryResourceFailedPublishRecoversExactDraftWithoutRefresh(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryRejectedPublishAdapter{delegate: server}
	recorder.delegate = fault
	config := managedEntryConfig

	var (
		providerFactoryCalls atomic.Int64
		preUpdateVersion     int
		draftVersion         int
	)

	ContentfulProviderMockedResourceTestWithFactoryCounter(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{
			{Config: config("one")},
			{
				PreConfig: func() {
					entry := getTestEntry(t, server)
					preUpdateVersion = entry.Sys.Version
					require.Positive(t, preUpdateVersion)

					recorder.reset()
					fault.shot.arm()
				},
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`(?s)Failed to publish entry.*publication was\s+not\s+confirmed`),
			},
			{
				PreConfig: func() {
					update, publish := requireEntryUpdateThenPublish(t, recorder.snapshot())
					require.Equal(t, []string{strconv.Itoa(preUpdateVersion)}, update.versionValues)
					require.Empty(t, update.contentTypeValues)
					require.JSONEq(t, `{"fields":{"managed":{"en-US":"two"}},"metadata":{"concepts":[],"tags":[]}}`, string(update.body))

					entry := getTestEntry(t, server)
					draftVersion = entry.Sys.Version
					require.Greater(t, draftVersion, preUpdateVersion)
					require.Equal(t, []string{strconv.Itoa(draftVersion)}, publish.versionValues)
					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionUpdate),
				}},
				Check: func(state *terraform.State) error {
					requests := recorder.snapshot()
					require.Len(t, requests, 1, "recovery must not repeat the confirmed draft PUT")
					requireEntryPublish(t, requests[0], entryTestPublishPath)
					require.Equal(t, []string{strconv.Itoa(draftVersion)}, requests[0].versionValues)

					entry := getTestEntry(t, server)
					publishedVersion := entry.Sys.PublishedVersion.Or(0)
					require.Equal(t, draftVersion, publishedVersion)
					require.Greater(t, entry.Sys.Version, publishedVersion)
					require.NoError(t, resource.TestCheckResourceAttr(
						"contentful_entry.test", "published_version", strconv.Itoa(publishedVersion),
					)(state))
					require.GreaterOrEqual(t, providerFactoryCalls.Load(), int64(3), "recovery must survive provider restart and private-state serialization")

					return nil
				},
			},
		},
	}, &providerFactoryCalls)
}

func TestAccEntryResourceUpdateDoesNotRecreateExternallyDeletedEntryWithoutRefresh(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	config := managedEntryConfig

	var priorVersion int

	ContentfulProviderMockedResourceTest(t, fixture.recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{Config: config("one")},
			{
				PreConfig: func() {
					entry := getTestEntry(t, fixture.server)
					priorVersion = entry.Sys.Version
					require.Positive(t, priorVersion)

					unpublishResponse, err := fixture.server.Handler().UnpublishEntry(t.Context(), cm.UnpublishEntryParams{
						SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
					})
					require.NoError(t, err)
					require.IsType(t, &cm.Entry{}, unpublishResponse)

					deleteResponse, err := fixture.server.Handler().DeleteEntry(t.Context(), cm.DeleteEntryParams{
						SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
					})
					require.NoError(t, err)
					require.IsType(t, &cm.NoContent{}, deleteResponse)

					getResponse, err := fixture.server.Handler().GetEntry(t.Context(), cm.GetEntryParams{
						SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
					})
					require.NoError(t, err)

					status, ok := getResponse.(cm.StatusCodeResponse)
					require.True(t, ok)
					require.Equal(t, http.StatusNotFound, status.GetStatusCode())

					fixture.recorder.reset()
				},
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`(?s)Failed to update entry.*BadRequest.*provide a content type`),
			},
			{
				PreConfig: func() {
					require.NoError(t, fixture.recorder.handlerError())

					requests := fixture.recorder.snapshot()
					require.Len(t, requests, 1, "a rejected stale Update must not be followed by Publish or Delete")
					update := requests[0]
					requireEntryUpdate(t, update)
					require.Equal(t, []string{strconv.Itoa(priorVersion)}, update.versionValues)
					require.Empty(t, update.contentTypeValues)
					require.JSONEq(t, `{"fields":{"managed":{"en-US":"two"}},"metadata":{"concepts":[],"tags":[]}}`, string(update.body))

					getResponse, err := fixture.server.Handler().GetEntry(t.Context(), cm.GetEntryParams{
						SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
					})
					require.NoError(t, err)

					status, ok := getResponse.(cm.StatusCodeResponse)
					require.True(t, ok)
					require.Equal(t, http.StatusNotFound, status.GetStatusCode(), "the rejected Update must leave the remote target absent")

					fixture.recorder.reset()
				},
				Config:             config("two"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEntryResourceInitialPublishVersionMismatchRevokesAuthority(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	recorder := fixture.recorder
	fault := &entryVersionMismatchPublishAdapter{delegate: fixture.server}
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
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`(?s)VersionMismatch.*revoked publication authority`),
			},
			{
				PreConfig: func() {
					update, publish := requireEntryUpdateThenPublish(t, recorder.snapshot())
					require.Equal(t, "2", update.version)
					require.Equal(t, "3", publish.version)
					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(state *terraform.State) error {
					requireNoEntryMutations(t, recorder, "initial VersionMismatch must not schedule later recovery")
					require.NoError(t, resource.TestCheckResourceAttr("contentful_entry.test", "published_version", "1")(state))

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourceDraftRateLimitDoesNotCreatePublicationAuthority(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	recorder := fixture.recorder
	fault := &entryRateLimitAdapter{delegate: fixture.server, path: entryTestUpdatePath}
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
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`Failed to update entry`),
			},
			{
				PreConfig: func() {
					requests := recorder.snapshot()
					require.Len(t, requests, 1, "the draft 429 must not be retried or followed by publication")
					requireEntryUpdate(t, requests[0])
					recorder.reset()
				},
				Config: config("two"),
				Check: func(*terraform.State) error {
					update, publish := requireEntryUpdateThenPublish(t, recorder.snapshot())
					require.Equal(t, "2", update.version)
					require.Equal(t, "3", publish.version)

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourcePublicationRateLimitRetainsExactAuthority(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	recorder := fixture.recorder
	fault := &entryRateLimitAdapter{delegate: fixture.server, path: entryTestPublishPath}
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
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`Failed to publish entry`),
			},
			{
				PreConfig: func() {
					update, publish := requireEntryUpdateThenPublish(t, recorder.snapshot())
					require.Equal(t, "2", update.version)
					require.Equal(t, "3", publish.version)
					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionUpdate),
				}},
				Check: func(*terraform.State) error {
					requests := recorder.snapshot()
					require.Len(t, requests, 1, "recovery must submit only the retained lifecycle mutation")
					requireEntryPublish(t, requests[0], entryTestPublishPath)
					require.Equal(t, "3", requests[0].version)

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

func TestAccEntryResourceCreatePublishFailureRecoversExactDraft(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		config      string
		draftMethod string
		draftPath   string
	}{
		"generated ID": {
			config: `
resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "environment"
  content_type_id = "article"
  fields = { managed = jsonencode({ "en-US" = "one" }) }
}
`,
			draftMethod: http.MethodPost,
			draftPath:   entryTestCollectionPath,
		},
		"specified ID": {
			config:      managedEntryConfig("one"),
			draftMethod: http.MethodPut,
			draftPath:   entryTestUpdatePath,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fixture := newEntryAcceptanceFixture(t)
			recorder := fixture.recorder
			publishFailure := &entryRejectedPublishAdapter{delegate: fixture.server}
			versionFault := &entryVersionOffsetAdapter{
				delegate: publishFailure, offset: 3, errorSink: fixture.errorSink,
			}
			recorder.delegate = versionFault

			var (
				entryID      string
				draftVersion int
			)

			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps: []resource.TestStep{
					{
						PreConfig:          publishFailure.shot.arm,
						Config:             test.config,
						ExpectNonEmptyPlan: true,
						Check: func(state *terraform.State) error {
							require.NoError(t, resource.TestCheckResourceAttrWith("contentful_entry.test", "entry_id", func(value string) error {
								entryID = value

								return nil
							})(state))

							require.NotEmpty(t, entryID)

							requests := recorder.snapshot()
							require.Len(t, requests, 2, "Create must send one draft mutation and one Publish attempt")
							draft, publish := requests[0], requests[1]
							require.Equal(t, test.draftMethod, draft.method)
							require.Equal(t, test.draftPath, draft.path)
							require.Empty(t, draft.versionValues)
							require.Equal(t, []string{"article"}, draft.contentTypeValues)
							require.JSONEq(t, `{"fields":{"managed":{"en-US":"one"}},"metadata":{"concepts":[],"tags":[]}}`, string(draft.body))
							require.Positive(t, draft.contentLength)
							requireEntryPublish(t, publish, entryTestCollectionPath+"/"+entryID+"/published")

							observed := versionFault.snapshotObservation()
							draftVersion = observed.draftVersion
							require.Greater(t, draftVersion, 1, "the Create response must expose the injected arbitrary positive version")
							require.Equal(t, []string{strconv.Itoa(draftVersion)}, publish.versionValues)

							recorder.reset()

							return nil
						},
					},
					{
						Config: test.config,
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionUpdate),
						}},
						Check: func(state *terraform.State) error {
							requests := recorder.snapshot()
							require.Len(t, requests, 1, "Create recovery must not repeat the confirmed draft mutation or delete the draft")
							requireEntryPublish(t, requests[0], entryTestCollectionPath+"/"+entryID+"/published")
							require.Equal(t, []string{strconv.Itoa(draftVersion)}, requests[0].versionValues)
							require.NoError(t, resource.TestCheckResourceAttr("contentful_entry.test", "entry_id", entryID)(state))

							observed := versionFault.snapshotObservation()
							require.Equal(t, draftVersion, observed.publishedVersion)
							require.Greater(t, observed.responseVersion, observed.publishedVersion)
							require.NoError(t, resource.TestCheckResourceAttr(
								"contentful_entry.test", "published_version", strconv.Itoa(observed.publishedVersion),
							)(state))

							entry := getTestEntryForIDs(t, fixture.server, "space", "environment", entryID)
							require.Positive(t, entry.Sys.PublishedVersion.Or(0))
							require.Equal(t, "article", entry.Sys.ContentType.Sys.ID)
							require.JSONEq(t, `{"en-US":"one"}`, string(entry.Fields.Value["managed"]))

							recorder.reset()

							return nil
						},
					},
					{
						Config: test.config,
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
						}},
						Check: func(*terraform.State) error {
							requireNoEntryMutations(t, recorder, "recovered Entry Create must converge")

							return nil
						},
					},
				},
			})
		})
	}
}

func TestAccEntryResourceUpdateUsesExactArbitraryPositiveReturnedVersion(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	recorder := fixture.recorder
	versionFault := &entryVersionOffsetAdapter{
		delegate: fixture.server, jumpOnDraftUpdate: 40, errorSink: fixture.errorSink,
	}
	recorder.delegate = versionFault
	config := managedEntryConfig

	var preUpdateVersion int

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{Config: config("one")},
		{
			PreConfig: func() {
				preUpdateVersion = getTestEntry(t, fixture.server).Sys.Version
				require.Positive(t, preUpdateVersion)
				versionFault.resetObservation()
				recorder.reset()
			},
			Config: config("two"),
			Check: func(state *terraform.State) error {
				update, publish := requireEntryUpdateThenPublish(t, recorder.snapshot())
				require.Equal(t, []string{strconv.Itoa(preUpdateVersion)}, update.versionValues)
				require.Empty(t, update.contentTypeValues)
				require.JSONEq(t, `{"fields":{"managed":{"en-US":"two"}},"metadata":{"concepts":[],"tags":[]}}`, string(update.body))

				observed := versionFault.snapshotObservation()
				require.Greater(t, observed.draftVersion, preUpdateVersion)
				require.Equal(t, observed.draftVersion, observed.publishedVersion)
				require.Greater(t, observed.responseVersion, observed.publishedVersion)
				require.Equal(t, []string{strconv.Itoa(observed.draftVersion)}, publish.versionValues)
				require.NoError(t, resource.TestCheckResourceAttr(
					"contentful_entry.test", "published_version", strconv.Itoa(observed.publishedVersion),
				)(state))

				return nil
			},
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
	}})
}

func TestAccEntryResourceFailedPublishRecoversExactDraftAfterRefresh(t *testing.T) {
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
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionUpdate),
			}},
			Check: func(state *terraform.State) error {
				requests := recorder.snapshot()
				require.Len(t, requests, 1, "refresh recovery must not repeat the confirmed draft PUT")
				requireEntryPublish(t, requests[0], entryTestPublishPath)
				require.Equal(t, "3", requests[0].version)
				require.NoError(t, resource.TestCheckResourceAttr("contentful_entry.test", "published_version", "3")(state))

				return nil
			},
		},
	}})
}

func TestAccEntryResourceHigherPostPublishVersionIsAccepted(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryHigherPostPublishVersionAdapter{delegate: server, server: server, errorSink: fixture.errorSink}
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

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
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
					requireNoEntryMutations(t, recorder, "the contradictory tuple must revoke publication authority immediately")

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourceRecoveryContradictionRevokesAuthority(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	rejected := &entryRejectedPublishAdapter{delegate: fixture.server}
	responseVersion, publishedVersion := 3, 1
	contradictory := &entryPublishTupleAdapter{
		delegate: rejected, version: &responseVersion, publishedVersion: &publishedVersion, errorSink: fixture.errorSink,
	}
	fixture.recorder.delegate = contradictory
	recorder := fixture.recorder
	config := managedEntryConfig

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{Config: config("one")},
			{
				PreConfig: func() {
					recorder.reset()
					rejected.shot.arm()
				},
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`Failed to publish entry`),
			},
			{
				PreConfig: func() {
					recorder.reset()
					contradictory.shot.arm()
				},
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`(?s)checkpointed the returned state and revoked\s+exact-version publication authority`),
			},
			{
				PreConfig: func() {
					requests := recorder.snapshot()
					require.Len(t, requests, 1)
					requireEntryPublish(t, requests[0], entryTestPublishPath)
					require.Equal(t, "3", requests[0].version)
					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "a contradictory decoded success must revoke recovery authority")

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourceChangedDraftVersionMismatchRevokesPendingAuthority(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	rejected := &entryRejectedPublishAdapter{delegate: fixture.server}
	versionMismatch := &entryUpdateVersionMismatchAdapter{
		delegate: rejected, server: fixture.server, errorSink: fixture.errorSink,
	}
	fixture.recorder.delegate = versionMismatch
	recorder := fixture.recorder
	config := managedEntryConfig

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{Config: config("one")},
			{
				PreConfig: func() {
					recorder.reset()
					rejected.shot.arm()
				},
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`Failed to publish entry`),
			},
			{
				PreConfig: func() {
					recorder.reset()
					versionMismatch.shot.arm()
				},
				Config:      config("three"),
				ExpectError: regexp.MustCompile(`VersionMismatch`),
			},
			{
				PreConfig: func() {
					requests := recorder.snapshot()
					require.Len(t, requests, 1, "the stale draft update must not be retried or followed by Publish")
					requireEntryUpdate(t, requests[0])
					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "draft VersionMismatch must revoke the pre-existing marker")

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourceMarkedReadRevokesEqualPublicationTupleAuthority(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	rejected := &entryRejectedPublishAdapter{delegate: fixture.server}
	version, publishedVersion := 3, 3
	malformedRead := &entryPublicationTupleReadAdapter{
		delegate: rejected, version: &version, publishedVersion: &publishedVersion, errorSink: fixture.errorSink,
	}
	fixture.recorder.delegate = malformedRead
	recorder := fixture.recorder
	config := managedEntryConfig
	cliOptions := &resource.AdditionalCLIOptions{}

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: cliOptions,
		Steps: []resource.TestStep{
			{Config: config("one")},
			{
				PreConfig: func() {
					recorder.reset()
					rejected.shot.arm()
				},
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`Failed to publish entry`),
			},
			{
				PreConfig: func() {
					recorder.reset()
					malformedRead.shot.arm()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "malformed Read must revoke authority without mutation")

					return nil
				},
			},
			{
				PreConfig: func() {
					cliOptions.Plan.NoRefresh = true

					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "malformed Read must revoke authority before returning an error")

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourceNonAdvancingPublishResponseRevokesAuthority(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	recorder := fixture.recorder
	responseVersion := 3
	publishedVersion := 3
	fault := &entryPublishTupleAdapter{
		delegate: fixture.server, version: &responseVersion, publishedVersion: &publishedVersion, errorSink: fixture.errorSink,
	}
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
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`(?s)revoked\s+exact-version publication authority.*current version must be greater`),
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

func TestAccEntryResourcePublicationTupleChangeAtMarkedVersionRevokesAuthority(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	recorder := fixture.recorder
	failure := &entryRejectedPublishAdapter{delegate: fixture.server}
	fault := &entryPublicationTupleReadAdapter{delegate: failure, errorSink: fixture.errorSink}
	recorder.delegate = fault
	config := managedEntryConfig
	cliOptions := &resource.AdditionalCLIOptions{}

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: cliOptions,
		Steps: []resource.TestStep{
			{Config: config("one")},
			{
				PreConfig: func() {
					recorder.reset()
					failure.shot.arm()
				},
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`Failed to publish entry`),
			},
			{
				PreConfig: func() {
					recorder.reset()
					fault.shot.arm()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_entry.test", tfjsonpath.New("published_version"), knownvalue.Null()),
				},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "a changed publication tuple at the marked version must revoke without mutation")

					return nil
				},
			},
			{
				PreConfig: func() {
					cliOptions.Plan.NoRefresh = true

					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "revoked tuple authority must stay cleared with refresh disabled")

					return nil
				},
			},
		},
	})
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
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_entry.test", tfjsonpath.New("published_version"), knownvalue.Null()),
			},
			Check: func(*terraform.State) error {
				requireNoEntryMutations(t, recorder, "external unpublish must not grant publication authority")

				return nil
			},
		},
	}})
}

func TestAccEntryResourceInterveningDraftRevokesRecoveryWithoutRefresh(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryRejectedPublishAdapter{delegate: server}
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
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionUpdate),
				}},
				ExpectError: regexp.MustCompile(`VersionMismatch`),
			},
			{
				PreConfig: func() {
					requests := recorder.snapshot()
					require.Len(t, requests, 1, "recovery must not write another draft")
					requireEntryPublish(t, requests[0], entryTestPublishPath)
					require.Equal(t, "3", requests[0].version, "recovery must never adopt the externally advanced version")

					entry := getTestEntry(t, server)
					require.Equal(t, 4, entry.Sys.Version)
					require.Equal(t, 1, entry.Sys.PublishedVersion.Or(0))
					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "VersionMismatch must revoke stale publication authority")

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

func TestAccEntryResourceRefreshDisabledRecoveryAfterCommittedPublishDoesNotPublishNewerVersion(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryCommittedPublishFailureAdapter{delegate: server}
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
				Config:      config("two"),
				ExpectError: regexp.MustCompile(`Failed to publish entry`),
			},
			{
				PreConfig: recorder.reset,
				Config:    config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionUpdate),
				}},
				ExpectError: regexp.MustCompile(`VersionMismatch`),
			},
			{
				PreConfig: func() {
					requests := recorder.snapshot()
					require.Len(t, requests, 1, "recovery must not repeat the draft PUT")
					requireEntryPublish(t, requests[0], entryTestPublishPath)
					require.Equal(t, "3", requests[0].version, "recovery must never adopt the post-publication version")

					entry := getTestEntry(t, server)
					require.Equal(t, 4, entry.Sys.Version)
					require.Equal(t, 3, entry.Sys.PublishedVersion.Or(0))
					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "VersionMismatch must revoke ambiguous publication authority")

					return nil
				},
			},
		},
	})
}

func TestAccEntryResourceExternalPublicationOfMarkedDraftClearsRecovery(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	fault := &entryRejectedPublishAdapter{delegate: server}
	recorder.delegate = fault
	config := managedEntryConfig
	additionalCLIOptions := &resource.AdditionalCLIOptions{}

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: additionalCLIOptions,
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
					pending := getTestEntry(t, server)
					require.Equal(t, 3, pending.Sys.Version)
					require.Equal(t, 1, pending.Sys.PublishedVersion.Or(0))

					response, publishErr := server.Handler().PublishEntry(t.Context(), cm.PublishEntryParams{
						SpaceID: "space", EnvironmentID: "environment", EntryID: "entry", XContentfulVersion: 3,
					})
					require.NoError(t, publishErr)
					require.IsType(t, &cm.EntryStatusCode{}, response)
					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_entry.test", "published_version", "3"),
					func(*terraform.State) error {
						requireNoEntryMutations(t, recorder, "refresh must observe external publication without replay")

						return nil
					},
				),
			},
			{
				PreConfig: func() {
					additionalCLIOptions.Plan.NoRefresh = true

					recorder.reset()
				},
				Config: config("two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
				}},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, recorder, "observing exact external publication must clear recovery authority")

					return nil
				},
			},
		},
	})
}
