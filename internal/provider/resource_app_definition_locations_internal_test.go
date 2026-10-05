package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func appDefinitionLocationsTestPlan(t *testing.T, locations string) tfsdk.Plan {
	t.Helper()
	schema := AppDefinitionResourceSchema(t.Context())
	raw, err := tftypes.ValueFromJSONWithOpts([]byte(testJSON(map[string]any{
		"id":                "org/app",
		"organization_id":   "org",
		"app_definition_id": "app",
		"name":              "App",
		"locations":         json.RawMessage(locations),
	})), schema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
	require.NoError(t, err)

	return tfsdk.Plan{Schema: schema, Raw: raw}
}

func appDefinitionLocationsTestValue(t *testing.T, plan tfsdk.Plan) types.List {
	t.Helper()

	var locations types.List
	require.False(t, plan.GetAttribute(t.Context(), path.Root("locations"), &locations).HasError())

	return locations
}

func TestAppDefinitionLocationValidation(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ name, locations, path, detail string }{
		{"dialog fields", testJSON([]any{map[string]any{"location": "dialog", "field_types": []any{}}}), "locations[0].field_types", `field_types cannot be configured for location "dialog". Omit this attribute.`},
		{"dialog navigation", testJSON([]any{
			map[string]any{"location": "dialog", "navigation_item": map[string]any{"name": "Name", "path": "/path"}},
		}), "locations[0].navigation_item", `navigation_item cannot be configured for location "dialog". Omit this attribute.`},
		{"entry fields absent", testJSON([]any{map[string]any{"location": "entry-field"}}), "locations[0].field_types", "An entry-field location requires at least one field_types definition."},
		{"entry fields empty", testJSON([]any{map[string]any{"location": "entry-field", "field_types": []any{}}}), "locations[0].field_types", "An entry-field location requires at least one field_types definition."},
		{"navigation name", testJSON([]any{map[string]any{"location": "page", "navigation_item": map[string]any{"name": "", "path": "/path"}}}), "locations[0].navigation_item.name", "Navigation item name must not be empty."},
		{"navigation path", testJSON([]any{map[string]any{"location": "page", "navigation_item": map[string]any{"name": "Name", "path": ""}}}), "locations[0].navigation_item.path", "Navigation item path must not be empty."},
		{"Array missing items", testJSON([]any{map[string]any{"location": "entry-field", "field_types": []any{map[string]any{"type": "Array"}}}}), "locations[0].field_types[0].items", "An Array field type requires an items definition."},
		{"Array link target", testJSON([]any{
			map[string]any{
				"location":    "entry-field",
				"field_types": []any{map[string]any{"type": "Array", "link_type": "Entry", "items": map[string]any{"type": "Symbol"}}},
			},
		}), "locations[0].field_types[0].link_type", "An Array field type cannot have link_type. Configure link_type inside items for linked items."},
		{"scalar items", testJSON([]any{
			map[string]any{
				"location":    "entry-field",
				"field_types": []any{map[string]any{"type": "Symbol", "items": map[string]any{"type": "Symbol"}}},
			},
		}), "locations[0].field_types[0].items", `Field type "Symbol" cannot have items. Omit this attribute.`},
		{"Link missing target", testJSON([]any{map[string]any{"location": "entry-field", "field_types": []any{map[string]any{"type": "Link"}}}}), "locations[0].field_types[0].link_type", `Field type "Link" requires link_type.`},
		{"ResourceLink missing target", testJSON([]any{map[string]any{"location": "entry-field", "field_types": []any{map[string]any{"type": "ResourceLink"}}}}), "locations[0].field_types[0].link_type", `Field type "ResourceLink" requires link_type.`},
		{"item Link missing target", testJSON([]any{
			map[string]any{
				"location":    "entry-field",
				"field_types": []any{map[string]any{"type": "Array", "items": map[string]any{"type": "Link"}}},
			},
		}), "locations[0].field_types[0].items.link_type", `Field type "Link" requires link_type.`},
		{"item ResourceLink missing target", testJSON([]any{
			map[string]any{
				"location":    "entry-field",
				"field_types": []any{map[string]any{"type": "Array", "items": map[string]any{"type": "ResourceLink"}}},
			},
		}), "locations[0].field_types[0].items.link_type", `Field type "ResourceLink" requires link_type.`},
		{"scalar link target", testJSON([]any{
			map[string]any{
				"location":    "entry-field",
				"field_types": []any{map[string]any{"type": "Symbol", "link_type": "Entry"}},
			},
		}), "locations[0].field_types[0].link_type", `Field type "Symbol" cannot have link_type. Omit this attribute.`},
		{"scalar item link target", testJSON([]any{
			map[string]any{
				"location":    "entry-field",
				"field_types": []any{map[string]any{"type": "Array", "items": map[string]any{"type": "Symbol", "link_type": "Entry"}}},
			},
		}), "locations[0].field_types[0].items.link_type", `Field type "Symbol" cannot have link_type. Omit this attribute.`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			locations := appDefinitionLocationsTestValue(t, appDefinitionLocationsTestPlan(t, test.locations))
			response := validator.ListResponse{}
			appDefinitionLocationsValidator{}.ValidateList(t.Context(), validator.ListRequest{Path: path.Root("locations"), ConfigValue: locations}, &response)
			require.Len(t, response.Diagnostics, 1)
			diagnostic, ok := response.Diagnostics[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			assert.Equal(t, test.path, diagnostic.Path().String())
			assert.Equal(t, test.detail, diagnostic.Detail())
			assert.Equal(t, "Invalid App Definition location configuration", diagnostic.Summary())
		})
	}
}

