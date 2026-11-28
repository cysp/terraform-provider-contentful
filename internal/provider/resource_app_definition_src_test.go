package provider_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccAppDefinitionResourceSourceUpgrade(t *testing.T) {
	const (
		address  = "contentful_app_definition.test"
		url      = "https://example.invalid/app"
		location = `[{ location = "app-config" }]`
		delegate = `lifecycle { ignore_changes = [src] }`
	)

	server := newTestCMAServer(t)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{
		Name: "App", Src: cm.NewOptString(url),
		Locations: []cm.AppDefinitionLocationsItem{{Location: "app-config"}},
	})

	type mutation struct {
		method, path, version string
		body                  string
		readErr               error
	}

	var (
		mutex     sync.Mutex
		mutations []mutation
	)

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			body, err := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body))

			mutex.Lock()

			mutations = append(mutations, mutation{
				method: r.Method, path: r.URL.Path, version: r.Header.Get("X-Contentful-Version"),
				body: string(body), readErr: err,
			})

			mutex.Unlock()
		}

		server.ServeHTTP(w, r)
	}))
	t.Cleanup(api.Close)
	t.Setenv("CONTENTFUL_URL", api.URL)
	t.Setenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN", "CFPAT-12345")
	t.Setenv(resource.EnvTfAccProviderNamespace, "cysp")

	config := func(name, locations, source string) string {
		return fmt.Sprintf(`
resource "contentful_app_definition" "test" {
  organization_id = "org"
  name            = %q
  locations       = %s
  %s
}
`, name, locations, source)
	}
	omitted := config("App", location, "")
	prior := map[string]resource.ExternalProvider{
		"contentful": {Source: "registry.terraform.io/cysp/contentful", VersionConstraint: "= 0.2.0"},
	}
	priorConfig := `terraform {
  required_providers {
    contentful = { source = "cysp/contentful", version = "= 0.2.0" }
  }
}
` + omitted
	candidate := makeTestAccProtoV6ProviderFactories(testProviderOptionsWithHTTPServer(api)...)
	update := plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)

	resource.Test(t, validateAcceptanceTestCase(t, resource.TestCase{Steps: []resource.TestStep{
		{
			ExternalProviders: prior, Config: priorConfig, ResourceName: address,
			ImportState: true, ImportStateId: "org/app", ImportStatePersist: true,
			ImportStateCheck: testAccImportAttributes(map[string]string{"src": url}),
		},
		{ExternalProviders: prior, Config: priorConfig, PlanOnly: true},
		{
			ProtoV6ProviderFactories: candidate, Config: omitted, PlanOnly: true, ExpectNonEmptyPlan: true,
			ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{
				update, plancheck.ExpectKnownValue(address, tfjsonpath.New("src"), knownvalue.Null()),
			}},
		},
		{ProtoV6ProviderFactories: candidate, Config: config("App", location, delegate), PlanOnly: true},
		{
			ProtoV6ProviderFactories: candidate, PlanOnly: true, ExpectNonEmptyPlan: true,
			Config: `resource "contentful_app_definition" "test" {
  organization_id = "another-org"
  name            = "App"
  locations       = []
  lifecycle { ignore_changes = [src] }
}`,
			ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(address, plancheck.ResourceActionDestroyBeforeCreate),
				plancheck.ExpectKnownValue(address, tfjsonpath.New("src"), knownvalue.Null()),
			}},
		},
		{
			ProtoV6ProviderFactories: candidate, Config: config("Renamed", location, delegate),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{update}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(address, tfjsonpath.New("src"), knownvalue.StringExact(url)),
			},
		},
		{ProtoV6ProviderFactories: candidate, Config: config("Renamed", location, `src = "https://example.invalid/app"`), PlanOnly: true},
		{
			ProtoV6ProviderFactories: candidate, Config: config("Renamed", location, `src = "https://example.invalid/new-app"`),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{update}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(address, tfjsonpath.New("src"), knownvalue.StringExact("https://example.invalid/new-app")),
			},
		},
		{
			ProtoV6ProviderFactories: candidate, Config: config("Renamed", "[]", ""),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{update}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(address, tfjsonpath.New("src"), knownvalue.Null()),
				statecheck.ExpectKnownValue(address, tfjsonpath.New("locations"), knownvalue.ListSizeExact(0)),
			},
		},
	}}))

	mutex.Lock()
	defer mutex.Unlock()

	require.Len(t, mutations, 4)

	for _, mutation := range mutations {
		require.NoError(t, mutation.readErr)
		assert.Equal(t, "/organizations/org/app_definitions/app", mutation.path)
		assert.Empty(t, mutation.version)
	}

	for _, update := range mutations[:3] {
		assert.Equal(t, http.MethodPut, update.method)
	}

	assert.JSONEq(t, testJSON(map[string]any{
		"name": "Renamed", "src": url, "locations": []any{map[string]any{"location": "app-config"}},
	}), mutations[0].body)
	assert.JSONEq(t, testJSON(map[string]any{
		"name": "Renamed", "src": "https://example.invalid/new-app", "locations": []any{map[string]any{"location": "app-config"}},
	}), mutations[1].body)
	assert.JSONEq(t, testJSON(map[string]any{"name": "Renamed", "locations": []any{}}), mutations[2].body)
	assert.Equal(t, http.MethodDelete, mutations[3].method)
}
