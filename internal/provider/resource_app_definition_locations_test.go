package provider_test

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccAppDefinitionResourceRejectsDiscardedLocationAttributes(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ location, name, attribute string }{
		{"dialog", "field_types", `field_types = [{ type = "Symbol" }]`},
		{"dialog", "navigation_item", `navigation_item = { name = "Page", path = "/page" }`},
		{"experience-sidebar", "field_types", `field_types = [{ type = "Symbol" }]`},
		{"experience-sidebar", "navigation_item", `navigation_item = { name = "Page", path = "/page" }`},
		{"component-sidebar", "field_types", `field_types = [{ type = "Symbol" }]`},
		{"component-sidebar", "navigation_item", `navigation_item = { name = "Page", path = "/page" }`},
	} {
		t.Run(test.location+"/"+test.name, func(t *testing.T) {
			t.Parallel()

			var mutations atomic.Int64

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost || r.Method == http.MethodPut {
					mutations.Add(1)
				}

				t.Errorf("invalid configuration reached Contentful: %s %s", r.Method, r.URL.Path)
				http.Error(w, "unexpected request", http.StatusInternalServerError)
			})
			config := fmt.Sprintf(`
resource "contentful_app_definition" "test" {
 organization_id = "org"
 name = "App"
 src = "https://example.invalid/app"
 locations = [{ location = "app-config" }, { location = %q, %s }]
}
`, test.location, test.attribute)
			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{{Config: config, ExpectError: regexp.MustCompile(regexp.QuoteMeta(test.name + ` cannot be configured for location "` + test.location + `"`))}}})
			require.Zero(t, mutations.Load())
		})
	}
}

func TestAccAppDefinitionResourceNestedLocations(t *testing.T) {
	t.Parallel()

	cases := []struct{ config, body string }{
		{`[
 { location = "entry-field", field_types = [
   { type = "Array", items = { type = "Symbol" } },
   { type = "Array", items = { type = "Link", link_type = "Asset" } },
   { type = "Link", link_type = "Entry" }
 ] },
 { location = "page", navigation_item = { name = "Page", path = "/page" } },
 { location = "dialog" }
]`, `[{"location":"entry-field","fieldTypes":[{"type":"Array","items":{"type":"Symbol"}},{"type":"Array","items":{"type":"Link","linkType":"Asset"}},{"type":"Link","linkType":"Entry"}]},{"location":"page","navigationItem":{"name":"Page","path":"/page"}},{"location":"dialog"}]`},
		{`[
 { location = "entry-field", field_types = [
   { type = "Symbol" },
   { type = "Array", items = { type = "Link", link_type = "Asset" } },
   { type = "Link", link_type = "Entry" }
 ] },
 { location = "page" },
 { location = "dialog" }
]`, `[{"location":"entry-field","fieldTypes":[{"type":"Symbol"},{"type":"Array","items":{"type":"Link","linkType":"Asset"}},{"type":"Link","linkType":"Entry"}]},{"location":"page"},{"location":"dialog"}]`},
		{`[
 { location = "dialog" },
 { location = "page" },
 { location = "entry-field", field_types = [
   { type = "Link", link_type = "Entry" },
   { type = "Symbol" }
 ] }
]`, `[{"location":"dialog"},{"location":"page"},{"location":"entry-field","fieldTypes":[{"type":"Link","linkType":"Entry"},{"type":"Symbol"}]}]`},
	}

	var (
		expected, current    atomic.Pointer[string]
		posts, puts, deletes atomic.Int64
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		assert.Empty(t, r.Header.Get("X-Contentful-Version"))

		switch r.Method {
		case http.MethodPost, http.MethodPut:
			request, err := io.ReadAll(r.Body)
			if !assert.NoError(t, err) {
				return
			}

			assert.JSONEq(t, *expected.Load(), string(request))
			body := `{"sys":{"type":"AppDefinition","id":"app","organization":{"sys":{"type":"Link","linkType":"Organization","id":"org"}}},` + (*expected.Load())[1:]
			current.Store(&body)

			if r.Method == http.MethodPost {
				posts.Add(1)
				assert.Equal(t, "/organizations/org/app_definitions", r.URL.Path)
				w.WriteHeader(http.StatusCreated)
			} else {
				puts.Add(1)
				assert.Equal(t, "/organizations/org/app_definitions/app", r.URL.Path)
			}

			fmt.Fprint(w, body)
		case http.MethodGet:
			assert.Equal(t, "/organizations/org/app_definitions/app", r.URL.Path)
			fmt.Fprint(w, *current.Load())
		case http.MethodDelete:
			assert.Equal(t, "/organizations/org/app_definitions/app", r.URL.Path)
			deletes.Add(1)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected method %s", r.Method)
			http.Error(w, "unexpected request", http.StatusInternalServerError)
		}
	})

	steps := make([]resource.TestStep, 0, len(cases)+1)
	for index, test := range cases {
		config := fmt.Sprintf(`
resource "contentful_app_definition" "test" {
 organization_id = "org"
 name = "App"
 src = "https://example.invalid/app"
 locations = %s
}
output "locations" { value = jsonencode(contentful_app_definition.test.locations) }
`, test.config)

		steps = append(steps, resource.TestStep{PreConfig: func() {
			body := `{"name":"App","src":"https://example.invalid/app","locations":` + test.body + `}`
			expected.Store(&body)
		}, Config: config})
		if index == 0 {
			steps = append(steps, resource.TestStep{ResourceName: "contentful_app_definition.test", ImportState: true, ImportStateVerify: true})
		}
	}

	testAccMockedResource(t, handler, resource.TestCase{Steps: steps})
	assert.EqualValues(t, 1, posts.Load())
	assert.EqualValues(t, 2, puts.Load())
	assert.EqualValues(t, 1, deletes.Load())
}
