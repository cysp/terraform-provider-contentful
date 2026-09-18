package provider_test

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccAppActionResourceRecoveryState(t *testing.T) {
	t.Parallel()

	for _, mutation := range []struct {
		name     string
		method   string
		recovery plancheck.ResourceActionType
	}{
		{"create", http.MethodPost, plancheck.ResourceActionDestroyBeforeCreate},
		{"update", http.MethodPut, plancheck.ResourceActionUpdate},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})

			var (
				contradict          atomic.Bool
				requestMutex        sync.Mutex
				requests            []string
				created             []string
				recoveryID          string
				recoveryMultipartID string
				recoveryIdentityID  string
			)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					requestMutex.Lock()

					requests = append(requests, r.Method+" "+r.URL.Path)
					requestMutex.Unlock()
				}

				response := httptest.NewRecorder()
				server.ServeHTTP(response, r)
				maps.Copy(w.Header(), response.Header())

				body := response.Body.String()

				actionID := appActionRecoveryResponseID(t, response.Body.Bytes())
				if r.Method == http.MethodPost && response.Code == http.StatusCreated {
					requestMutex.Lock()

					created = append(created, actionID)
					requestMutex.Unlock()
				}

				if r.Method == mutation.method && contradict.Swap(false) {
					// Only the response differs: refresh cannot supply the recovery checkpoint.
					body = strings.Replace(body, `"name":"Planned"`, `"name":"Returned","resultSchema":{"type":"string"}`, 1)
					body = strings.Replace(body, `"message":{"type":"string"}`, `"returned":{"type":"boolean"}`, 1)
					body = strings.Replace(body, `"id":"org"`, `"id":"wrong-org"`, 1)
					body = strings.Replace(body, `"id":"app"`, `"id":"wrong-app"`, 1)
					assert.Contains(t, body, `"id":"wrong-org"`)
					assert.Contains(t, body, `"id":"wrong-app"`)
					// Create has no requested action ID to pin; retain its generated ID.
					if mutation.method == http.MethodPut {
						body = strings.Replace(body, `"id":"`+actionID+`"`, `"id":"wrong-action"`, 1)
						assert.Contains(t, body, `"id":"wrong-action"`)
					}
				}

				w.WriteHeader(response.Code)
				_, err := io.WriteString(w, body)
				assert.NoError(t, err)
			})

			recordID := func(target *string) knownvalue.Check {
				return knownvalue.StringFunc(func(value string) error {
					requestMutex.Lock()
					*target = value
					requestMutex.Unlock()

					return nil
				})
			}

			var steps []resource.TestStep
			if mutation.method == http.MethodPut {
				steps = append(steps, resource.TestStep{Config: actionSchemaConfig})
			}

			planned := strings.Replace(actionSchemaConfig, `name = "Action"`, `name = "Planned"`, 1)
			steps = append(steps, resource.TestStep{
				PreConfig:   func() { contradict.Store(true) },
				Config:      planned,
				ExpectError: regexp.MustCompile("Contentful returned a different App Action identity"),
			}, resource.TestStep{
				Config: planned,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("name"), knownvalue.StringExact("Returned"))},
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("parameters_schema"), knownvalue.StringExact(`{"properties":{"returned":{"type":"boolean"}},"type":"object"}`))},
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("result_schema"), knownvalue.StringExact(`{"type":"string"}`))},
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("organization_id"), knownvalue.StringExact("org"))},
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("app_definition_id"), knownvalue.StringExact("app"))},
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("app_action_id"), recordID(&recoveryID))},
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("id"), recordID(&recoveryMultipartID))},
					testAccPriorState{check: statecheck.ExpectIdentity(actionAddress, map[string]knownvalue.Check{
						"organization_id": knownvalue.StringExact("org"), "app_definition_id": knownvalue.StringExact("app"), "app_action_id": recordID(&recoveryIdentityID),
					})},
					plancheck.ExpectResourceAction(actionAddress, mutation.recovery),
				}},
			})
			testAccMockedResource(t, handler, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps:                steps,
			})
			requestMutex.Lock()
			defer requestMutex.Unlock()

			if mutation.method == http.MethodPost {
				require.Len(t, created, 2)
				assert.Equal(t, []string{"POST " + actionCollectionPath, "DELETE " + actionCollectionPath + "/" + created[0], "POST " + actionCollectionPath, "DELETE " + actionCollectionPath + "/" + created[1]}, requests)
			} else {
				require.Len(t, created, 1)
				assert.Equal(t, []string{"POST " + actionCollectionPath, "PUT " + actionCollectionPath + "/" + created[0], "PUT " + actionCollectionPath + "/" + created[0], "DELETE " + actionCollectionPath + "/" + created[0]}, requests)
			}

			assert.Equal(t, created[0], recoveryID, "persisted action ID must remain the original target")
			assert.Equal(t, "org/app/"+created[0], recoveryMultipartID)
			assert.Equal(t, created[0], recoveryIdentityID)

			result, err := server.Handler().GetAppActions(t.Context(), cm.GetAppActionsParams{OrganizationID: "org", AppDefinitionID: "app"})
			require.NoError(t, err)

			actions, ok := result.(*cm.AppActionCollection)
			require.True(t, ok)
			assert.Empty(t, actions.Items, "recovery and destroy must not orphan actions")
		})
	}
}

