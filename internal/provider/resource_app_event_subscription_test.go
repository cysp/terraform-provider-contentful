package provider_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
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
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const appEventResourceAddress = "contentful_app_event_subscription.test"
const appEventResourcePath = "/organizations/organization/app_definitions/app/event_subscription"
const appEventBaseConfig = `resource "contentful_app_event_subscription" "test" {
 organization_id = "organization"
 app_definition_id = "app"
 topics = ["Entry.publish", "Asset.publish"]
`
const appEventHTTPConfig = appEventBaseConfig + `target_url = "https://example.invalid/events"
}`
const appEventHTTPBody = `{"topics":["Asset.publish","Entry.publish"],"targetUrl":"https://example.invalid/events"}`

// TestAccAppEventSubscriptionResourceLifecycle tests the provider against a
// replacement fixture. Function clearing and switching are mock assumptions;
// this test is deliberately never selected as live CMA acceptance.
func TestAccAppEventSubscriptionResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "App"})

	var (
		requestMutex sync.Mutex
		puts         []string
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != appEventResourcePath {
			t.Errorf("unexpected path %s", r.URL.Path)
		}

		assert.Empty(t, r.Header.Get("X-Contentful-Version"))

		if r.Method == http.MethodPut {
			raw, readErr := io.ReadAll(r.Body)
			assert.NoError(t, readErr)

			r.Body = io.NopCloser(bytes.NewReader(raw))

			requestMutex.Lock()

			puts = append(puts, string(raw))
			requestMutex.Unlock()
			assert.Equal(t, "application/vnd.contentful.management.v1+json", r.Header.Get("Content-Type"))
		}

		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, r)
		maps.Copy(w.Header(), rec.Header())

		raw := rec.Body.Bytes()
		if (r.Method == http.MethodGet || r.Method == http.MethodPut) && rec.Code < 300 {
			var body map[string]json.RawMessage
			if !assert.NoError(t, json.Unmarshal(raw, &body)) {
				return
			}

			if !bytes.Contains(body["topics"], []byte("FutureEntity")) {
				// A fixed independently authored response order opposes sorted requests.
				body["topics"] = json.RawMessage(`["Entry.publish","Asset.publish"]`)
			}

			encoded, encodeErr := json.Marshal(body)
			assert.NoError(t, encodeErr)

			raw = encoded
		}

		w.WriteHeader(rec.Code)
		_, writeErr := w.Write(raw)
		assert.NoError(t, writeErr)
	})
	cases := []struct{ config, body string }{
		{appEventHTTPConfig, appEventHTTPBody},
		{appEventBaseConfig + `target_url = "https://example.invalid/events"
filter_function_id = "filter"
transformation_function_id = "transform"
}`, `{"topics":["Asset.publish","Entry.publish"],"targetUrl":"https://example.invalid/events","functions":{"filter":{"sys":{"type":"Link","linkType":"Function","id":"filter"}},"transformation":{"sys":{"type":"Link","linkType":"Function","id":"transform"}}}}`},
		{appEventBaseConfig + `target_url = "https://example.invalid/events"
transformation_function_id = "transform"
}`, `{"topics":["Asset.publish","Entry.publish"],"targetUrl":"https://example.invalid/events","functions":{"transformation":{"sys":{"type":"Link","linkType":"Function","id":"transform"}}}}`},
		{appEventBaseConfig + `filter_function_id = "filter"
transformation_function_id = "transform"
handler_function_id = "handler"
}`, `{"topics":["Asset.publish","Entry.publish"],"functions":{"filter":{"sys":{"type":"Link","linkType":"Function","id":"filter"}},"transformation":{"sys":{"type":"Link","linkType":"Function","id":"transform"}},"handler":{"sys":{"type":"Link","linkType":"Function","id":"handler"}}}}`},
		{appEventBaseConfig + `filter_function_id = "filter"
handler_function_id = "handler"
}`, `{"topics":["Asset.publish","Entry.publish"],"functions":{"filter":{"sys":{"type":"Link","linkType":"Function","id":"filter"}},"handler":{"sys":{"type":"Link","linkType":"Function","id":"handler"}}}}`},
		{appEventBaseConfig + `handler_function_id = "handler"
}`, `{"topics":["Asset.publish","Entry.publish"],"functions":{"handler":{"sys":{"type":"Link","linkType":"Function","id":"handler"}}}}`},
		{appEventBaseConfig + `target_url = "https://example.invalid/events"
filter_function_id = "filter"
}`, `{"topics":["Asset.publish","Entry.publish"],"targetUrl":"https://example.invalid/events","functions":{"filter":{"sys":{"type":"Link","linkType":"Function","id":"filter"}}}}`},
		{appEventHTTPConfig, appEventHTTPBody},
		// A synthetic future topic is a provider forward-compatibility oracle, not
		// an assertion that the real CMA accepts this topic.
		{strings.Replace(appEventHTTPConfig, `["Entry.publish", "Asset.publish"]`, `["FutureEntity.futureAction"]`, 1), `{"topics":["FutureEntity.futureAction"],"targetUrl":"https://example.invalid/events"}`},
	}

	steps := make([]resource.TestStep, 0, len(cases)+4)
	for index, test := range cases {
		steps = append(steps, resource.TestStep{Config: test.config, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}, ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("id"), knownvalue.StringExact("organization/app"))}})
		if index == 0 {
			steps = append(steps, resource.TestStep{ResourceName: appEventResourceAddress, ImportState: true, ImportStateId: "organization/app", ImportStateVerify: true})
			steps = append(steps, resource.TestStep{Config: strings.Replace(appEventHTTPConfig, `["Entry.publish", "Asset.publish"]`, `["Asset.publish", "Entry.publish", "Entry.publish"]`, 1), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}})
			steps = append(steps, resource.TestStep{Config: strings.Replace(appEventHTTPConfig, "\n}", "\ntimeouts = { update = \"30s\" }\n}", 1), ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}})
		}
	}

	steps = append(steps, resource.TestStep{PreConfig: func() {
		_, deleteErr := server.Handler().DeleteAppEventSubscription(context.Background(), cm.DeleteAppEventSubscriptionParams{OrganizationID: "organization", AppDefinitionID: "app"})
		require.NoError(t, deleteErr)
	}, Config: cases[len(cases)-1].config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appEventResourceAddress, plancheck.ResourceActionCreate)}}})
	testAccMockedResource(t, handler, resource.TestCase{Steps: steps})
	requestMutex.Lock()
	defer requestMutex.Unlock()

	require.Len(t, puts, len(cases)+1)

	for index, test := range cases {
		assert.JSONEq(t, test.body, puts[index])
	}

	assert.JSONEq(t, cases[len(cases)-1].body, puts[len(cases)])
	result, err := server.Handler().GetAppEventSubscription(t.Context(), cm.GetAppEventSubscriptionParams{OrganizationID: "organization", AppDefinitionID: "app"})
	require.NoError(t, err)

	status, ok := result.(cm.StatusCodeResponse)
	require.True(t, ok)
	assert.Equal(t, 404, status.GetStatusCode())
}

