package provider

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	raw, err := tftypes.ValueFromJSONWithOpts([]byte(`{"id":"org/app","organization_id":"org","app_definition_id":"app","name":"App","locations":`+locations+`}`), schema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
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
		{"dialog fields", `[{"location":"dialog","field_types":[]}]`, "locations[0].field_types", `field_types cannot be configured for location "dialog". Omit this attribute.`},
		{"dialog navigation", `[{"location":"dialog","navigation_item":{"name":"Name","path":"/path"}}]`, "locations[0].navigation_item", `navigation_item cannot be configured for location "dialog". Omit this attribute.`},
		{"entry fields absent", `[{"location":"entry-field"}]`, "locations[0].field_types", "An entry-field location requires at least one field_types definition."},
		{"entry fields empty", `[{"location":"entry-field","field_types":[]}]`, "locations[0].field_types", "An entry-field location requires at least one field_types definition."},
		{"navigation name", `[{"location":"page","navigation_item":{"name":"","path":"/path"}}]`, "locations[0].navigation_item.name", "Navigation item name must not be empty."},
		{"navigation path", `[{"location":"page","navigation_item":{"name":"Name","path":""}}]`, "locations[0].navigation_item.path", "Navigation item path must not be empty."},
		{"Array missing items", `[{"location":"entry-field","field_types":[{"type":"Array"}]}]`, "locations[0].field_types[0].items", "An Array field type requires an items definition."},
		{"Array link target", `[{"location":"entry-field","field_types":[{"type":"Array","link_type":"Entry","items":{"type":"Symbol"}}]}]`, "locations[0].field_types[0].link_type", "An Array field type cannot have link_type. Configure link_type inside items for linked items."},
		{"scalar items", `[{"location":"entry-field","field_types":[{"type":"Symbol","items":{"type":"Symbol"}}]}]`, "locations[0].field_types[0].items", `Field type "Symbol" cannot have items. Omit this attribute.`},
		{"Link missing target", `[{"location":"entry-field","field_types":[{"type":"Link"}]}]`, "locations[0].field_types[0].link_type", `Field type "Link" requires link_type.`},
		{"ResourceLink missing target", `[{"location":"entry-field","field_types":[{"type":"ResourceLink"}]}]`, "locations[0].field_types[0].link_type", `Field type "ResourceLink" requires link_type.`},
		{"item Link missing target", `[{"location":"entry-field","field_types":[{"type":"Array","items":{"type":"Link"}}]}]`, "locations[0].field_types[0].items.link_type", `Field type "Link" requires link_type.`},
		{"item ResourceLink missing target", `[{"location":"entry-field","field_types":[{"type":"Array","items":{"type":"ResourceLink"}}]}]`, "locations[0].field_types[0].items.link_type", `Field type "ResourceLink" requires link_type.`},
		{"scalar link target", `[{"location":"entry-field","field_types":[{"type":"Symbol","link_type":"Entry"}]}]`, "locations[0].field_types[0].link_type", `Field type "Symbol" cannot have link_type. Omit this attribute.`},
		{"scalar item link target", `[{"location":"entry-field","field_types":[{"type":"Array","items":{"type":"Symbol","link_type":"Entry"}}]}]`, "locations[0].field_types[0].items.link_type", `Field type "Symbol" cannot have link_type. Omit this attribute.`},
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
		`[]`, `[{"location":"page"}]`,
		`[{"location":"FutureLocation","field_types":[],"navigation_item":{"name":"Name","path":"/path"}}]`,
		`[{"location":"entry-field","field_types":[{"type":"Array","items":{"type":"Integer"}},{"type":"Link","link_type":"FutureTarget"},{"type":"ResourceLink","link_type":""},{"type":"FutureType","link_type":"FutureTarget","items":{"type":"FutureItem","link_type":"FutureTarget"}}]}]`,
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
	baseline := appDefinitionLocationsTestPlan(t, `[{"location":"entry-field","field_types":[{"type":"Array","items":{"type":"Link","link_type":"Entry"}}]},{"location":"page","navigation_item":{"name":"Page","path":"/page"}}]`)
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

func appDefinitionLocationsTestResource(t *testing.T, method, expectedBody string) (*appDefinitionResource, *atomic.Int64) {
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

		fmt.Fprint(w, `{"sys":{"type":"AppDefinition","id":"app","organization":{"sys":{"type":"Link","linkType":"Organization","id":"org"}}},`+strings.TrimPrefix(expectedBody, "{"))
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
			name, locations, body string
			accepted, unknown     bool
		}{
			{name: "misplaced fields", locations: `[{"location":"dialog","field_types":[]}]`},
			{name: "missing items", locations: `[{"location":"entry-field","field_types":[{"type":"Array"}]}]`},
			{name: "null root", locations: `null`},
			{name: "null location", locations: `[null]`},
			{name: "null field", locations: `[{"location":"entry-field","field_types":[null]}]`},
			{name: "null field type", locations: `[{"location":"entry-field","field_types":[{}]}]`},
			{name: "unknown root", locations: `[]`, unknown: true},
			{name: "effective valid plan", locations: `[]`, body: `{"name":"App","locations":[]}`, accepted: true},
			{name: "open vocabulary", locations: `[{"location":"FutureLocation","field_types":[{"type":"FutureType","link_type":"FutureTarget","items":{"type":"FutureItem","link_type":"FutureItemTarget"}},{"type":"ResourceLink","link_type":"Contentful:Entry"},{"type":"Array","items":{"type":"ResourceLink","link_type":"Example:Record"}}],"navigation_item":{"name":"Page","path":"/page"}}]`, body: `{"name":"App","locations":[{"location":"FutureLocation","fieldTypes":[{"type":"FutureType","linkType":"FutureTarget","items":{"type":"FutureItem","linkType":"FutureItemTarget"}},{"type":"ResourceLink","linkType":"Contentful:Entry"},{"type":"Array","items":{"type":"ResourceLink","linkType":"Example:Record"}}],"navigationItem":{"name":"Page","path":"/page"}}]}`, accepted: true},
		} {
			t.Run(method+"/"+test.name, func(t *testing.T) {
				t.Parallel()
				implementation, calls := appDefinitionLocationsTestResource(t, method, test.body)

				plan := appDefinitionLocationsTestPlan(t, test.locations)
				if test.unknown {
					locations := appDefinitionLocationsTestValue(t, plan)
					require.False(t, plan.SetAttribute(t.Context(), path.Root("locations"), types.ListUnknown(locations.ElementType(t.Context()))).HasError())
				}

				config := tfsdk.Config(appDefinitionLocationsTestPlan(t, `[]`))
				if test.accepted {
					config = tfsdk.Config(appDefinitionLocationsTestPlan(t, `[{"location":"dialog","field_types":[]}]`))
				}

				state := tfsdk.State(appDefinitionLocationsTestPlan(t, `[]`))
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

	const responseBody = `{"sys":{"type":"AppDefinition","id":"app","organization":{"sys":{"type":"Link","linkType":"Organization","id":"org"}}},"name":"App","locations":[{"location":"dialog","fieldTypes":[{"type":"Array"}],"navigationItem":{"name":"","path":"/page"}},{"location":"entry-field","fieldTypes":[]}]}`

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
	imported := resource.ImportStateResponse{State: tfsdk.State(appDefinitionLocationsTestPlan(t, `[]`)), Identity: identity}
	implementation.ImportState(t.Context(), resource.ImportStateRequest{ID: "org/app"}, &imported)
	require.Empty(t, imported.Diagnostics)
	read := resource.ReadResponse{State: imported.State, Identity: identity}
	implementation.Read(t.Context(), resource.ReadRequest{State: imported.State}, &read)
	require.Empty(t, read.Diagnostics)
	expected := appDefinitionLocationsTestValue(t, appDefinitionLocationsTestPlan(t, `[{"location":"dialog","field_types":[{"type":"Array"}],"navigation_item":{"name":"","path":"/page"}},{"location":"entry-field","field_types":[]}]`))

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
		{name: "unknown type with absent items", locations: `[{"location":"entry-field","field_types":[{"type":"Symbol"}]}]`, unknown: fieldPath.AtName("type")},
		{name: "known object presence with unknown child", locations: `[{"location":"entry-field","field_types":[{"type":"Symbol","items":{"type":"Symbol"}}]}]`, unknown: fieldPath.AtName("items").AtName("type"), detail: `Field type "Symbol" cannot have items. Omit this attribute.`},
		{name: "unfamiliar parents with known child", locations: `[{"location":"FutureLocation","field_types":[{"type":"FutureType","items":{"type":"Link"}}]}]`, detail: `Field type "Link" requires link_type.`},
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
