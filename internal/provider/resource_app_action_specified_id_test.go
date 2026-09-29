package provider_test

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
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

func TestAccAppActionResourceSpecifiedID(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})

	var (
		requestMutex sync.Mutex
		requests     []struct{ method, path, body string }
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			body, readErr := io.ReadAll(r.Body)
			assert.NoError(t, readErr)

			r.Body = io.NopCloser(bytes.NewReader(body))

			requestMutex.Lock()

			requests = append(requests, struct{ method, path, body string }{r.Method, r.URL.Path, string(body)})
			requestMutex.Unlock()
		}

		server.ServeHTTP(w, r)
	})

	chosen := appActionFunctionWithID("chosenAction")
	second := appActionFunctionWithID("secondAction")
	checkID := func(id string) statecheck.StateCheck {
		return statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("app_action_id"), knownvalue.StringExact(id))
	}
	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{Config: chosen, ConfigStateChecks: []statecheck.StateCheck{checkID("chosenAction")}},
		{Config: chosen, ResourceName: actionAddress, ImportState: true, ImportStateVerify: true},
		{Config: actionFunctionConfig, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(actionAddress, plancheck.ResourceActionNoop),
		}}, ConfigStateChecks: []statecheck.StateCheck{checkID("chosenAction")}},
		{Config: appActionEndpointWithID("chosenAction"), ConfigStateChecks: []statecheck.StateCheck{checkID("chosenAction")}},
		{Config: second, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(actionAddress, plancheck.ResourceActionDestroyBeforeCreate),
		}}, ConfigStateChecks: []statecheck.StateCheck{checkID("secondAction")}},
	}})

	requestMutex.Lock()
	defer requestMutex.Unlock()

	require.Len(t, requests, 5)
	assert.Equal(t, http.MethodPost, requests[0].method)
	assert.Equal(t, actionCollectionPath, requests[0].path)
	assert.JSONEq(t, `{"id":"chosenAction","name":"Action","category":"Custom","type":"function-invocation","function":{"sys":{"type":"Link","linkType":"Function","id":"function"}},"parameters":[]}`, requests[0].body)
	assert.Equal(t, http.MethodPut, requests[1].method)
	assert.Equal(t, actionCollectionPath+"/chosenAction", requests[1].path)
	assert.JSONEq(t, actionLegacyBody, requests[1].body)
	assert.Equal(t, http.MethodDelete, requests[2].method)
	assert.Equal(t, actionCollectionPath+"/chosenAction", requests[2].path)
	assert.Equal(t, http.MethodPost, requests[3].method)
	assert.JSONEq(t, `{"id":"secondAction","name":"Action","category":"Custom","type":"function-invocation","function":{"sys":{"type":"Link","linkType":"Function","id":"function"}},"parameters":[]}`, requests[3].body)
	assert.Equal(t, http.MethodDelete, requests[4].method)
	assert.Equal(t, actionCollectionPath+"/secondAction", requests[4].path)
}

func TestAccAppActionResourceEndpointSpecifiedIDPassThrough(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})

	var (
		requestMutex sync.Mutex
		postBodies   []string
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == actionCollectionPath {
			body, readErr := io.ReadAll(r.Body)
			assert.NoError(t, readErr)

			r.Body = io.NopCloser(bytes.NewReader(body))

			requestMutex.Lock()

			postBodies = append(postBodies, string(body))
			requestMutex.Unlock()
		}

		server.ServeHTTP(w, r)
	})

	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{{
		Config:      appActionEndpointWithID("chosenAction"),
		ExpectError: regexp.MustCompile("Failed to create app action"),
	}}})

	requestMutex.Lock()
	defer requestMutex.Unlock()

	require.Len(t, postBodies, 1)
	assert.JSONEq(t, `{"id":"chosenAction","name":"Action","category":"Custom","type":"endpoint","url":"https://example.invalid/action","parameters":[]}`, postBodies[0])
}

func TestAccAppActionResourceSpecifiedIDCreateWithoutPreflight(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})
	_, err = server.Handler().CreateAppAction(t.Context(), &cm.AppActionCreateData{
		ID: cm.NewOptString("chosenAction"), Name: "Existing", Category: "Custom", Type: "function-invocation",
		Function: cm.NewOptFunctionLink(cm.NewFunctionLink("existingFunction")), Parameters: []byte(`[]`),
	}, cm.CreateAppActionParams{OrganizationID: "org", AppDefinitionID: "app"})
	require.NoError(t, err)

	var (
		requestMutex sync.Mutex
		requests     []string
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == actionCollectionPath || r.URL.Path == actionCollectionPath+"/chosenAction" {
			requestMutex.Lock()

			requests = append(requests, r.Method+" "+r.URL.Path)
			requestMutex.Unlock()
		}

		server.ServeHTTP(w, r)
	})
	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{Config: appActionFunctionWithID("chosenAction"), ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("app_action_id"), knownvalue.StringExact("chosenAction")),
			statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("name"), knownvalue.StringExact("Action")),
			statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("function_id"), knownvalue.StringExact("function")),
		}},
	}})

	requestMutex.Lock()
	defer requestMutex.Unlock()

	require.NotEmpty(t, requests)
	assert.Equal(t, http.MethodPost+" "+actionCollectionPath, requests[0], "Create must send POST without a prior GET")
}

func TestAccAppActionResourceAddSpecifiedID(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})

	testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{
		{Config: actionFunctionConfig},
		{Config: appActionFunctionWithID("chosenAction"), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(actionAddress, plancheck.ResourceActionDestroyBeforeCreate),
		}}, ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("app_action_id"), knownvalue.StringExact("chosenAction")),
		}},
	}})
}

func appActionFunctionWithID(id string) string {
	return actionFunctionConfig[:len(actionFunctionConfig)-1] + " app_action_id = \"" + id + "\"\n}"
}

func appActionEndpointWithID(id string) string {
	return strings.Replace(actionLegacyConfig, "\n}", "\n app_action_id = \""+id+"\"\n}", 1)
}