func TestAppDefinitionLocationsOpenVocabulary(t *testing.T) {
	t.Parallel()

	for _, locations := range []string{
		testJSON([]any{}), testJSON([]any{map[string]any{"location": "page"}}),
		testJSON([]any{
			map[string]any{
				"location":        "FutureLocation",
				"field_types":     []any{},
				"navigation_item": map[string]any{"name": "Name", "path": "/path"},
			},
		}),
		testJSON([]any{
			map[string]any{
				"location": "entry-field",
				"field_types": []any{
					map[string]any{"type": "Array", "items": map[string]any{"type": "Integer"}},
					map[string]any{"type": "Link", "link_type": "FutureTarget"},
					map[string]any{"type": "ResourceLink", "link_type": ""},
					map[string]any{
						"type":      "FutureType",
						"link_type": "FutureTarget",
						"items":     map[string]any{"type": "FutureItem", "link_type": "FutureTarget"},
					},
				},
			},
		}),
	} {
		t.Run(locations, func(t *testing.T) {
			t.Parallel()
			plan := appDefinitionLocationsTestPlan(t, locations)
			assert.Empty(t, validateAppDefinitionPlannedLocations(t.Context(), plan))
		})
	}
}

func TestAppDefinitionLocationsUnknownValues(t *testing.T) {
	t.Parallel()
	baseline := appDefinitionLocationsTestPlan(t, testJSON([]any{
		map[string]any{
			"location":    "entry-field",
			"field_types": []any{map[string]any{"type": "Array", "items": map[string]any{"type": "Link", "link_type": "Entry"}}},
		},
		map[string]any{"location": "page", "navigation_item": map[string]any{"name": "Page", "path": "/page"}},
	}))
	root := path.Root("locations")

	var (
		locations                          types.List
		location, field, items, navigation types.Object
		fields                             types.List
	)

	require.False(t, baseline.GetAttribute(t.Context(), root, &locations).HasError())
	require.False(t, baseline.GetAttribute(t.Context(), root.AtListIndex(0), &location).HasError())
	require.False(t, baseline.GetAttribute(t.Context(), root.AtListIndex(0).AtName("field_types"), &fields).HasError())
	require.False(t, baseline.GetAttribute(t.Context(), root.AtListIndex(0).AtName("field_types").AtListIndex(0), &field).HasError())
	require.False(t, baseline.GetAttribute(t.Context(), root.AtListIndex(0).AtName("field_types").AtListIndex(0).AtName("items"), &items).HasError())
	require.False(t, baseline.GetAttribute(t.Context(), root.AtListIndex(1).AtName("navigation_item"), &navigation).HasError())

	for _, test := range []struct {
		path  path.Path
		value attr.Value
	}{
		{root, locations},
		{root.AtListIndex(0), location},
		{root.AtListIndex(0).AtName("location"), types.StringValue("entry-field")},
		{root.AtListIndex(0).AtName("field_types"), fields},
		{root.AtListIndex(0).AtName("field_types").AtListIndex(0), field},
		{root.AtListIndex(0).AtName("field_types").AtListIndex(0).AtName("type"), types.StringValue("Array")},
		{root.AtListIndex(0).AtName("field_types").AtListIndex(0).AtName("items"), items},
		{root.AtListIndex(0).AtName("field_types").AtListIndex(0).AtName("items").AtName("link_type"), types.StringValue("Entry")},
		{root.AtListIndex(1).AtName("navigation_item"), navigation},
	} {
		t.Run(test.path.String(), func(t *testing.T) {
			t.Parallel()

			plan := baseline
			valueType := test.value.Type(t.Context())
			unknown, err := valueType.ValueFromTerraform(t.Context(), tftypes.NewValue(valueType.TerraformType(t.Context()), tftypes.UnknownValue))
			require.NoError(t, err)
			require.False(t, plan.SetAttribute(t.Context(), test.path, unknown).HasError())
			assert.Empty(t, validateAppDefinitionLocations(appDefinitionLocationsTestValue(t, plan), root))
			diags := validateAppDefinitionPlannedLocations(t.Context(), plan)
			require.Len(t, diags, 1)
			diagnostic, ok := diags[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			assert.Equal(t, test.path, diagnostic.Path())
			assert.Equal(t, "Unknown App Definition location value", diagnostic.Summary())
		})
	}

	siblingPlan := baseline
	require.False(t, siblingPlan.SetAttribute(t.Context(), root.AtListIndex(0).AtName("field_types"), types.ListUnknown(fields.ElementType(t.Context()))).HasError())
	require.False(t, siblingPlan.SetAttribute(t.Context(), root.AtListIndex(1).AtName("navigation_item").AtName("name"), types.StringValue("")).HasError())
	diags := validateAppDefinitionLocations(appDefinitionLocationsTestValue(t, siblingPlan), root)
	require.Len(t, diags, 1)
	assert.Equal(t, "Navigation item name must not be empty.", diags[0].Detail())
}

func appDefinitionLocationsTestResource(t *testing.T, method, expectedBody, responseBody string) (*appDefinitionResource, *atomic.Int64) {
	t.Helper()

	var calls atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		assert.Equal(t, method, r.Method)
		assert.Empty(t, r.Header.Get("X-Contentful-Version"))

		expectedPath := "/organizations/org/app_definitions"
		if method == http.MethodPut {
			expectedPath += "/app"
		}

		assert.Equal(t, expectedPath, r.URL.Path)

		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			return
		}

		assert.JSONEq(t, expectedBody, string(body))
		w.Header().Set("Content-Type", "application/json")

		if method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}

		fmt.Fprint(w, responseBody)
	}))
	t.Cleanup(server.Close)
	client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource("test-token"))
	require.NoError(t, err)

	return &appDefinitionResource{providerData: ContentfulProviderData{client: client}}, &calls
}

