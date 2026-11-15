package provider_test

import (
	"net/http"
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccEntryResourceContentTypeContradictionPreservesRecoveryState(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name                   string
		update, refresh, retry bool
	}{
		{name: "create replacement"},
		{name: "update replacement", update: true},
		{name: "update refresh", update: true, refresh: true},
		{name: "update exact version retry", update: true, retry: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fixture := newEntryAcceptanceFixture(t)
			fixture.server.SetEntry("space", "environment", "remote", "unrelated", cm.EntryRequest{})
			rejected := &entryRejectedPublishAdapter{delegate: fixture.server}
			adapter := &mutationJSONResponseAdapter{delegate: rejected}
			fixture.recorder.delegate = adapter
			cliOptions := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}

			const address = "contentful_entry.test"

			var steps []resource.TestStep
			if test.update {
				// Leave version 1 pending so the contradictory draft must revoke
				// existing publication authority as well as grant none.
				steps = append(steps, resource.TestStep{
					PreConfig:          rejected.shot.arm,
					Config:             managedEntryConfig("prior"),
					ExpectNonEmptyPlan: true,
				})
			}

			steps = append(steps,
				resource.TestStep{
					PreConfig: func() {
						fixture.recorder.reset()
						adapter.mutateNext(http.MethodPut, func(response map[string]any) {
							sys, ok := response["sys"].(map[string]any)
							if !assert.True(t, ok) {
								return
							}

							sys["contentType"] = map[string]any{"sys": map[string]any{"type": "Link", "linkType": "ContentType", "id": "remote"}}
						})
					},
					Config:      managedEntryConfig("planned"),
					ExpectError: regexp.MustCompile(`Unexpected entry content type`),
				},
				resource.TestStep{
					PreConfig: func() {
						requests := fixture.recorder.snapshot()
						require.Len(t, requests, 1, "the contradictory draft must not be published")
						requireEntryUpdate(t, requests[0])

						if test.update {
							assert.Equal(t, []string{"1"}, requests[0].versionValues)
							assert.Empty(t, requests[0].contentTypeValues)
						} else {
							assert.Empty(t, requests[0].versionValues)
							assert.Equal(t, []string{"article"}, requests[0].contentTypeValues)
						}

						assert.JSONEq(t, testJSON(map[string]any{
							"fields":   map[string]any{"managed": map[string]any{"en-US": "planned"}},
							"metadata": map[string]any{"concepts": []any{}, "tags": []any{}},
						}), string(requests[0].body))
						fixture.recorder.reset()
					},
					Config:             managedEntryConfig("planned"),
					PlanOnly:           true,
					ExpectNonEmptyPlan: true,
					ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionReplace),
						expectResponseStateInPlan{address: address, valuePath: tfjsonpath.New("content_type_id"), value: "remote"},
						expectResponseStateInPlan{address: address, valuePath: tfjsonpath.New("id"), value: "space/environment/entry"},
						expectResponseStateInPlan{address: address, valuePath: tfjsonpath.New("fields").AtMapKey("managed"), value: testJSON(map[string]string{"en-US": "planned"})},
					}},
				},
			)

			if test.update {
				// Adopting the returned reference without refresh must not
				// publish using the previous draft's authority.
				steps = append(steps, resource.TestStep{
					Config: managedEntryConfigForContentType("remote", "planned"),
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionNoop),
					}},
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(address, tfjsonpath.New("content_type_id"), knownvalue.StringExact("remote")),
						statecheck.ExpectIdentity(address, map[string]knownvalue.Check{
							"space_id":       knownvalue.StringExact("space"),
							"environment_id": knownvalue.StringExact("environment"),
							"entry_id":       knownvalue.StringExact("entry"),
						}),
					},
					Check: func(*terraform.State) error {
						requireNoEntryMutations(t, fixture.recorder)

						return nil
					},
				})
			}

			switch {
			case test.retry:
				steps = append(steps,
					resource.TestStep{
						Config:      managedEntryConfigForContentType("remote", "next"),
						ExpectError: regexp.MustCompile(`Unexpected entry content type`),
					},
					resource.TestStep{
						PreConfig: func() {
							requests := fixture.recorder.snapshot()
							require.Len(t, requests, 1)
							requireEntryUpdate(t, requests[0])
							assert.Equal(t, []string{"2"}, requests[0].versionValues, "the confirmed endpoint's returned version must survive the contradiction")
							assert.Empty(t, requests[0].contentTypeValues)
							fixture.recorder.reset()
						},
						Config: managedEntryConfig("next"),
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(address, plancheck.ResourceActionNoop),
						}},
						Check: func(*terraform.State) error {
							requireNoEntryMutations(t, fixture.recorder)

							return nil
						},
					},
				)
			case test.refresh:
				steps = append(steps, resource.TestStep{
					PreConfig: func() { cliOptions.Plan.NoRefresh = false },
					Config:    managedEntryConfig("planned"),
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionNoop),
					}},
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(address, tfjsonpath.New("content_type_id"), knownvalue.StringExact("article")),
					},
					Check: func(*terraform.State) error {
						requireNoEntryMutations(t, fixture.recorder, "a confirming refresh must not recreate publication authority")

						return nil
					},
				})
			default:
				steps = append(steps, resource.TestStep{
					Config: managedEntryConfig("planned"),
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionReplace),
					}},
					Check: func(*terraform.State) error {
						requests := fixture.recorder.snapshot()
						require.Len(t, requests, 4)
						assert.Equal(t, http.MethodDelete, requests[0].method)
						assert.Equal(t, entryTestPublishPath, requests[0].path)
						assert.Equal(t, http.MethodDelete, requests[1].method)
						assert.Equal(t, entryTestUpdatePath, requests[1].path)
						requireEntryUpdate(t, requests[2])
						assert.Empty(t, requests[2].versionValues)
						assert.Equal(t, []string{"article"}, requests[2].contentTypeValues)
						requireEntryPublish(t, requests[3], entryTestPublishPath)
						assert.Equal(t, []string{"1"}, requests[3].versionValues)

						return nil
					},
				})
			}

			testAccMockedResource(t, fixture.recorder, resource.TestCase{
				AdditionalCLIOptions: cliOptions,
				Steps:                steps,
				CheckDestroy: func(*terraform.State) error {
					remote, err := fixture.server.Handler().GetEntry(t.Context(), cm.GetEntryParams{SpaceID: "space", EnvironmentID: "environment", EntryID: "entry"})
					require.NoError(t, err)

					missing, ok := remote.(*cm.ErrorStatusCode)
					require.True(t, ok)
					assert.Equal(t, http.StatusNotFound, missing.StatusCode)
					unrelated := getTestEntryForIDs(t, fixture.server, "space", "environment", "unrelated")
					assert.Equal(t, "remote", unrelated.Sys.ContentType.Sys.ID)
					assert.Equal(t, 1, unrelated.Sys.Version)
					require.NoError(t, fixture.recorder.handlerError())

					return nil
				},
			})
		})
	}
}
