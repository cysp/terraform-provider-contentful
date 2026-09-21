package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
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

func discoveryAcceptanceFixture(t *testing.T, name string) string {
	t.Helper()

	body, err := os.ReadFile("testdata/discovery/" + name + ".json")
	require.NoError(t, err)

	return strings.TrimSpace(string(body))
}

func TestAccDiscoveryDataSourcesComposition(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	var space cm.Space
	require.NoError(t, json.Unmarshal([]byte(strings.ReplaceAll(discoveryAcceptanceFixture(t, "space"), "item-b", "space")), &space))
	server.SetSpace(space)
	server.RegisterSpaceEnvironment("space", "master")
	server.RegisterSpaceEnvironment("space", "target")
	server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space", "master", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("target")}))

	var locale cm.Locale
	require.NoError(t, json.Unmarshal([]byte(strings.ReplaceAll(discoveryAcceptanceFixture(t, "locale"), `"id":"master"`, `"id":"target"`)), &locale))
	server.SetLocale(locale)
	server.SetContentType("space", "target", "article", cm.ContentTypeRequestData{Name: "Article", Fields: []cm.ContentTypeRequestDataFieldsItem{}})
	testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{
		{ConfigDirectory: config.TestNameDirectory(), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectUnknownValue("data.contentful_space.test", tfjsonpath.New("space_id")), plancheck.ExpectUnknownValue("data.contentful_locale.test", tfjsonpath.New("code"))}}, ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("data.contentful_space.test", tfjsonpath.New("organization_id"), knownvalue.StringExact("org")),
			statecheck.ExpectKnownValue("data.contentful_spaces.test", tfjsonpath.New("spaces"), knownvalue.ListSizeExact(1)),
			statecheck.ExpectKnownValue("data.contentful_environment.test", tfjsonpath.New("environment_id"), knownvalue.StringExact("target")),
			statecheck.ExpectKnownValue("data.contentful_environment.test", tfjsonpath.New("aliased_environment_id"), knownvalue.Null()),
			statecheck.ExpectKnownValue("data.contentful_environments.test", tfjsonpath.New("environments"), knownvalue.ListSizeExact(2)),
			statecheck.ExpectKnownValue("data.contentful_environment_alias.test", tfjsonpath.New("target_environment_id"), knownvalue.StringExact("target")),
			statecheck.ExpectKnownValue("data.contentful_environment_aliases.test", tfjsonpath.New("environment_aliases"), knownvalue.ListSizeExact(1)),
			statecheck.ExpectKnownValue("data.contentful_locales.test", tfjsonpath.New("locales").AtSliceIndex(0).AtMapKey("locale_id"), knownvalue.StringExact("item-b")),
			statecheck.ExpectKnownValue("data.contentful_locale.test", tfjsonpath.New("code"), knownvalue.StringExact("en-GB")),
			statecheck.ExpectKnownValue("data.contentful_locale.test", tfjsonpath.New("fallback_code"), knownvalue.Null()),
			statecheck.ExpectKnownValue("contentful_entry.localized", tfjsonpath.New("fields").AtMapKey("title"), knownvalue.StringExact(`{"en-GB":"Hello"}`)),
		}},
		{ConfigDirectory: config.TestNameDirectory(), PlanOnly: true},
	}})
}

func TestAccLocalesDataSourceSelectionGuard(t *testing.T) {
	t.Parallel()

	locale := discoveryAcceptanceFixture(t, "locale")
	for _, test := range []struct {
		name      string
		items     []string
		predicate string
	}{
		{"missing default", []string{strings.Replace(locale, `"default":true`, `"default":false`, 1)}, `locale.default`},
		// Two defaults are an adversarial discovery response, not normal CMA state.
		{"ambiguous default", []string{locale, strings.NewReplacer("item-b", "item-a", "en-GB", "fr-FR").Replace(locale)}, `locale.default`},
		{"case sensitive code", []string{locale}, `locale.code == "en-gb"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var mutations atomic.Int64

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					mutations.Add(1)
					http.Error(w, "unexpected mutation", http.StatusBadRequest)

					return
				}

				assert.Equal(t, "/spaces/space/environments/master/locales", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_, err := fmt.Fprintf(w, `{"sys":{"type":"Array"},"skip":0,"limit":100,"total":%d,"items":[%s]}`, len(test.items), strings.Join(test.items, ","))
				assert.NoError(t, err)
			})
			configuration := fmt.Sprintf(`
data "contentful_locales" "test" {
 space_id = "space"
 environment_id = "master"
 lifecycle {
  postcondition {
   condition = length([for locale in self.locales : locale if %[1]s]) == 1
   error_message = "Exactly one selected Locale is required."
  }
 }
}
locals { selected = one([for locale in data.contentful_locales.test.locales : locale if %[1]s]) }
resource "contentful_entry" "test" {
 space_id = data.contentful_locales.test.space_id
 environment_id = data.contentful_locales.test.environment_id
 content_type_id = "article"
 fields = {title = jsonencode({(local.selected.code) = "Hello"})}
}
`, test.predicate)
			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{{Config: configuration, ExpectError: regexp.MustCompile("Exactly one selected Locale is required")}}})
			assert.Zero(t, mutations.Load())
		})
	}
}
