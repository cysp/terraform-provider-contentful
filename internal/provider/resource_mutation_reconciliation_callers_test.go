package provider_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strconv"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var (
	errMutationResourceAbsentFromPlan = errors.New("mutation resource was absent from the Terraform plan")
	errUnexpectedMutationTransition   = errors.New("unexpected mutation response recovery transition")
)

func TestAccEditorInterfaceUpdateMutationResponseContradictionRetainsTruthfulState(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetContentType("space", "master", "article", cm.ContentTypeRequestData{Name: "Article"})
	server.SetEditorInterface("space", "master", "article", cm.EditorInterfaceData{
		EditorLayout: cm.NewOptNilEditorInterfaceEditorLayoutItemArray([]cm.EditorInterfaceEditorLayoutItem{
			cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(cm.EditorInterfaceEditorLayoutGroupItem{
				GroupId: "response",
				Name:    "Group",
				Items:   []cm.EditorInterfaceEditorLayoutItem{},
			}),
		}),
	})

	adapter := &mutationJSONResponseAdapter{delegate: server}
	config := func(groupID string) string {
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
	groupIDPath := tfjsonpath.New("editor_layout").AtSliceIndex(0).AtMapKey("group").AtMapKey("group_id")

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				Config:             config("response"),
				ResourceName:       "contentful_editor_interface.test",
				ImportState:        true,
				ImportStateId:      "space/master/article",
				ImportStatePersist: true,
			},
			{
				PreConfig: func() {
					adapter.arm(http.MethodPut, func(response map[string]any) {
						response["editorLayout"] = []any{map[string]any{
							"groupId": "response",
							"name":    "Group",
							"items":   []any{},
						}}
					})
				},
				Config:      config("planned"),
				ExpectError: regexp.MustCompile(`editor_layout response differed meaningfully from the Terraform plan`),
			},
			{
				Config: config("planned"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionUpdate),
					expectMutationPlanTransition{
						address:   "contentful_editor_interface.test",
						valuePath: groupIDPath,
						before:    "response",
						after:     "planned",
					},
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_editor_interface.test", groupIDPath, knownvalue.StringExact("planned")),
				},
			},
		},
	})
}

func TestAccRoleCreateMutationResponseContradictionRetainsTaintedState(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	adapter := &mutationJSONResponseAdapter{delegate: server}
	config := `
resource "contentful_role" "test" {
  space_id = "space"
  name     = "Role"
  permissions = {
    Entry = ["read"]
  }
  policies = []
}
`
	permissionPath := tfjsonpath.New("permissions").AtMapKey("Entry").AtSliceIndex(0)

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					adapter.arm(http.MethodPost, func(response map[string]any) {
						response["permissions"] = map[string]any{"Entry": []any{"manage"}}
					})
				},
				Config:      config,
				ExpectError: regexp.MustCompile(`permissions response differed meaningfully from the Terraform plan`),
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_role.test", plancheck.ResourceActionReplace),
					expectMutationPlanTransition{
						address:   "contentful_role.test",
						valuePath: permissionPath,
						before:    "manage",
						after:     "read",
					},
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_role.test", permissionPath, knownvalue.StringExact("read")),
				},
			},
		},
	})
}

func TestAccRoleUpdateMutationResponseContradictionRetainsTruthfulState(t *testing.T) {
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
				Config: config("read"),
			},
			{
				PreConfig: func() {
					adapter.arm(http.MethodPut, func(response map[string]any) {
						response["permissions"] = map[string]any{"Entry": []any{"manage"}}
					})
				},
				Config:      config("create"),
				ExpectError: regexp.MustCompile(`permissions response differed meaningfully from the Terraform plan`),
			},
			{
				Config: config("create"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_role.test", plancheck.ResourceActionUpdate),
					expectMutationPlanTransition{
						address:   "contentful_role.test",
						valuePath: permissionPath,
						before:    "manage",
						after:     "create",
					},
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_role.test", permissionPath, knownvalue.StringExact("create")),
				},
			},
		},
	})
}

type mutationJSONResponseAdapter struct {
	delegate http.Handler
	mu       sync.Mutex
	armed    bool
	method   string
	mutate   func(map[string]any)
}

func (a *mutationJSONResponseAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	recorder := httptest.NewRecorder()
	a.delegate.ServeHTTP(recorder, request)

	body := recorder.Body.Bytes()
	if mutate, ok := a.responseMutation(request.Method, recorder.Code); ok {
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

	for key, values := range recorder.Header() {
		for _, value := range values {
			responseWriter.Header().Add(key, value)
		}
	}

	responseWriter.Header().Set("Content-Length", strconv.Itoa(len(body)))
	responseWriter.WriteHeader(recorder.Code)
	_, _ = io.Copy(responseWriter, bytes.NewReader(body))
}

func (a *mutationJSONResponseAdapter) arm(method string, mutate func(map[string]any)) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.armed = true
	a.method = method
	a.mutate = mutate
}

func (a *mutationJSONResponseAdapter) responseMutation(method string, statusCode int) (func(map[string]any), bool) {
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, false
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.armed || method != a.method {
		return nil, false
	}

	a.armed = false

	return a.mutate, true
}

type expectMutationPlanTransition struct {
	address   string
	valuePath tfjsonpath.Path
	before    any
	after     any
}

func (check expectMutationPlanTransition) CheckPlan(_ context.Context, request plancheck.CheckPlanRequest, response *plancheck.CheckPlanResponse) {
	for _, change := range request.Plan.ResourceChanges {
		if change.Address != check.address {
			continue
		}

		before, err := tfjsonpath.Traverse(change.Change.Before, check.valuePath)
		if err != nil {
			response.Error = fmt.Errorf("read response-derived value from plan: %w", err)

			return
		}

		after, err := tfjsonpath.Traverse(change.Change.After, check.valuePath)
		if err != nil {
			response.Error = fmt.Errorf("read configured value from plan: %w", err)

			return
		}

		if !reflect.DeepEqual(before, check.before) || !reflect.DeepEqual(after, check.after) {
			response.Error = fmt.Errorf("%w: before=%#v after=%#v, want before=%#v after=%#v", errUnexpectedMutationTransition, before, after, check.before, check.after)
		}

		return
	}

	response.Error = fmt.Errorf("%w: %s", errMutationResourceAbsentFromPlan, check.address)
}
