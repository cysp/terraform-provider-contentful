package provider_test

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccAppDefinitionResourceImportedLocationLifecycle(t *testing.T) {
	t.Parallel()

	const (
		ignored = `resource "contentful_app_definition" "test" {
 organization_id = "org"
 name = "Before"
 locations = [{ location = "app-config" }]
 lifecycle { ignore_changes = [locations] }
}`
		renamed = `resource "contentful_app_definition" "test" {
 organization_id = "org"
 name = "After"
 locations = [{ location = "app-config" }]
 lifecycle { ignore_changes = [locations] }
}`
		corrected = `resource "contentful_app_definition" "test" {
 organization_id = "org"
 name = "After"
 locations = [{ location = "app-config" }]
}`
		originalBody  = `{"sys":{"type":"AppDefinition","id":"app","organization":{"sys":{"type":"Link","linkType":"Organization","id":"org"}}},"name":"Before","src":"https://example.invalid/app","locations":[{"location":"app-config"},{"location":"dialog","fieldTypes":[]}]}`
		correctedBody = `{"sys":{"type":"AppDefinition","id":"app","organization":{"sys":{"type":"Link","linkType":"Organization","id":"org"}}},"name":"After","src":"https://example.invalid/app","locations":[{"location":"app-config"}]}`
	)

	for _, operation := range []string{"destroy", "correct"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()

			var (
				puts, deletes atomic.Int64
				current       atomic.Pointer[string]
			)

			initial := originalBody
			current.Store(&initial)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				assert.Equal(t, "/organizations/org/app_definitions/app", r.URL.Path)
				assert.Empty(t, r.Header.Get("X-Contentful-Version"))

				switch r.Method {
				case http.MethodGet:
					fmt.Fprint(w, *current.Load())
				case http.MethodPut:
					puts.Add(1)

					request, err := io.ReadAll(r.Body)
					if !assert.NoError(t, err) {
						return
					}

					assert.JSONEq(t, `{"name":"After","src":"https://example.invalid/app","locations":[{"location":"app-config"}]}`, string(request))

					body := correctedBody
					current.Store(&body)
					fmt.Fprint(w, body)
				case http.MethodDelete:
					deletes.Add(1)
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.Error(w, "unexpected request", http.StatusInternalServerError)
				}
			})

			steps := []resource.TestStep{
				{Config: ignored, ResourceName: "contentful_app_definition.test", ImportState: true, ImportStateId: "org/app", ImportStatePersist: true},
				{Config: ignored, PlanOnly: true},
				{Config: renamed, PlanOnly: true, ExpectNonEmptyPlan: true},
				{Config: renamed, ExpectError: regexp.MustCompile(`field_types cannot be configured for location "dialog"`)},
			}
			if operation == "destroy" {
				steps = append(steps, resource.TestStep{Config: ignored, Destroy: true})
			} else {
				steps = append(steps, resource.TestStep{Config: corrected, PreConfig: func() {
					require.Zero(t, puts.Load())
				}})
			}

			testAccMockedResource(t, handler, resource.TestCase{Steps: steps})

			if operation == "destroy" {
				assert.Zero(t, puts.Load())
			} else {
				assert.EqualValues(t, 1, puts.Load())
			}

			assert.EqualValues(t, 1, deletes.Load())
		})
	}
}

func TestAccAppDefinitionResourceUnknownLocationLifecycle(t *testing.T) {
	t.Parallel()

	for _, fieldType := range []string{"Symbol", "Array"} {
		t.Run(fieldType, func(t *testing.T) {
			t.Parallel()

			var posts, deletes atomic.Int64

			const body = `{"sys":{"type":"AppDefinition","id":"app","organization":{"sys":{"type":"Link","linkType":"Organization","id":"org"}}},"name":"App","src":"https://example.invalid/app","locations":[{"location":"entry-field","fieldTypes":[{"type":"Symbol"}]}]}`

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				assert.Empty(t, r.Header.Get("X-Contentful-Version"))

				switch r.Method {
				case http.MethodPost:
					posts.Add(1)
					assert.Equal(t, "/organizations/org/app_definitions", r.URL.Path)

					request, err := io.ReadAll(r.Body)
					if !assert.NoError(t, err) {
						return
					}

					assert.JSONEq(t, `{"name":"App","src":"https://example.invalid/app","locations":[{"location":"entry-field","fieldTypes":[{"type":"Symbol"}]}]}`, string(request))

					w.WriteHeader(http.StatusCreated)
					fmt.Fprint(w, body)
				case http.MethodGet:
					assert.Equal(t, "/organizations/org/app_definitions/app", r.URL.Path)
					fmt.Fprint(w, body)
				case http.MethodDelete:
					assert.Equal(t, "/organizations/org/app_definitions/app", r.URL.Path)
					deletes.Add(1)
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.Error(w, "unexpected request", http.StatusInternalServerError)
				}
			})
			config := fmt.Sprintf(`resource "terraform_data" "type" { input = %q }
resource "contentful_app_definition" "test" {
 organization_id = "org"
 name = "App"
 src = "https://example.invalid/app"
 locations = [{ location = "entry-field", field_types = [{ type = terraform_data.type.output }] }]
}`, fieldType)

			apply := resource.TestStep{Config: config}
			if fieldType == "Array" {
				apply.ExpectError = regexp.MustCompile("An Array field type requires an items definition")
			}

			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
				{Config: config, PlanOnly: true, ExpectNonEmptyPlan: true},
				apply,
			}})

			if fieldType == "Array" {
				assert.Zero(t, posts.Load())
				assert.Zero(t, deletes.Load())
			} else {
				assert.EqualValues(t, 1, posts.Load())
				assert.EqualValues(t, 1, deletes.Load())
			}
		})
	}
}

