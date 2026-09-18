package provider_test

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	actionAddress        = "contentful_app_action.test"
	actionCollectionPath = "/organizations/org/app_definitions/app/actions"
	actionConfigPrefix   = `resource "contentful_app_action" "test" {
 organization_id = "org"
 app_definition_id = "app"
 name = "Action"
`
)

const actionLegacyConfig = actionConfigPrefix + `category = "Custom"
 type = "endpoint"
 url = "https://example.invalid/action"
 parameters = jsonencode([])
}`

const actionFunctionConfig = actionConfigPrefix + `category = "Custom"
 type = "function-invocation"
 function_id = "function"
 parameters = jsonencode([])
}`

const actionSchemaConfig = actionConfigPrefix + `category = "Custom"
 type = "endpoint"
 url = "https://example.invalid/action"
 parameters_schema = jsonencode({type="object", properties={message={type="string"}}})
}`

const actionBuiltinConfig = actionConfigPrefix + `category = "Entries.v1.0"
 type = "endpoint"
 url = "https://example.invalid/action"
}`

const (
	actionLegacyBody  = `{"name":"Action","category":"Custom","type":"endpoint","url":"https://example.invalid/action","parameters":[]}`
	actionSchemaBody  = `{"name":"Action","category":"Custom","type":"endpoint","url":"https://example.invalid/action","parametersSchema":{"type":"object","properties":{"message":{"type":"string"}}}}`
	actionBuiltinBody = `{"name":"Action","category":"Entries.v1.0","type":"endpoint","url":"https://example.invalid/action"}`
)

type actionMutationFixture struct{ config, request string }

func testActionLifecycle(t *testing.T, cases []actionMutationFixture, checks ...plancheck.PlanCheck) {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})

	var (
		requestMutex sync.Mutex
		requests     []string
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("X-Contentful-Version"))

		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			raw, err := io.ReadAll(r.Body)
			assert.NoError(t, err)

			r.Body = io.NopCloser(bytes.NewReader(raw))
			assert.Equal(t, "application/vnd.contentful.management.v1+json", r.Header.Get("Content-Type"))

			requestMutex.Lock()
			if len(requests) == 0 {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, actionCollectionPath, r.URL.Path)
			} else {
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Regexp(t, "^"+actionCollectionPath+"/[^/]+$", r.URL.Path)
			}

			requests = append(requests, string(raw))
			requestMutex.Unlock()
		}

		server.ServeHTTP(w, r)
	})

	identity := statecheck.CompareValue(compare.ValuesSame())

	steps := make([]resource.TestStep, 0, len(cases))
	for _, test := range cases {
		steps = append(steps, resource.TestStep{Config: test.config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: checks}, ConfigStateChecks: []statecheck.StateCheck{
			identity.AddStateValue(actionAddress, tfjsonpath.New("id")),
		}})
	}

	testAccMockedResource(t, handler, resource.TestCase{Steps: steps})
	requestMutex.Lock()
	defer requestMutex.Unlock()

	require.Len(t, requests, len(cases))

	for i, test := range cases {
		assert.JSONEq(t, test.request, requests[i])
	}

	response, err := server.Handler().GetAppActions(t.Context(), cm.GetAppActionsParams{OrganizationID: "org", AppDefinitionID: "app"})
	require.NoError(t, err)

	actions, ok := response.(*cm.AppActionCollection)
	require.True(t, ok)
	assert.Empty(t, actions.Items, "destroy must delete the stored action")
}
