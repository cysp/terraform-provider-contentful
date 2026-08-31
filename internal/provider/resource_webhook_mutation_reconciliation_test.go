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
	"regexp"
	"strconv"
	"sync"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var (
	errUnexpectedWebhookFilterTransition = errors.New("unexpected webhook filter transition")
	errWebhookAbsentFromPlan             = errors.New("contentful_webhook.test was absent from the Terraform plan")
)

func TestAccWebhookMutationResponseContradictionRetainsTruthfulState(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	adapter := &webhookMutationResponseAdapter{delegate: server}
	config := func(name, filterType string) string {
		return fmt.Sprintf(`
resource "contentful_webhook" "test" {
  space_id = "space"
  name     = %q
  url      = "https://example.com/webhook"
  topics   = []
  filters = [{
    equals = {
      doc   = "sys.type"
      value = %q
    }
  }]
}
`, name, filterType)
	}

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				Config: config("Initial webhook", "Asset"),
			},
			{
				PreConfig: func() {
					adapter.armFilter(http.MethodPut, "Asset")
				},
				Config:      config("Updated webhook", "Entry"),
				ExpectError: regexp.MustCompile(`filters response differed meaningfully from the Terraform plan`),
			},
			{
				Config: config("Updated webhook", "Entry"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
					expectWebhookFilterPlanTransition{before: "Asset", after: "Entry"},
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_webhook.test",
						tfjsonpath.New("filters").AtSliceIndex(0).AtMapKey("equals").AtMapKey("value"),
						knownvalue.StringExact("Entry"),
					),
				},
			},
		},
	})
}

func TestAccWebhookCreateMutationResponseContradictionRetainsTaintedState(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	adapter := &webhookMutationResponseAdapter{delegate: server}
	config := `
resource "contentful_webhook" "test" {
  space_id = "space"
  name     = "Created webhook"
  url      = "https://example.com/webhook"
  topics   = []
  filters = [{
    equals = {
      doc   = "sys.type"
      value = "Entry"
    }
  }]
}
`

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					adapter.armFilter(http.MethodPost, "Asset")
				},
				Config:      config,
				ExpectError: regexp.MustCompile(`filters response differed meaningfully from the Terraform plan`),
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionReplace),
					expectWebhookFilterPlanTransition{before: "Asset", after: "Entry"},
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_webhook.test",
						tfjsonpath.New("filters").AtSliceIndex(0).AtMapKey("equals").AtMapKey("value"),
						knownvalue.StringExact("Entry"),
					),
				},
			},
		},
	})
}

func TestAccWebhookMutationReconciliationUsesEffectivePlanWithIgnoreChanges(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	config := func(name, filterType string, ignoreFilters bool) string {
		lifecycle := ""
		if ignoreFilters {
			lifecycle = "lifecycle { ignore_changes = [filters] }"
		}

		return fmt.Sprintf(`
resource "contentful_webhook" "test" {
  space_id = "space"
  name     = %q
  url      = "https://example.com/webhook"
  topics   = []
  filters = [{
    equals = {
      doc   = "sys.type"
      value = %q
    }
  }]
  %s
}
`, name, filterType, lifecycle)
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: config("Initial webhook", "Asset", false),
		},
		{
			Config: config("Updated webhook", "Entry", true),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(
					"contentful_webhook.test",
					tfjsonpath.New("filters").AtSliceIndex(0).AtMapKey("equals").AtMapKey("value"),
					knownvalue.StringExact("Asset"),
				),
			},
		},
		{
			Config: config("Updated webhook", "Entry", true),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
			}},
		},
	}})
}

type webhookMutationResponseAdapter struct {
	delegate http.Handler
	mu       sync.Mutex
	armed    bool
	method   string
	value    string
}

func (a *webhookMutationResponseAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	recorder := httptest.NewRecorder()
	a.delegate.ServeHTTP(recorder, request)

	body := recorder.Body.Bytes()
	if replacement, ok := a.responseFilterReplacement(request, recorder.Code); ok {
		var response map[string]any

		err := json.Unmarshal(body, &response)
		if err == nil {
			response["filters"] = []any{map[string]any{
				"equals": []any{map[string]any{"doc": "sys.type"}, replacement},
			}}

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

func (a *webhookMutationResponseAdapter) armFilter(method, value string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.armed = true
	a.method = method
	a.value = value
}

func (a *webhookMutationResponseAdapter) responseFilterReplacement(request *http.Request, statusCode int) (string, bool) {
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return "", false
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.armed || request.Method != a.method {
		return "", false
	}

	a.armed = false

	return a.value, true
}

type expectWebhookFilterPlanTransition struct {
	before string
	after  string
}

func (check expectWebhookFilterPlanTransition) CheckPlan(_ context.Context, request plancheck.CheckPlanRequest, response *plancheck.CheckPlanResponse) {
	valuePath := tfjsonpath.New("filters").AtSliceIndex(0).AtMapKey("equals").AtMapKey("value")

	for _, change := range request.Plan.ResourceChanges {
		if change.Address != "contentful_webhook.test" {
			continue
		}

		before, err := tfjsonpath.Traverse(change.Change.Before, valuePath)
		if err != nil {
			response.Error = fmt.Errorf("read prior webhook filter from plan: %w", err)

			return
		}

		after, err := tfjsonpath.Traverse(change.Change.After, valuePath)
		if err != nil {
			response.Error = fmt.Errorf("read planned webhook filter: %w", err)

			return
		}

		if before != check.before || after != check.after {
			response.Error = fmt.Errorf("%w: before=%#v after=%#v, want before=%q after=%q", errUnexpectedWebhookFilterTransition, before, after, check.before, check.after)
		}

		return
	}

	response.Error = errWebhookAbsentFromPlan
}