func TestAccAppDefinitionResourceReplacesImportedLocations(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, organization           string
		createBeforeDestroy, tainted bool
	}{
		{name: "organization delete before create", organization: "new-org"},
		{name: "organization create before destroy", organization: "new-org", createBeforeDestroy: true},
		{name: "tainted", organization: "org", tainted: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			const (
				originalPath = "/organizations/org/app_definitions/app"
				originalBody = `{"sys":{"type":"AppDefinition","id":"app","organization":{"sys":{"type":"Link","linkType":"Organization","id":"org"}}},"name":"App","src":"https://example.invalid/app","locations":[{"location":"app-config"},{"location":"dialog","fieldTypes":[]}]}`
			)

			collectionPath := "/organizations/" + test.organization + "/app_definitions"
			replacementPath := collectionPath + "/replacement"
			replacementBody := fmt.Sprintf(`{"sys":{"type":"AppDefinition","id":"replacement","organization":{"sys":{"type":"Link","linkType":"Organization","id":%q}}},"name":"App","src":"https://example.invalid/app","locations":[{"location":"app-config"}]}`, test.organization)

			var (
				mutex     sync.Mutex
				mutations []string
			)

			bodies := map[string]string{originalPath: originalBody}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mutex.Lock()
				defer mutex.Unlock()

				w.Header().Set("Content-Type", "application/json")
				assert.Empty(t, r.Header.Get("X-Contentful-Version"))

				switch r.Method {
				case http.MethodGet:
					assert.Contains(t, []string{originalPath, replacementPath}, r.URL.Path)

					body, exists := bodies[r.URL.Path]
					if !exists {
						w.WriteHeader(http.StatusNotFound)
						fmt.Fprint(w, `{"sys":{"type":"Error","id":"NotFound"},"message":"Not found"}`)

						return
					}

					fmt.Fprint(w, body)
				case http.MethodPost:
					assert.Equal(t, collectionPath, r.URL.Path)
					mutations = append(mutations, r.Method+" "+r.URL.Path)

					request, err := io.ReadAll(r.Body)
					if !assert.NoError(t, err) {
						return
					}

					assert.JSONEq(t, `{"name":"App","src":"https://example.invalid/app","locations":[{"location":"app-config"}]}`, string(request))

					bodies[replacementPath] = replacementBody

					w.WriteHeader(http.StatusCreated)
					fmt.Fprint(w, replacementBody)
				case http.MethodDelete:
					assert.Contains(t, bodies, r.URL.Path)
					mutations = append(mutations, r.Method+" "+r.URL.Path)
					delete(bodies, r.URL.Path)
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.Error(w, "unexpected request", http.StatusInternalServerError)
				}
			})

			const configTemplate = `resource "contentful_app_definition" "test" {
 organization_id = %q
 name = "App"
 src = "https://example.invalid/app"
 locations = [{ location = "app-config" }]
 lifecycle {
  ignore_changes = [locations]
  create_before_destroy = %t
 }
}`

			existing := fmt.Sprintf(configTemplate, "org", test.createBeforeDestroy)
			replaced := fmt.Sprintf(configTemplate, test.organization, test.createBeforeDestroy)

			replacementPlan := resource.TestStep{Config: replaced, PlanOnly: true, ExpectNonEmptyPlan: true}
			if test.tainted {
				replacementPlan.Taint = []string{"contentful_app_definition.test"}
			}

			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
				{Config: existing, ResourceName: "contentful_app_definition.test", ImportState: true, ImportStateId: "org/app", ImportStatePersist: true},
				replacementPlan,
				{Config: replaced, PreConfig: func() {
					mutex.Lock()
					defer mutex.Unlock()

					require.Empty(t, mutations)
				}},
			}})

			expected := []string{"DELETE " + originalPath, "POST " + collectionPath, "DELETE " + replacementPath}
			if test.createBeforeDestroy {
				expected = []string{"POST " + collectionPath, "DELETE " + originalPath, "DELETE " + replacementPath}
			}

			mutex.Lock()
			defer mutex.Unlock()

			assert.Equal(t, expected, mutations)
		})
	}
}
