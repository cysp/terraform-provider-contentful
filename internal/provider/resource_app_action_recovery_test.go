package provider_test

import (
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
				contradict   atomic.Bool
				requestMutex sync.Mutex
				methods      []string
			)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					requestMutex.Lock()

					methods = append(methods, r.Method)
					requestMutex.Unlock()
				}

				response := httptest.NewRecorder()
				server.ServeHTTP(response, r)
				maps.Copy(w.Header(), response.Header())

				body := response.Body.String()
				if r.Method == mutation.method && contradict.Swap(false) {
					// Only the response differs: refresh cannot supply the recovery checkpoint.
					body = strings.Replace(body, `"name":"Planned"`, `"name":"Returned"`, 1)
				}

				w.WriteHeader(response.Code)
				_, err := io.WriteString(w, body)
				assert.NoError(t, err)
			})

			var steps []resource.TestStep
			if mutation.method == http.MethodPut {
				steps = append(steps, resource.TestStep{Config: actionLegacyConfig})
			}

			planned := strings.Replace(actionLegacyConfig, `name = "Action"`, `name = "Planned"`, 1)
			steps = append(steps, resource.TestStep{
				PreConfig:   func() { contradict.Store(true) },
				Config:      planned,
				ExpectError: regexp.MustCompile("Contentful returned a different name"),
			}, resource.TestStep{
				Config: planned,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("name"), knownvalue.StringExact("Returned"))},
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile(`^org/app/[^/]+$`)))},
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
				assert.Equal(t, []string{http.MethodPost, http.MethodDelete, http.MethodPost, http.MethodDelete}, methods)
			} else {
				assert.Equal(t, []string{http.MethodPost, http.MethodPut, http.MethodPut, http.MethodDelete}, methods)
			}

			result, err := server.Handler().GetAppActions(t.Context(), cm.GetAppActionsParams{OrganizationID: "org", AppDefinitionID: "app"})
			require.NoError(t, err)

			actions, ok := result.(*cm.AppActionCollection)
			require.True(t, ok)
			assert.Empty(t, actions.Items, "recovery and destroy must not orphan actions")
		})
	}
}