func TestAppDefinitionLocationsMutationUsesPlan(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPost, http.MethodPut} {
		for _, test := range []struct {
			name, locations, body, response string
			accepted, unknown               bool
		}{
			{name: "misplaced fields", locations: testJSON([]any{map[string]any{"location": "dialog", "field_types": []any{}}})},
			{name: "missing items", locations: testJSON([]any{map[string]any{"location": "entry-field", "field_types": []any{map[string]any{"type": "Array"}}}})},
			{name: "null root", locations: testJSON(nil)},
			{name: "null location", locations: testJSON([]any{nil})},
			{name: "null field", locations: testJSON([]any{map[string]any{"location": "entry-field", "field_types": []any{nil}}})},
			{name: "null field type", locations: testJSON([]any{map[string]any{"location": "entry-field", "field_types": []any{map[string]any{}}}})},
			{name: "unknown root", locations: testJSON([]any{}), unknown: true},
			{name: "effective valid plan", locations: testJSON([]any{}), body: testJSON(map[string]any{"name": "App", "locations": []any{}}), response: testJSON(map[string]any{"sys": map[string]any{"type": "AppDefinition", "id": "app", "organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}}}, "name": "App", "locations": []any{}}), accepted: true},
			{name: "open vocabulary", locations: testJSON([]any{
				map[string]any{
					"location": "FutureLocation",
					"field_types": []any{
						map[string]any{
							"type":      "FutureType",
							"link_type": "FutureTarget",
							"items":     map[string]any{"type": "FutureItem", "link_type": "FutureItemTarget"},
						},
						map[string]any{"type": "ResourceLink", "link_type": "Contentful:Entry"},
						map[string]any{
							"type":  "Array",
							"items": map[string]any{"type": "ResourceLink", "link_type": "Example:Record"},
						},
					},
					"navigation_item": map[string]any{"name": "Page", "path": "/page"},
				},
			}), body: testJSON(map[string]any{
				"name": "App",
				"locations": []any{
					map[string]any{
						"location": "FutureLocation",
						"fieldTypes": []any{
							map[string]any{
								"type":     "FutureType",
								"linkType": "FutureTarget",
								"items":    map[string]any{"type": "FutureItem", "linkType": "FutureItemTarget"},
							},
							map[string]any{"type": "ResourceLink", "linkType": "Contentful:Entry"},
							map[string]any{"type": "Array", "items": map[string]any{"type": "ResourceLink", "linkType": "Example:Record"}},
						},
						"navigationItem": map[string]any{"name": "Page", "path": "/page"},
					},
				},
			}), response: testJSON(map[string]any{
				"sys":  map[string]any{"type": "AppDefinition", "id": "app", "organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}}},
				"name": "App",
				"locations": []any{
					map[string]any{
						"location": "FutureLocation",
						"fieldTypes": []any{
							map[string]any{
								"type":     "FutureType",
								"linkType": "FutureTarget",
								"items":    map[string]any{"type": "FutureItem", "linkType": "FutureItemTarget"},
							},
							map[string]any{"type": "ResourceLink", "linkType": "Contentful:Entry"},
							map[string]any{"type": "Array", "items": map[string]any{"type": "ResourceLink", "linkType": "Example:Record"}},
						},
						"navigationItem": map[string]any{"name": "Page", "path": "/page"},
					},
				},
			}), accepted: true},
		} {
			t.Run(method+"/"+test.name, func(t *testing.T) {
				t.Parallel()
				implementation, calls := appDefinitionLocationsTestResource(t, method, test.body, test.response)

				plan := appDefinitionLocationsTestPlan(t, test.locations)
				if test.unknown {
					locations := appDefinitionLocationsTestValue(t, plan)
					require.False(t, plan.SetAttribute(t.Context(), path.Root("locations"), types.ListUnknown(locations.ElementType(t.Context()))).HasError())
				}

				config := tfsdk.Config(appDefinitionLocationsTestPlan(t, testJSON([]any{})))
				if test.accepted {
					config = tfsdk.Config(appDefinitionLocationsTestPlan(t, testJSON([]any{map[string]any{"location": "dialog", "field_types": []any{}}})))
				}

				state := tfsdk.State(appDefinitionLocationsTestPlan(t, testJSON([]any{})))
				identitySchema := resourceIdentitySchema(appDefinitionIdentityAttributeNames())
				identity := &tfsdk.ResourceIdentity{Schema: identitySchema, Raw: tftypes.NewValue(identitySchema.Type().TerraformType(t.Context()), nil)}

				var diags diag.Diagnostics

				if method == http.MethodPost {
					response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}, Identity: identity}
					implementation.Create(t.Context(), resource.CreateRequest{Plan: plan, Config: config}, &response)
					diags = response.Diagnostics
				} else {
					response := resource.UpdateResponse{State: state, Identity: identity}
					implementation.Update(t.Context(), resource.UpdateRequest{Plan: plan, Config: config, State: state}, &response)

					diags = response.Diagnostics
					if !test.accepted {
						assert.True(t, state.Raw.Equal(response.State.Raw))
					}
				}

				if test.accepted {
					assert.Empty(t, diags)
					assert.EqualValues(t, 1, calls.Load())
				} else {
					assert.True(t, diags.HasError())
					assert.Zero(t, calls.Load())
				}
			})
		}
	}
}

func TestAppDefinitionLocationsReadPreservesIrregularValues(t *testing.T) {
	t.Parallel()

	responseBody := testJSON(map[string]any{
		"sys": map[string]any{
			"type":         "AppDefinition",
			"id":           "app",
			"organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}},
		},
		"name": "App",
		"locations": []any{
			map[string]any{
				"location":       "dialog",
				"fieldTypes":     []any{map[string]any{"type": "Array"}},
				"navigationItem": map[string]any{"name": "", "path": "/page"},
			},
			map[string]any{"location": "entry-field", "fieldTypes": []any{}},
		},
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/organizations/org/app_definitions/app", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, responseBody)
	}))
	t.Cleanup(server.Close)
	client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource("test-token"))
	require.NoError(t, err)

	implementation := appDefinitionResource{providerData: ContentfulProviderData{client: client}}
	identitySchema := resourceIdentitySchema(appDefinitionIdentityAttributeNames())
	identity := &tfsdk.ResourceIdentity{Schema: identitySchema, Raw: tftypes.NewValue(identitySchema.Type().TerraformType(t.Context()), nil)}
	imported := resource.ImportStateResponse{State: tfsdk.State(appDefinitionLocationsTestPlan(t, testJSON([]any{}))), Identity: identity}
	implementation.ImportState(t.Context(), resource.ImportStateRequest{ID: "org/app"}, &imported)
	require.Empty(t, imported.Diagnostics)
	read := resource.ReadResponse{State: imported.State, Identity: identity}
	implementation.Read(t.Context(), resource.ReadRequest{State: imported.State}, &read)
	require.Empty(t, read.Diagnostics)
	expected := appDefinitionLocationsTestValue(t, appDefinitionLocationsTestPlan(t, testJSON([]any{
		map[string]any{
			"location":        "dialog",
			"field_types":     []any{map[string]any{"type": "Array"}},
			"navigation_item": map[string]any{"name": "", "path": "/page"},
		},
		map[string]any{"location": "entry-field", "field_types": []any{}},
	})))

	var actual types.List
	require.False(t, read.State.GetAttribute(t.Context(), path.Root("locations"), &actual).HasError())
	assert.True(t, expected.Equal(actual), "expected %s, got %s", expected, actual)
}

