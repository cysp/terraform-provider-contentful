package provider_test

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
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

func discoveryAcceptanceFixture(t *testing.T, name string) []byte {
	t.Helper()

	body, err := os.ReadFile("testdata/discovery/" + name + ".json")
	require.NoError(t, err)

	return body
}

func TestAccDiscoveryDataSourcesComposition(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	var space cm.Space
	require.NoError(t, json.Unmarshal(discoveryAcceptanceFixture(t, "space"), &space))
	space.Sys.ID = "space"
	server.SetSpace(space)
	server.RegisterSpaceEnvironment("space", "master")
	server.RegisterSpaceEnvironment("space", "target")
	server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space", "master", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("target")}))

	var locale cm.Locale
	require.NoError(t, json.Unmarshal(discoveryAcceptanceFixture(t, "locale"), &locale))
	locale.Sys.Environment = cm.NewEnvironmentLink("target")
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

	var locale cm.Locale
	require.NoError(t, json.Unmarshal(discoveryAcceptanceFixture(t, "locale"), &locale))

	nondefault := locale
	nondefault.Default = false
	second := locale
	second.Sys.ID = "item-a"
	second.Code = "fr-FR"

	for _, test := range []struct {
		name  string
		items []cm.Locale
		file  string
	}{
		{"missing default", []cm.Locale{nondefault}, "default.tf"},
		// Two defaults are an adversarial discovery response, not normal CMA state.
		{"ambiguous default", []cm.Locale{locale, second}, "default.tf"},
		{"case sensitive code", []cm.Locale{locale}, "code.tf"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(&cm.LocaleCollection{
				Sys:  cm.LocaleCollectionSys{Type: cm.LocaleCollectionSysTypeArray},
				Skip: cm.NewOptInt(0), Limit: cm.NewOptInt(100), Total: cm.NewOptInt(len(test.items)), Items: test.items,
			})
			require.NoError(t, err)

			var mutations atomic.Int64

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					mutations.Add(1)
					http.Error(w, "unexpected mutation", http.StatusBadRequest)

					return
				}

				assert.Equal(t, "/spaces/space/environments/master/locales", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write(body)
				assert.NoError(t, err)
			})

			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{{ConfigFile: config.StaticFile("testdata/TestAccLocalesDataSourceSelectionGuard/" + test.file), ExpectError: regexp.MustCompile("Exactly one selected Locale is required")}}})
			assert.Zero(t, mutations.Load())
		})
	}
}
