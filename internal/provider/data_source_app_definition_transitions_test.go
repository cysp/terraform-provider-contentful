package provider_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccAppDefinitionDataSourcesRawTransitions(t *testing.T) {
	t.Parallel()

	for _, marketplace := range []bool{false, true} {
		t.Run(fmt.Sprintf("marketplace=%v", marketplace), func(t *testing.T) {
			t.Parallel()

			var payload atomic.Pointer[string]

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/organizations/org/app_definitions/app"
				if marketplace {
					expectedPath = "/app_definitions"
				}

				if r.Method != http.MethodGet || r.URL.Path != expectedPath {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL)
					http.Error(w, "unexpected request", http.StatusInternalServerError)

					return
				}

				if marketplace && r.URL.Query().Get("sys.id[in]") != "app" {
					t.Errorf("unexpected marketplace query: %s", r.URL.RawQuery)
				}

				body := *payload.Load()
				if marketplace {
					body = testJSON(map[string]any{"sys": map[string]any{"type": "Array"}, "total": 1, "items": []any{json.RawMessage(body)}})
				}

				w.Header().Set("Content-Type", "application/vnd.contentful.management.v1+json")
				fmt.Fprint(w, body)
			})
			kind := "app_definition"
			org := `organization_id = "org"`

			if marketplace {
				kind = "marketplace_app_definition"
				org = ""
			}

			address := "data.contentful_" + kind + ".test"
			config := fmt.Sprintf(`
data "contentful_%s" "test" {
 %s
 app_definition_id = "app"
}
output "item_types" {
 value = jsonencode( %s.locations == null ? [] : flatten([
  for location in %s.locations : location.field_types == null ? [] : [
   for field in location.field_types : field.items == null ? "absent" : field.items.type
  ]
 ]))
}
`, kind, org, address, address)
			fieldTypes := tfjsonpath.New("locations").AtSliceIndex(0).AtMapKey("field_types")
			states := []struct {
				attributes map[string]any
				checks     []statecheck.StateCheck
				output     string
			}{
				{map[string]any{
					"locations": []any{
						map[string]any{
							"location":   "entry-field",
							"fieldTypes": []any{map[string]any{"type": "Array", "items": map[string]any{"type": "Symbol"}}},
						},
					},
				}, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(0).AtMapKey("items"), knownvalue.ObjectExact(map[string]knownvalue.Check{"type": knownvalue.StringExact("Symbol"), "link_type": knownvalue.Null()}))}, testJSON([]any{"Symbol"})},
				{map[string]any{
					"locations": []any{map[string]any{"location": "entry-field", "fieldTypes": []any{map[string]any{"type": "Symbol"}}}},
				}, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(0).AtMapKey("items"), knownvalue.Null())}, testJSON([]any{"absent"})},
				{map[string]any{
					"locations": []any{map[string]any{"location": "entry-field", "fieldTypes": []any{map[string]any{"type": "Array"}}}},
				}, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(0).AtMapKey("items"), knownvalue.Null())}, testJSON([]any{"absent"})},
				{map[string]any{
					"locations": []any{
						map[string]any{
							"location": "entry-field",
							"fieldTypes": []any{
								map[string]any{
									"type":     "FutureType",
									"linkType": "Sibling",
									"items":    map[string]any{"type": "", "linkType": ""},
								},
							},
						},
					},
				}, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(0).AtMapKey("type"), knownvalue.StringExact("FutureType")), statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(0).AtMapKey("link_type"), knownvalue.StringExact("Sibling")), statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(0).AtMapKey("items"), knownvalue.ObjectExact(map[string]knownvalue.Check{"type": knownvalue.StringExact(""), "link_type": knownvalue.StringExact("")}))}, testJSON([]any{""})},
				{map[string]any{"locations": []any{map[string]any{"location": "entry-field", "fieldTypes": []any{}}}}, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, fieldTypes, knownvalue.ListExact([]knownvalue.Check{}))}, testJSON([]any{})},
				{map[string]any{"locations": []any{map[string]any{"location": "entry-field"}}}, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, fieldTypes, knownvalue.Null())}, testJSON([]any{})},
				{map[string]any{"locations": []any{}}, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, tfjsonpath.New("locations"), knownvalue.ListExact([]knownvalue.Check{}))}, testJSON([]any{})},
				{map[string]any{}, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, tfjsonpath.New("locations"), knownvalue.Null())}, testJSON([]any{})},
			}

			steps := make([]resource.TestStep, 0, len(states))
			for _, state := range states {
				steps = append(steps, resource.TestStep{PreConfig: func() {
					document := map[string]any{"sys": map[string]any{"type": "AppDefinition", "id": "app", "organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}}}, "name": "Raw fixture"}
					maps.Copy(document, state.attributes)

					body := testJSON(document)
					payload.Store(&body)
				}, Config: config, ConfigStateChecks: state.checks, Check: resource.TestCheckOutput("item_types", state.output)})
			}

			testAccMockedResource(t, handler, resource.TestCase{Steps: steps})
		})
	}
}