func TestAppDefinitionLocationsPartialKnowledge(t *testing.T) {
	t.Parallel()

	fieldPath := path.Root("locations").AtListIndex(0).AtName("field_types").AtListIndex(0)
	for _, test := range []struct {
		name, locations, detail string
		unknown                 path.Path
	}{
		{name: "unknown type with absent items", locations: testJSON([]any{map[string]any{"location": "entry-field", "field_types": []any{map[string]any{"type": "Symbol"}}}}), unknown: fieldPath.AtName("type")},
		{name: "known object presence with unknown child", locations: testJSON([]any{
			map[string]any{
				"location":    "entry-field",
				"field_types": []any{map[string]any{"type": "Symbol", "items": map[string]any{"type": "Symbol"}}},
			},
		}), unknown: fieldPath.AtName("items").AtName("type"), detail: `Field type "Symbol" cannot have items. Omit this attribute.`},
		{name: "unfamiliar parents with known child", locations: testJSON([]any{
			map[string]any{
				"location":    "FutureLocation",
				"field_types": []any{map[string]any{"type": "FutureType", "items": map[string]any{"type": "Link"}}},
			},
		}), detail: `Field type "Link" requires link_type.`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			plan := appDefinitionLocationsTestPlan(t, test.locations)
			if len(test.unknown.Steps()) != 0 {
				require.False(t, plan.SetAttribute(t.Context(), test.unknown, types.StringUnknown()).HasError())
			}

			diags := validateAppDefinitionLocations(appDefinitionLocationsTestValue(t, plan), path.Root("locations"))
			if test.detail == "" {
				assert.Empty(t, diags)
			} else {
				require.Len(t, diags, 1)
				assert.Equal(t, test.detail, diags[0].Detail())
			}
		})
	}
}
