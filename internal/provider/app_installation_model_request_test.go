package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToXContentfulMarketplaceHeaderValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		model               AppInstallationModel
		expected            cm.OptString
		expectedDiagnostics []string
	}{
		"absent": {
			model:    AppInstallationModel{},
			expected: cm.OptString{},
		},
		"null": {
			model: AppInstallationModel{
				Marketplace: types.SetNull(types.StringType),
			},
			expected: cm.OptString{},
		},
		"unknown": {
			model: AppInstallationModel{
				Marketplace: types.SetUnknown(types.StringType),
			},
			expected:            cm.OptString{},
			expectedDiagnostics: []string{"marketplace"},
		},
		"null child": {
			model: AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringNull()}),
			},
			expected:            cm.OptString{},
			expectedDiagnostics: []string{"marketplace[Value(<null>)]"},
		},
		"unknown child": {
			model: AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringUnknown()}),
			},
			expected:            cm.OptString{},
			expectedDiagnostics: []string{"marketplace[Value(<unknown>)]"},
		},
		"mixed valid and unknown children fails closed": {
			model: AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("foo"),
					types.StringUnknown(),
				}),
			},
			expected:            cm.OptString{},
			expectedDiagnostics: []string{"marketplace[Value(<unknown>)]"},
		},
		"empty": {
			model: AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{}),
			},
			expected: cm.OptString{},
		},
		"foo": {
			model: AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("foo")}),
			},
			expected: cm.NewOptString("foo"),
		},
		"foo,bar": {
			model: AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("foo"), types.StringValue("bar")}),
			},
			expected: cm.NewOptString("bar,foo"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, diags := test.model.ToXContentfulMarketplaceHeaderValue()

			assert.Equal(t, test.expected, value)

			if len(test.expectedDiagnostics) > 0 {
				require.True(t, diags.HasError())
				assert.Equal(t, test.expectedDiagnostics, attributeDiagnosticPaths(t, diags))
			} else {
				assert.Empty(t, diags.Errors())
			}
		})
	}
}

func TestToAppInstallationData(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		model               AppInstallationModel
		diagnosticPath      string
		expectedRequestBody string
	}{
		"null": {
			model: AppInstallationModel{
				Parameters: types.MapNull(jsontypes.NormalizedType{}),
			},
			expectedRequestBody: testJSON(map[string]any{}),
		},
		"unknown": {
			model: AppInstallationModel{
				Parameters: types.MapUnknown(jsontypes.NormalizedType{}),
			},
			diagnosticPath:      "parameters",
			expectedRequestBody: testJSON(map[string]any{}),
		},
		"empty": {
			model: AppInstallationModel{
				Parameters: testInstallationParameterMap(nil),
			},
			expectedRequestBody: testJSON(map[string]any{"parameters": map[string]any{}}),
		},
		"foo=bar": {
			model: AppInstallationModel{
				Parameters: testInstallationParameterMap(map[string]string{"foo": `"bar"`}),
			},
			expectedRequestBody: testJSON(map[string]any{"parameters": map[string]any{"foo": "bar"}}),
		},
		"bare null member": {
			model:          AppInstallationModel{Parameters: types.MapValueMust(jsontypes.NormalizedType{}, map[string]attr.Value{"foo": jsontypes.NewNormalizedNull()})},
			diagnosticPath: `parameters["foo"]`, expectedRequestBody: "{}",
		},
		"unknown member": {
			model:          AppInstallationModel{Parameters: types.MapValueMust(jsontypes.NormalizedType{}, map[string]attr.Value{"foo": jsontypes.NewNormalizedUnknown()})},
			diagnosticPath: `parameters["foo"]`, expectedRequestBody: "{}",
		},
		"JSON null member": {
			model:               AppInstallationModel{Parameters: testInstallationParameterMap(map[string]string{"foo": "null"})},
			expectedRequestBody: `{"parameters":{"foo":null}}`,
		},
		"invalid": {
			model: AppInstallationModel{
				Parameters: testInstallationParameterMap(map[string]string{"foo": "invalid"}),
			},
			diagnosticPath:      `parameters["foo"]`,
			expectedRequestBody: "{}",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, diags := test.model.ToAppInstallationData()

			requestBody, err := req.MarshalJSON()
			require.NoError(t, err)

			assert.Equal(t, test.expectedRequestBody, string(requestBody))

			if test.diagnosticPath != "" {
				require.True(t, diags.HasError())
				assert.Equal(t, cm.AppInstallationData{}, req)
				assert.Equal(t, []string{test.diagnosticPath}, attributeDiagnosticPaths(t, diags))
			} else {
				assert.Empty(t, diags.Errors())
			}

			assert.Empty(t, diags.Warnings())
		})
	}
}

func testInstallationParameterMap(members map[string]string) types.Map {
	elements := make(map[string]attr.Value, len(members))
	for id, raw := range members {
		elements[id] = jsontypes.NewNormalizedValue(raw)
	}

	return types.MapValueMust(jsontypes.NormalizedType{}, elements)
}
