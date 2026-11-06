package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests access Read implementations directly to assert that errors never
// publish partial state. Raw HTTP fixtures are independent of CMA constructors.
type discoveryTestFamily struct {
	name                  string
	singular              func() datasource.DataSource
	plural                func() datasource.DataSource
	scopes                map[string]any
	id                    string
	collection            string
	endpoint              string
	organization          string
	expected              string
	preserveResponseOrder bool
}

func discoveryTestFamilies() []discoveryTestFamily {
	return []discoveryTestFamily{
		{name: "space", singular: NewSpaceDataSource, plural: NewSpacesDataSource, scopes: map[string]any{}, id: "space_id", collection: "spaces", endpoint: "/spaces", organization: "org", expected: testJSON(map[string]any{"space_id": "item-b", "name": "Second", "organization_id": "org"})},
		{name: "environment", singular: NewEnvironmentDataSource, plural: NewEnvironmentsDataSource, scopes: map[string]any{"space_id": "space"}, id: "environment_id", collection: "environments", endpoint: "/spaces/space/environments", organization: "", expected: testJSON(map[string]any{
			"environment_id":         "item-b",
			"name":                   "Second",
			"status":                 "queued",
			"aliased_environment_id": "target",
		})},
		{name: "environment_alias", singular: NewEnvironmentAliasDataSource, plural: NewEnvironmentAliasesDataSource, scopes: map[string]any{"space_id": "space"}, id: "environment_alias_id", collection: "environment_aliases", endpoint: "/spaces/space/environment_aliases", organization: "", expected: testJSON(map[string]any{"environment_alias_id": "item-b", "target_environment_id": "target"})},
		{name: "locale", singular: NewLocaleDataSource, plural: NewLocalesDataSource, scopes: map[string]any{"space_id": "space", "environment_id": "master"}, id: "locale_id", collection: "locales", endpoint: "/spaces/space/environments/master/locales", organization: "", expected: testJSON(map[string]any{
			"locale_id":              "item-b",
			"name":                   "Second",
			"code":                   "en-GB",
			"default":                true,
			"fallback_code":          nil,
			"optional":               false,
			"content_management_api": true,
			"content_delivery_api":   false,
		})},
		{name: "role", singular: NewRoleDataSource, plural: NewRolesDataSource, scopes: map[string]any{"space_id": "space"}, id: "role_id", collection: "roles", endpoint: "/spaces/space/roles", organization: "", preserveResponseOrder: true, expected: testJSON(map[string]any{
			"role_id":     "item-b",
			"name":        "Second",
			"description": nil,
			"permissions": map[string]any{"ContentDelivery": []any{"all"}, "ContentModel": []any{"read", "read"}, "Tags": []any{}},
			"policies": []any{
				map[string]any{"actions": []any{"all"}, "constraint": "{\"a\":[1,2],\"z\":1}", "effect": "allow"},
				map[string]any{"actions": []any{"read", "read"}, "constraint": nil, "effect": "deny"},
			},
		})},
	}
}

func discoveryFixture(t *testing.T, family string) string {
	t.Helper()

	body, err := os.ReadFile("testdata/discovery/" + family + ".json")
	require.NoError(t, err)

	return strings.TrimSpace(string(body))
}

func discoveryReadTest(ctx context.Context, t *testing.T, factory func() datasource.DataSource, inputs map[string]any, transport http.RoundTripper) datasource.ReadResponse {
	t.Helper()

	implementation := factory()

	var schemaResponse datasource.SchemaResponse
	implementation.Schema(ctx, datasource.SchemaRequest{}, &schemaResponse)
	require.False(t, schemaResponse.Diagnostics.HasError(), schemaResponse.Diagnostics)
	schema := schemaResponse.Schema

	values := make(map[string]tftypes.Value, len(schema.Attributes))
	for name, attribute := range schema.Attributes {
		values[name] = tftypes.NewValue(attribute.GetType().TerraformType(ctx), inputs[name])
	}

	client, err := cm.NewClient("https://example.invalid/base", cm.NewAccessTokenSecuritySource("synthetic-token"), cm.WithClient(&http.Client{Transport: transport}))
	require.NoError(t, err)

	configurable, ok := implementation.(datasource.DataSourceWithConfigure)
	require.True(t, ok)

	var configureResponse datasource.ConfigureResponse
	configurable.Configure(ctx, datasource.ConfigureRequest{ProviderData: ContentfulProviderData{client: client}}, &configureResponse)
	require.False(t, configureResponse.Diagnostics.HasError(), configureResponse.Diagnostics)

	response := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	implementation.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: schema, Raw: tftypes.NewValue(schema.Type().TerraformType(ctx), values)}}, &response)

	return response
}

func discoveryHTTPResponse(request *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}
}

func discoveryPage(skip, limit, total int, items ...string) string {
	documents := make([]json.RawMessage, len(items))
	for i, item := range items {
		documents[i] = json.RawMessage(item)
	}

	return testJSON(map[string]any{
		"sys":   map[string]any{"type": "Array"},
		"skip":  skip,
		"limit": limit,
		"total": total,
		"items": documents,
	})
}

func discoveryStateAttribute(t *testing.T, response datasource.ReadResponse, name string, expected string) {
	t.Helper()

	var attributes map[string]tftypes.Value
	require.NoError(t, response.State.Raw.As(&attributes))
	actual, exists := attributes[name]
	require.True(t, exists)

	var value any
	require.NoError(t, json.Unmarshal([]byte(expected), &value))
	assert.Equal(t, value, discoveryPlainValue(t, actual), name)
}

func discoveryPlainValue(t *testing.T, value tftypes.Value) any {
	t.Helper()

	if value.IsNull() {
		return nil
	}

	switch value.Type().(type) {
	case tftypes.Object:
		var fields map[string]tftypes.Value
		require.NoError(t, value.As(&fields))

		result := make(map[string]any, len(fields))
		for name, field := range fields {
			result[name] = discoveryPlainValue(t, field)
		}

		return result
	case tftypes.Map:
		var fields map[string]tftypes.Value
		require.NoError(t, value.As(&fields))

		result := make(map[string]any, len(fields))
		for name, field := range fields {
			result[name] = discoveryPlainValue(t, field)
		}

		return result
	case tftypes.List:
		var items []tftypes.Value
		require.NoError(t, value.As(&items))

		result := make([]any, 0, len(items))
		for _, item := range items {
			result = append(result, discoveryPlainValue(t, item))
		}

		return result
	default:
		if value.Type().Is(tftypes.String) {
			var result string
			require.NoError(t, value.As(&result))

			return result
		}

		var result bool
		require.NoError(t, value.As(&result))

		return result
	}
}
