package provider_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var errUnexpectedTerraformPlan = errors.New("unexpected Terraform plan")

func editorInterfaceLayoutConfig(groupID string) string {
	return fmt.Sprintf(`
resource "contentful_editor_interface" "test" {
  space_id        = "space"
  environment_id  = "master"
  content_type_id = "article"
  editor_layout = [{
    group = {
      group_id = %q
      name     = "Group"
      items    = []
    }
  }]
}
`, groupID)
}

func replaceEditorInterfaceResponseLayout(groupID string) func(map[string]any) {
	return func(response map[string]any) {
		response["editorLayout"] = []any{map[string]any{
			"groupId": groupID,
			"name":    "Group",
			"items":   []any{},
		}}
	}
}

func TestAccEditorInterfaceResourceUpdateConsistencyErrorRetainsResponseState(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetContentType("space", "master", "article", cm.ContentTypeRequestData{Name: "Article"})
	server.SetEditorInterface("space", "master", "article", cm.EditorInterfaceData{
		EditorLayout: cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{
			editorLayoutResponseGroup("response", "Group"),
		}),
	})

	adapter := &mutationJSONResponseAdapter{delegate: server}
	groupIDPath := tfjsonpath.New("editor_layout").AtSliceIndex(0).AtMapKey("group").AtMapKey("group_id")

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				Config:             editorInterfaceLayoutConfig("response"),
				ResourceName:       "contentful_editor_interface.test",
				ImportState:        true,
				ImportStateId:      "space/master/article",
				ImportStatePersist: true,
			},
			{
				PreConfig: func() {
					adapter.mutateNext(http.MethodPut, replaceEditorInterfaceResponseLayout("response"))
				},
				Config:      editorInterfaceLayoutConfig("planned"),
				ExpectError: regexp.MustCompile(`Contentful returned a different Editor Interface layout`),
			},
			{
				Config: editorInterfaceLayoutConfig("planned"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionUpdate),
					expectResponseStateInPlan{
						address:   "contentful_editor_interface.test",
						valuePath: groupIDPath,
						value:     "response",
					},
				}},
			},
		},
	})
}

func TestAccEditorInterfaceResourceCreateConsistencyErrorRetainsStateAndRequiresImport(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetContentType("space", "master", "article", cm.ContentTypeRequestData{Name: "Article"})
	server.SetEditorInterface("space", "master", "article", cm.EditorInterfaceData{})

	adapter := &mutationJSONResponseAdapter{delegate: server}
	handler := &editorInterfaceRequestRecorder{next: adapter}
	groupIDPath := tfjsonpath.New("editor_layout").AtSliceIndex(0).AtMapKey("group").AtMapKey("group_id")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					adapter.mutateNext(http.MethodPut, replaceEditorInterfaceResponseLayout("response"))
				},
				Config:      editorInterfaceLayoutConfig("planned"),
				ExpectError: regexp.MustCompile(`Contentful returned a different Editor Interface layout`),
			},
			{
				Config:      editorInterfaceLayoutConfig("planned"),
				ExpectError: regexp.MustCompile(`Editor Interface requires import`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionReplace),
					expectResponseStateInPlan{
						address:   "contentful_editor_interface.test",
						valuePath: groupIDPath,
						value:     "response",
					},
				}},
			},
		},
	})

	require.Equal(t, []string{"PUT:1", "PUT:1"}, handler.Requests())
}