func appActionRecoveryResponseID(t *testing.T, body []byte) string {
	t.Helper()

	if len(body) == 0 {
		return ""
	}

	var action struct {
		Sys struct {
			ID string `json:"id"`
		} `json:"sys"`
	}
	if !assert.NoError(t, json.Unmarshal(body, &action)) { //nolint:testifylint // Called from an HTTP handler goroutine.
		return ""
	}

	assert.NotEmpty(t, action.Sys.ID)

	return action.Sys.ID
}

func TestAccAppActionResourceAmbiguousMutation(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, method, diagnostic string
		truncate                 bool
		methods                  []string
	}{
		{"create lost", http.MethodPost, "Failed to create app action", false, []string{http.MethodPost, http.MethodDelete}},
		{"create truncated", http.MethodPost, "Failed to create app action", true, []string{http.MethodPost, http.MethodDelete}},
		{"update lost", http.MethodPut, "Failed to update app action", false, []string{http.MethodPost, http.MethodPut, http.MethodDelete}},
		{"update truncated", http.MethodPut, "Failed to update app action", true, []string{http.MethodPost, http.MethodPut, http.MethodDelete}},
		{"delete lost", http.MethodDelete, "Failed to delete app action", false, []string{http.MethodPost, http.MethodDelete}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})
			fault := &appActionResponseFault{t: t, next: server, method: test.method, truncate: test.truncate}
			planned := strings.Replace(actionLegacyConfig, `name = "Action"`, `name = "Planned"`, 1)

			steps := make([]resource.TestStep, 0, 4)
			if test.method != http.MethodPost {
				steps = append(steps, resource.TestStep{Config: actionLegacyConfig})
			}

			failed := resource.TestStep{Config: planned, PreConfig: func() { fault.fail.Store(true) }, ExpectError: regexp.MustCompile(test.diagnostic)}
			if test.method == http.MethodDelete {
				failed.Config = actionLegacyConfig
				failed.Destroy = true
			}

			steps = append(steps, failed)

			verifyRemote := func() {
				result, err := server.Handler().GetAppActions(t.Context(), cm.GetAppActionsParams{OrganizationID: "org", AppDefinitionID: "app"})
				require.NoError(t, err)

				actions, ok := result.(*cm.AppActionCollection)
				require.True(t, ok)

				if test.method == http.MethodDelete {
					require.Empty(t, actions.Items)
				} else {
					require.Len(t, actions.Items, 1)
					require.Equal(t, "Planned", actions.Items[0].Name)
				}
			}
			if test.method == http.MethodPost {
				steps = append(steps, resource.TestStep{
					Config: planned, PreConfig: verifyRemote, ResourceName: actionAddress,
					ImportState: true, ImportStatePersist: true,
					ImportStateIdFunc: func(_ *terraform.State) (string, error) {
						fault.requestMutex.Lock()
						defer fault.requestMutex.Unlock()

						return "org/app/" + fault.actionID, nil
					},
				})
			}

			if test.method == http.MethodDelete {
				steps = append(steps, resource.TestStep{Config: actionLegacyConfig, PreConfig: verifyRemote, Destroy: true})
			} else {
				steps = append(steps, resource.TestStep{Config: planned, PreConfig: verifyRemote, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}})
			}

			testAccMockedResource(t, fault, resource.TestCase{Steps: steps})
			fault.requestMutex.Lock()
			defer fault.requestMutex.Unlock()

			assert.Equal(t, test.methods, fault.methods)
		})
	}
}
