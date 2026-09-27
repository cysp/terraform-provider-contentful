package provider_test

import (
	"fmt"
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

			const prefix = `{"sys":{"type":"AppDefinition","id":"app","organization":{"sys":{"type":"Link","linkType":"Organization","id":"org"}}},"name":"Raw fixture"`

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
					body = `{"sys":{"type":"Array"},"total":1,"items":[` + body + `]}`
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
				suffix string
				checks []statecheck.StateCheck
				output string
			}{
				{`,"locations":[{"location":"entry-field","fieldTypes":[{"type":"Array","items":{"type":"Symbol"}}]}]}`, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(0).AtMapKey("items"), knownvalue.ObjectExact(map[string]knownvalue.Check{"type": knownvalue.StringExact("Symbol"), "link_type": knownvalue.Null()}))}, `["Symbol"]`},
				{`,"locations":[{"location":"entry-field","fieldTypes":[{"type":"Symbol"}]}]}`, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(0).AtMapKey("items"), knownvalue.Null())}, `["absent"]`},
				{`,"locations":[{"location":"entry-field","fieldTypes":[]}]}`, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, fieldTypes, knownvalue.ListExact([]knownvalue.Check{}))}, `[]`},
				{`,"locations":[{"location":"entry-field"}]}`, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, fieldTypes, knownvalue.Null())}, `[]`},
				{`,"locations":[]}`, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, tfjsonpath.New("locations"), knownvalue.ListExact([]knownvalue.Check{}))}, `[]`},
				{`}`, []statecheck.StateCheck{statecheck.ExpectKnownValue(address, tfjsonpath.New("locations"), knownvalue.Null())}, `[]`},
			}

			steps := make([]resource.TestStep, 0, len(states))
			for _, state := range states {
				steps = append(steps, resource.TestStep{PreConfig: func() { body := prefix + state.suffix; payload.Store(&body) }, Config: config, ConfigStateChecks: state.checks, Check: resource.TestCheckOutput("item_types", state.output)})
			}

			testAccMockedResource(t, handler, resource.TestCase{Steps: steps})
		})
	}
}
