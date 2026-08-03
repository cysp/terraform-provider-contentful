//nolint:testpackage // List request conversion is intentionally tested through its package-private boundary.
package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeListResourceConfigRequestParamsRejectsUnresolvedValues(t *testing.T) {
	t.Parallel()

	params, diags := (contentTypeListResourceConfig{
		SpaceID:       types.StringUnknown(),
		EnvironmentID: types.StringNull(),
	}).requestParams()

	assert.Empty(t, params)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"space_id", "environment_id"}, requestDiagnosticPaths(t, diags))
}

func TestContentTypeListResourceListReturnsConfigurationDiagnosticsOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	configSchema := ContentTypeListResourceConfigSchema(ctx)
	config := tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"space_id":       tftypes.String,
				"environment_id": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"space_id":       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"environment_id": tftypes.NewValue(tftypes.String, "environment"),
		}),
		Schema: configSchema,
	}

	var stream list.ListResultsStream
	(&contentTypeListResource{}).List(ctx, list.ListRequest{Config: config}, &stream)

	result := requireSingleDiagnosticOnlyListResult(t, stream)
	assert.Equal(t, []string{"space_id"}, requestDiagnosticPaths(t, result.Diagnostics))
}

func TestEntryListResourceConfigRequestRejectsUnresolvedValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configure           func(*entryListResourceConfig)
		expectedDiagnostics []string
	}{
		"required identifiers": {
			configure: func(config *entryListResourceConfig) {
				config.SpaceID = types.StringUnknown()
				config.EnvironmentID = types.StringNull()
			},
			expectedDiagnostics: []string{"space_id", "environment_id"},
		},
		"content type": {
			configure: func(config *entryListResourceConfig) {
				config.ContentType = types.StringUnknown()
			},
			expectedDiagnostics: []string{"content_type"},
		},
		"order container": {
			configure: func(config *entryListResourceConfig) {
				config.Order = NewTypedListUnknown[types.String]()
			},
			expectedDiagnostics: []string{"order"},
		},
		"order elements": {
			configure: func(config *entryListResourceConfig) {
				config.Order = NewTypedList([]types.String{
					types.StringValue("sys.createdAt"),
					types.StringNull(),
					types.StringUnknown(),
				})
			},
			expectedDiagnostics: []string{"order[1]", "order[2]"},
		},
		"query container": {
			configure: func(config *entryListResourceConfig) {
				config.Query = NewTypedMapUnknown[types.String]()
			},
			expectedDiagnostics: []string{"query"},
		},
		"query values": {
			configure: func(config *entryListResourceConfig) {
				config.Query = NewTypedMap(map[string]types.String{
					"null":    types.StringNull(),
					"unknown": types.StringUnknown(),
				})
			},
			expectedDiagnostics: []string{`query["null"]`, `query["unknown"]`},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := validEntryListResourceConfig()
			test.configure(&config)

			request, diags := config.request()

			assert.Empty(t, request)
			require.True(t, diags.HasError())
			assert.Equal(t, test.expectedDiagnostics, requestDiagnosticPaths(t, diags))
		})
	}
}

func TestEntryListResourceConfigRequestPreservesExistingOptionalSemantics(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configure           func(*entryListResourceConfig)
		expectedContentType string
		expectedOrder       []string
		expectedQuery       map[string]string
	}{
		"null values are omitted": {
			configure: func(_ *entryListResourceConfig) {},
		},
		"known empty containers remain empty": {
			configure: func(config *entryListResourceConfig) {
				config.Order = NewTypedList([]types.String{})
				config.Query = NewTypedMap(map[string]types.String{})
			},
			expectedOrder: []string{},
			expectedQuery: map[string]string{},
		},
		"empty content type and order elements remain omitted": {
			configure: func(config *entryListResourceConfig) {
				config.ContentType = types.StringValue("")
				config.Order = NewTypedList([]types.String{
					types.StringValue(""),
					types.StringValue("sys.createdAt"),
					types.StringValue(""),
				})
			},
			expectedOrder: []string{"sys.createdAt"},
		},
		"known values are preserved": {
			configure: func(config *entryListResourceConfig) {
				config.ContentType = types.StringValue("author")
				config.Query = NewTypedMap(map[string]types.String{
					"fields.name[match]": types.StringValue(""),
				})
			},
			expectedContentType: "author",
			expectedQuery: map[string]string{
				"fields.name[match]": "",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config := validEntryListResourceConfig()
			test.configure(&config)

			request, diags := config.request()

			require.False(t, diags.HasError(), diags.Errors())
			assert.Equal(t, test.expectedOrder, request.params.Order)
			assert.Equal(t, test.expectedQuery, request.query)

			contentType, contentTypeSet := request.params.ContentType.Get()
			if test.expectedContentType == "" {
				assert.False(t, contentTypeSet)
			} else {
				require.True(t, contentTypeSet)
				assert.Equal(t, test.expectedContentType, contentType)
			}
		})
	}
}

func TestEntryListResourceListReturnsConfigurationDiagnosticsOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := entryListResourceTerraformConfig(ctx, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))

	var stream list.ListResultsStream
	(&entryListResource{}).List(ctx, list.ListRequest{Config: config}, &stream)

	result := requireSingleDiagnosticOnlyListResult(t, stream)
	assert.Equal(t, []string{"content_type"}, requestDiagnosticPaths(t, result.Diagnostics))
}

func validEntryListResourceConfig() entryListResourceConfig {
	return entryListResourceConfig{
		SpaceID:       types.StringValue("space"),
		EnvironmentID: types.StringValue("environment"),
		ContentType:   types.StringNull(),
		Order:         NewTypedListNull[types.String](),
		Query:         NewTypedMapNull[types.String](),
	}
}

func entryListResourceTerraformConfig(ctx context.Context, contentType tftypes.Value) tfsdk.Config {
	listType := tftypes.List{ElementType: tftypes.String}
	mapType := tftypes.Map{ElementType: tftypes.String}

	return tfsdk.Config{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"space_id":       tftypes.String,
				"environment_id": tftypes.String,
				"content_type":   tftypes.String,
				"order":          listType,
				"query":          mapType,
			},
		}, map[string]tftypes.Value{
			"space_id":       tftypes.NewValue(tftypes.String, "space"),
			"environment_id": tftypes.NewValue(tftypes.String, "environment"),
			"content_type":   contentType,
			"order":          tftypes.NewValue(listType, nil),
			"query":          tftypes.NewValue(mapType, nil),
		}),
		Schema: EntryListResourceConfigSchema(ctx),
	}
}

func requireSingleDiagnosticOnlyListResult(t *testing.T, stream list.ListResultsStream) list.ListResult {
	t.Helper()

	var results []list.ListResult

	stream.Results(func(result list.ListResult) bool {
		results = append(results, result)

		return true
	})

	require.Len(t, results, 1)
	result := results[0]
	require.True(t, result.Diagnostics.HasError())
	assert.Nil(t, result.Identity)
	assert.Nil(t, result.Resource)
	assert.Empty(t, result.DisplayName)

	return result
}
