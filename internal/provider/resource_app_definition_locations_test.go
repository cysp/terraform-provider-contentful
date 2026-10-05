package provider_test

import (
	"encoding/json"
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

	cases := []struct{ config, body, response string }{
		{`[
 { location = "entry-field", field_types = [
   { type = "Array", items = { type = "Symbol" } },
   { type = "Array", items = { type = "Link", link_type = "Asset" } },
   { type = "Link", link_type = "Entry" }
 ] },
 { location = "page", navigation_item = { name = "Page", path = "/page" } },
 { location = "dialog" }
]`, testJSON([]any{
			map[string]any{
				"location": "entry-field",
				"fieldTypes": []any{
					map[string]any{"type": "Array", "items": map[string]any{"type": "Symbol"}},
					map[string]any{"type": "Array", "items": map[string]any{"type": "Link", "linkType": "Asset"}},
					map[string]any{"type": "Link", "linkType": "Entry"},
				},
			},
			map[string]any{"location": "page", "navigationItem": map[string]any{"name": "Page", "path": "/page"}},
			map[string]any{"location": "dialog"},
		}), testJSON(map[string]any{"sys": map[string]any{"type": "AppDefinition", "id": "app", "organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}}}, "name": "App", "src": "https://example.invalid/app", "locations": []any{
			map[string]any{
				"location": "entry-field",
				"fieldTypes": []any{
					map[string]any{"type": "Array", "items": map[string]any{"type": "Symbol"}},
					map[string]any{"type": "Array", "items": map[string]any{"type": "Link", "linkType": "Asset"}},
					map[string]any{"type": "Link", "linkType": "Entry"},
				},
			},
			map[string]any{"location": "page", "navigationItem": map[string]any{"name": "Page", "path": "/page"}},
			map[string]any{"location": "dialog"},
		}})},
		{`[
 { location = "entry-field", field_types = [
   { type = "Symbol" },
   { type = "Array", items = { type = "Link", link_type = "Asset" } },
   { type = "Link", link_type = "Entry" }
 ] },
 { location = "page" },
 { location = "dialog" }
]`, testJSON([]any{
			map[string]any{
				"location": "entry-field",
				"fieldTypes": []any{
					map[string]any{"type": "Symbol"},
					map[string]any{"type": "Array", "items": map[string]any{"type": "Link", "linkType": "Asset"}},
					map[string]any{"type": "Link", "linkType": "Entry"},
				},
			},
			map[string]any{"location": "page"},
			map[string]any{"location": "dialog"},
		}), testJSON(map[string]any{"sys": map[string]any{"type": "AppDefinition", "id": "app", "organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}}}, "name": "App", "src": "https://example.invalid/app", "locations": []any{
			map[string]any{
				"location": "entry-field",
				"fieldTypes": []any{
					map[string]any{"type": "Symbol"},
					map[string]any{"type": "Array", "items": map[string]any{"type": "Link", "linkType": "Asset"}},
					map[string]any{"type": "Link", "linkType": "Entry"},
				},
			},
			map[string]any{"location": "page"},
			map[string]any{"location": "dialog"},
		}})},
		{`[
 { location = "dialog" },
 { location = "page" },
 { location = "entry-field", field_types = [
   { type = "Link", link_type = "Entry" },
   { type = "Symbol" }
 ] }
]`, testJSON([]any{
			map[string]any{"location": "dialog"},
			map[string]any{"location": "page"},
			map[string]any{
				"location":   "entry-field",
				"fieldTypes": []any{map[string]any{"type": "Link", "linkType": "Entry"}, map[string]any{"type": "Symbol"}},
			},
		}), testJSON(map[string]any{"sys": map[string]any{"type": "AppDefinition", "id": "app", "organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}}}, "name": "App", "src": "https://example.invalid/app", "locations": []any{
			map[string]any{"location": "dialog"},
			map[string]any{"location": "page"},
			map[string]any{
				"location":   "entry-field",
				"fieldTypes": []any{map[string]any{"type": "Link", "linkType": "Entry"}, map[string]any{"type": "Symbol"}},
			},
		}})},
	}

	var (
		expected, current, response atomic.Pointer[string]
		posts, puts, deletes        atomic.Int64
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
			body := *response.Load()
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
			body := testJSON(map[string]any{"name": "App", "src": "https://example.invalid/app", "locations": json.RawMessage(test.body)})
			expected.Store(&body)
			response.Store(&test.response)
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