func TestAccRoleResourceConsistencyErrorsRetainResponseState(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	adapter := &mutationJSONResponseAdapter{delegate: server}
	config := func(permission string) string {
		return fmt.Sprintf(`
resource "contentful_role" "test" {
  space_id = "space"
  name     = "Role"
  permissions = {
    Entry = [%q]
  }
  policies = []
}
`, permission)
	}
	permissionPath := tfjsonpath.New("permissions").AtMapKey("Entry").AtSliceIndex(0)

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					adapter.mutateNext(http.MethodPost, func(response map[string]any) {
						response["permissions"] = map[string]any{"Entry": []any{"manage"}}
					})
				},
				Config:      config("read"),
				ExpectError: regexp.MustCompile(`Contentful returned different role permissions`),
			},
			{
				Config: config("read"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_role.test", plancheck.ResourceActionReplace),
					expectResponseStateInPlan{
						address:   "contentful_role.test",
						valuePath: permissionPath,
						value:     "manage",
					},
				}},
			},
			{
				PreConfig: func() {
					adapter.mutateNext(http.MethodPut, func(response map[string]any) {
						response["permissions"] = map[string]any{"Entry": []any{"manage"}}
					})
				},
				Config:      config("create"),
				ExpectError: regexp.MustCompile(`Contentful returned different role permissions`),
			},
			{
				Config: config("create"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_role.test", plancheck.ResourceActionUpdate),
					expectResponseStateInPlan{
						address:   "contentful_role.test",
						valuePath: permissionPath,
						value:     "manage",
					},
				}},
			},
		},
	})
}

type mutationJSONResponseAdapter struct {
	delegate http.Handler
	mu       sync.Mutex
	method   string
	mutate   func(map[string]any)
}

func (a *mutationJSONResponseAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	recorder := httptest.NewRecorder()
	a.delegate.ServeHTTP(recorder, request)

	body := recorder.Body.Bytes()
	if mutate, ok := a.takeMutation(request.Method, recorder.Code); ok {
		var response map[string]any

		err := json.Unmarshal(body, &response)
		if err == nil {
			mutate(response)

			encoded, marshalErr := json.Marshal(response)
			if marshalErr == nil {
				body = encoded
			}
		}
	}

	maps.Copy(responseWriter.Header(), recorder.Header())

	responseWriter.Header().Del("Content-Length")
	responseWriter.WriteHeader(recorder.Code)
	_, _ = responseWriter.Write(body)
}

func (a *mutationJSONResponseAdapter) mutateNext(method string, mutate func(map[string]any)) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.method = method
	a.mutate = mutate
}

func (a *mutationJSONResponseAdapter) takeMutation(method string, statusCode int) (func(map[string]any), bool) {
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, false
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.mutate == nil || method != a.method {
		return nil, false
	}

	mutate := a.mutate
	a.mutate = nil

	return mutate, true
}

type expectResponseStateInPlan struct {
	address            string
	valuePath          tfjsonpath.Path
	value              any
	nonEmptyBeforePath *tfjsonpath.Path
}

func (check expectResponseStateInPlan) CheckPlan(_ context.Context, request plancheck.CheckPlanRequest, response *plancheck.CheckPlanResponse) {
	for _, change := range request.Plan.ResourceChanges {
		if change.Address != check.address {
			continue
		}

		value, err := tfjsonpath.Traverse(change.Change.Before, check.valuePath)
		if err != nil {
			response.Error = fmt.Errorf("read response-derived value from plan: %w", err)

			return
		}

		if !reflect.DeepEqual(value, check.value) {
			response.Error = fmt.Errorf("%w: response-derived value is %#v, want %#v", errUnexpectedTerraformPlan, value, check.value)

			return
		}

		if check.nonEmptyBeforePath != nil {
			beforeValue, err := tfjsonpath.Traverse(change.Change.Before, *check.nonEmptyBeforePath)
			if err != nil {
				response.Error = fmt.Errorf("read response-derived value from prior state: %w", err)

				return
			}

			stringValue, ok := beforeValue.(string)
			if !ok || stringValue == "" {
				response.Error = fmt.Errorf("%w: prior state value at %s is %#v, want a nonempty string", errUnexpectedTerraformPlan, check.nonEmptyBeforePath.String(), beforeValue)
			}
		}

		return
	}

	response.Error = fmt.Errorf("%w: resource %s is absent", errUnexpectedTerraformPlan, check.address)
}