func TestAccAppEventSubscriptionResourceExistingSingleton(t *testing.T) {
	t.Parallel()

	for _, importFirst := range []bool{false, true} {
		t.Run(strconv.FormatBool(importFirst), func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "App"})
			_, err = server.Handler().PutAppEventSubscription(t.Context(), &cm.AppEventSubscriptionData{Topics: []string{"Entry.save"}, TargetUrl: cm.NewOptString("https://example.invalid/previous")}, cm.PutAppEventSubscriptionParams{OrganizationID: "organization", AppDefinitionID: "app"})
			require.NoError(t, err)

			var putCount atomic.Int64

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPut {
					putCount.Add(1)
				}

				server.ServeHTTP(w, r)
			})

			steps := []resource.TestStep{}
			if importFirst {
				steps = append(steps, resource.TestStep{Config: appEventHTTPConfig, PlanOnly: true, ExpectNonEmptyPlan: true}, resource.TestStep{Config: appEventHTTPConfig, ResourceName: appEventResourceAddress, ImportState: true, ImportStateId: "organization/app", ImportStatePersist: true, ImportStateCheck: testAccImportAttributes(map[string]string{"target_url": "https://example.invalid/previous"})})
			}

			steps = append(steps, resource.TestStep{Config: appEventHTTPConfig, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}, ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("target_url"), knownvalue.StringExact("https://example.invalid/events")), statecheck.ExpectSensitiveValue(appEventResourceAddress, tfjsonpath.New("target_url"))}}, resource.TestStep{Config: appEventHTTPConfig, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}})
			testAccMockedResource(t, handler, resource.TestCase{Steps: steps})
			assert.EqualValues(t, 1, putCount.Load())
		})
	}
}

