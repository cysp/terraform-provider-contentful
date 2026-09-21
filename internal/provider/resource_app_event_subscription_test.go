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
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	appEventResourceAddress = "contentful_app_event_subscription.test"
	appEventResourcePath    = "/organizations/organization/app_definitions/app/event_subscription"
	appEventConfigFile      = "testdata/app_event_subscription/main.tf"
)

const appEventHTTPBody = `{"topics":["Asset.publish","Entry.publish"],"targetUrl":"https://example.invalid/events"}`

func TestAccAppEventSubscriptionResourceInvalidTarget(t *testing.T) {
	t.Parallel()

	var count atomic.Int64

	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { count.Add(1) })
	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{{
		ConfigFile: config.StaticFile(appEventConfigFile),
		ConfigVariables: config.Variables{"subscription": config.ObjectVariable(map[string]config.Variable{
			"target_url": config.StringVariable("http://example.invalid/events"),
		})},
		ExpectError: regexp.MustCompile("Invalid app event target URL"),
	}}})
	assert.Zero(t, count.Load())
}

// TestAccAppEventSubscriptionResourceLifecycle tests the provider against a
// replacement fixture. Function clearing and switching are mock assumptions;
// this test is deliberately never selected as live CMA acceptance.
func TestAccAppEventSubscriptionResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "App"})
	_, err = server.Handler().PutAppEventSubscription(t.Context(), &cm.AppEventSubscriptionData{
		Topics: []string{"Entry.publish", "Asset.publish"}, TargetUrl: cm.NewOptString("https://example.invalid/events"),
	}, cm.PutAppEventSubscriptionParams{OrganizationID: "organization", AppDefinitionID: "app"})
	require.NoError(t, err)

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

			var topics []string
			if !assert.NoError(t, json.Unmarshal(body["topics"], &topics)) {
				return
			}

			if !slices.Contains(topics, "FutureEntity.futureAction") {
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
	cases := []struct {
		subscription map[string]config.Variable
		body         string
	}{
		{
			subscription: map[string]config.Variable{
				"target_url":                 config.StringVariable("https://example.invalid/events"),
				"filter_function_id":         config.StringVariable("filter"),
				"transformation_function_id": config.StringVariable("transform"),
			},
			body: `{"topics":["Asset.publish","Entry.publish"],"targetUrl":"https://example.invalid/events","functions":{"filter":{"sys":{"type":"Link","linkType":"Function","id":"filter"}},"transformation":{"sys":{"type":"Link","linkType":"Function","id":"transform"}}}}`,
		},
		{
			subscription: map[string]config.Variable{
				"target_url":                 config.StringVariable("https://example.invalid/events"),
				"transformation_function_id": config.StringVariable("transform"),
			},
			body: `{"topics":["Asset.publish","Entry.publish"],"targetUrl":"https://example.invalid/events","functions":{"transformation":{"sys":{"type":"Link","linkType":"Function","id":"transform"}}}}`,
		},
		{
			subscription: map[string]config.Variable{
				"filter_function_id":         config.StringVariable("filter"),
				"transformation_function_id": config.StringVariable("transform"),
				"handler_function_id":        config.StringVariable("handler"),
			},
			body: `{"topics":["Asset.publish","Entry.publish"],"functions":{"filter":{"sys":{"type":"Link","linkType":"Function","id":"filter"}},"transformation":{"sys":{"type":"Link","linkType":"Function","id":"transform"}},"handler":{"sys":{"type":"Link","linkType":"Function","id":"handler"}}}}`,
		},
		{
			subscription: map[string]config.Variable{
				"filter_function_id":  config.StringVariable("filter"),
				"handler_function_id": config.StringVariable("handler"),
			},
			body: `{"topics":["Asset.publish","Entry.publish"],"functions":{"filter":{"sys":{"type":"Link","linkType":"Function","id":"filter"}},"handler":{"sys":{"type":"Link","linkType":"Function","id":"handler"}}}}`,
		},
		{
			subscription: map[string]config.Variable{
				"handler_function_id": config.StringVariable("handler"),
			},
			body: `{"topics":["Asset.publish","Entry.publish"],"functions":{"handler":{"sys":{"type":"Link","linkType":"Function","id":"handler"}}}}`,
		},
		{
			subscription: map[string]config.Variable{
				"target_url":         config.StringVariable("https://example.invalid/events"),
				"filter_function_id": config.StringVariable("filter"),
			},
			body: `{"topics":["Asset.publish","Entry.publish"],"targetUrl":"https://example.invalid/events","functions":{"filter":{"sys":{"type":"Link","linkType":"Function","id":"filter"}}}}`,
		},
		{
			subscription: map[string]config.Variable{
				"target_url": config.StringVariable("https://example.invalid/events"),
			},
			body: appEventHTTPBody,
		},
		// A synthetic future topic tests provider extensibility, not CMA acceptance.
		{
			subscription: map[string]config.Variable{
				"target_url": config.StringVariable("https://example.invalid/events"),
				"topics":     config.ListVariable(config.StringVariable("FutureEntity.futureAction")),
			},
			body: `{"topics":["FutureEntity.futureAction"],"targetUrl":"https://example.invalid/events"}`,
		},
	}

	steps := make([]resource.TestStep, 0, len(cases)+5)
	steps = append(steps,
		resource.TestStep{
			// Apply identity import before exercising updates to the singleton.
			ConfigDirectory: config.StaticDirectory("testdata/app_event_subscription"),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("id"), knownvalue.StringExact("organization/app")),
				statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("target_url"), knownvalue.StringExact("https://example.invalid/events")),
				statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("filter_function_id"), knownvalue.Null()),
				statecheck.ExpectIdentity(appEventResourceAddress, map[string]knownvalue.Check{
					"organization_id":   knownvalue.StringExact("organization"),
					"app_definition_id": knownvalue.StringExact("app"),
				}),
			},
		},
		resource.TestStep{ResourceName: appEventResourceAddress, ImportState: true, ImportStateId: "organization/app", ImportStateVerify: true},
		resource.TestStep{ConfigFile: config.StaticFile(appEventConfigFile), ConfigVariables: config.Variables{"subscription": config.ObjectVariable(map[string]config.Variable{
			"target_url": config.StringVariable("https://example.invalid/events"),
			"topics":     config.ListVariable(config.StringVariable("Asset.publish"), config.StringVariable("Entry.publish"), config.StringVariable("Entry.publish")),
		})}, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		resource.TestStep{ConfigFile: config.StaticFile(appEventConfigFile), ConfigVariables: config.Variables{"subscription": config.ObjectVariable(map[string]config.Variable{
			"target_url": config.StringVariable("https://example.invalid/events"),
			"timeouts":   config.ObjectVariable(map[string]config.Variable{"update": config.StringVariable("30s")}),
		})}},
	)

	for _, test := range cases {
		steps = append(steps, resource.TestStep{ConfigFile: config.StaticFile(appEventConfigFile), ConfigVariables: config.Variables{"subscription": config.ObjectVariable(test.subscription)}, ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("id"), knownvalue.StringExact("organization/app"))}})
	}

	steps = append(steps, resource.TestStep{PreConfig: func() {
		_, deleteErr := server.Handler().DeleteAppEventSubscription(context.Background(), cm.DeleteAppEventSubscriptionParams{OrganizationID: "organization", AppDefinitionID: "app"})
		require.NoError(t, deleteErr)
	}, ConfigFile: config.StaticFile(appEventConfigFile), ConfigVariables: config.Variables{"subscription": config.ObjectVariable(cases[len(cases)-1].subscription)}, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appEventResourceAddress, plancheck.ResourceActionCreate)}}})
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
				steps = append(steps, resource.TestStep{ConfigFile: config.StaticFile(appEventConfigFile), PlanOnly: true, ExpectNonEmptyPlan: true}, resource.TestStep{ConfigFile: config.StaticFile(appEventConfigFile), ResourceName: appEventResourceAddress, ImportState: true, ImportStateId: "organization/app", ImportStatePersist: true, ImportStateCheck: testAccImportAttributes(map[string]string{"target_url": "https://example.invalid/previous"})})
			}

			steps = append(steps, resource.TestStep{ConfigFile: config.StaticFile(appEventConfigFile), ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("target_url"), knownvalue.StringExact("https://example.invalid/events")), statecheck.ExpectSensitiveValue(appEventResourceAddress, tfjsonpath.New("target_url"))}}, resource.TestStep{ConfigFile: config.StaticFile(appEventConfigFile), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}})
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
		{ConfigFile: config.StaticFile(appEventConfigFile)},
		{PreConfig: func() {
			_, deleteErr := server.Handler().DeleteAppDefinition(context.Background(), cm.DeleteAppDefinitionParams{OrganizationID: "organization", AppDefinitionID: "app"})
			require.NoError(t, deleteErr)
		}, ConfigFile: config.StaticFile(appEventConfigFile), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appEventResourceAddress, plancheck.ResourceActionCreate)}}, ExpectError: regexp.MustCompile(`Failed to upsert app event subscription`)},
		{PreConfig: func() { server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "Restored app"}) }, ConfigFile: config.StaticFile(appEventConfigFile), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appEventResourceAddress, plancheck.ResourceActionCreate)}}},
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

				raw := rec.Body.Bytes()

				if r.Method == http.MethodPut && contradict.Swap(false) {
					// Change only the response: a GET cannot supply this checkpoint.
					var body map[string]json.RawMessage
					if !assert.NoError(t, json.Unmarshal(raw, &body)) {
						return
					}

					body["targetUrl"] = json.RawMessage(`"https://example.invalid/returned"`)
					encoded, err := json.Marshal(body)
					assert.NoError(t, err)

					raw = encoded
				}

				w.WriteHeader(rec.Code)
				_, writeErr := w.Write(raw)
				assert.NoError(t, writeErr)
			})

			steps := []resource.TestStep{}
			if update {
				steps = append(steps, resource.TestStep{ConfigFile: config.StaticFile(appEventConfigFile)})
			}

			planned := config.Variables{"subscription": config.ObjectVariable(map[string]config.Variable{
				"target_url": config.StringVariable("https://example.invalid/planned"),
			})}
			steps = append(steps, resource.TestStep{
				PreConfig: func() { contradict.Store(true) }, ConfigFile: config.StaticFile(appEventConfigFile), ConfigVariables: planned,
				ExpectError: regexp.MustCompile("Contentful returned a different target_url"),
			})

			action := plancheck.ResourceActionDestroyBeforeCreate
			if update {
				action = plancheck.ResourceActionUpdate
			}

			steps = append(steps, resource.TestStep{ConfigFile: config.StaticFile(appEventConfigFile), ConfigVariables: planned, ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					testAccPriorState{check: statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("target_url"), knownvalue.StringExact("https://example.invalid/returned"))},
					testAccPriorState{check: statecheck.ExpectKnownValue(appEventResourceAddress, tfjsonpath.New("id"), knownvalue.StringExact("organization/app"))},
					plancheck.ExpectResourceAction(appEventResourceAddress, action),
				},
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
			fixture := appEventConfigFile
			action := plancheck.ResourceActionDestroyBeforeCreate

			if createBeforeDestroy {
				fixture = "testdata/TestAccAppEventSubscriptionResourceParentReplacement/create_before_destroy.tf"
				action = plancheck.ResourceActionCreateBeforeDestroy
			}

			replacement := config.Variables{"subscription": config.ObjectVariable(map[string]config.Variable{
				"organization_id": config.StringVariable("other"), "app_definition_id": config.StringVariable("replacement"),
				"target_url": config.StringVariable("https://example.invalid/events"),
			})}
			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
				{ConfigFile: config.StaticFile(fixture)},
				{ConfigFile: config.StaticFile(fixture), ConfigVariables: replacement, ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appEventResourceAddress, action)},
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