func TestAccAppEventSubscriptionResourceParentDisappears(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "App"})

	var putCount atomic.Int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			putCount.Add(1)
		}

		server.ServeHTTP(w, r)
	})
	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{Config: appEventHTTPConfig},
		{PreConfig: func() {
			_, deleteErr := server.Handler().DeleteAppDefinition(context.Background(), cm.DeleteAppDefinitionParams{OrganizationID: "organization", AppDefinitionID: "app"})
			require.NoError(t, deleteErr)
		}, Config: appEventHTTPConfig, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appEventResourceAddress, plancheck.ResourceActionCreate)}}, ExpectError: regexp.MustCompile(`Failed to upsert app event subscription`)},
		{PreConfig: func() { server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "Restored app"}) }, Config: appEventHTTPConfig, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appEventResourceAddress, plancheck.ResourceActionCreate)}, PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
	}})
	assert.EqualValues(t, 3, putCount.Load())
}

func TestAccAppEventSubscriptionResourceRecoveryState(t *testing.T) {
	t.Parallel()

	for _, update := range []bool{false, true} {
		t.Run(strconv.FormatBool(update), func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "App"})

			var contradict atomic.Bool

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rec := httptest.NewRecorder()
				server.ServeHTTP(rec, r)
				maps.Copy(w.Header(), rec.Header())

				raw := rec.Body.String()
				if r.Method == http.MethodPut && contradict.Swap(false) {
					// Change only the response: a GET cannot supply this checkpoint.
					raw = strings.Replace(raw, "https://example.invalid/planned", "https://example.invalid/returned", 1)
				}

				w.WriteHeader(rec.Code)
				_, writeErr := io.WriteString(w, raw)
				assert.NoError(t, writeErr)
			})

			steps := []resource.TestStep{}
			if update {
				steps = append(steps, resource.TestStep{Config: appEventHTTPConfig})
			}

			planned := strings.Replace(appEventHTTPConfig, "/events", "/planned", 1)
			steps = append(steps, resource.TestStep{
				PreConfig: func() { contradict.Store(true) }, Config: planned,
				ExpectError: regexp.MustCompile("Contentful returned a different target_url"),
			})

			action := plancheck.ResourceActionDestroyBeforeCreate
			if update {
				action = plancheck.ResourceActionUpdate
			}

			steps = append(steps, resource.TestStep{Config: planned, ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					testAccPriorState{check: statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("target_url"), knownvalue.StringExact("https://example.invalid/returned"))},
					testAccPriorState{check: statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("id"), knownvalue.StringExact("organization/app"))},
					plancheck.ExpectResourceAction(appEventResourceAddress, action),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			}})
			testAccMockedResource(t, handler, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps:                steps,
			})
		})
	}
}

func TestAccAppEventSubscriptionResourceParentReplacement(t *testing.T) {
	t.Parallel()

	for _, createBeforeDestroy := range []bool{false, true} {
		t.Run(strconv.FormatBool(createBeforeDestroy), func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "App"})
			server.SetAppDefinition("other", "replacement", cm.AppDefinitionData{Name: "Replacement"})

			var (
				mutations     []string
				mutationMutex sync.Mutex
			)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					mutationMutex.Lock()

					mutations = append(mutations, r.Method+" "+r.URL.Path)
					mutationMutex.Unlock()
				}

				server.ServeHTTP(w, r)
			})
			initial := appEventHTTPConfig
			action := plancheck.ResourceActionDestroyBeforeCreate

			if createBeforeDestroy {
				initial = strings.Replace(initial, "\n}", "\nlifecycle { create_before_destroy = true }\n}", 1)
				action = plancheck.ResourceActionCreateBeforeDestroy
			}

			replacement := strings.NewReplacer(`"organization"`, `"other"`, `"app"`, `"replacement"`).Replace(initial)
			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
				{Config: initial},
				{Config: replacement, ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(appEventResourceAddress, action)},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				}, ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("id"), knownvalue.StringExact("other/replacement"))}},
			}})

			const (
				oldPath = "/organizations/organization/app_definitions/app/event_subscription"
				newPath = "/organizations/other/app_definitions/replacement/event_subscription"
			)

			want := []string{"PUT " + oldPath, "DELETE " + oldPath, "PUT " + newPath, "DELETE " + newPath}
			if createBeforeDestroy {
				want[1], want[2] = want[2], want[1]
			}

			mutationMutex.Lock()
			defer mutationMutex.Unlock()

			assert.Equal(t, want, mutations)
		})
	}
}
