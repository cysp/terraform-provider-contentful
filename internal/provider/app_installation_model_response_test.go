package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppInstallationResponsePreservesReferenceValues(t *testing.T) {
	t.Parallel()

	reference := AppInstallationModel{
		Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("terms"), types.StringValue("privacy")}),
		Parameters:  testInstallationParameterMap(map[string]string{"visible": `"prior"`, "secret": `"retained secret"`}),
		Timeouts: timeouts.Value{Object: types.ObjectValueMust(
			map[string]attr.Type{"create": types.StringType, "update": types.StringType},
			map[string]attr.Value{"create": types.StringValue("2m"), "update": types.StringValue("3m")},
		)},
	}
	response := cm.AppInstallation{
		Sys:        cm.NewAppInstallationSys("space", "environment", "app"),
		Parameters: []byte(testJSON(map[string]any{"visible": "remote", "secret": "***"})),
	}

	actual, diags := NewAppInstallationResourceModelFromResponse(response, reference)

	require.Empty(t, diags)
	assert.Equal(t, AppInstallationModel{
		ID:              types.StringValue("space/environment/app"),
		SpaceID:         types.StringValue("space"),
		EnvironmentID:   types.StringValue("environment"),
		AppDefinitionID: types.StringValue("app"),
		Marketplace:     reference.Marketplace,
		Timeouts:        reference.Timeouts,
		Parameters:      testInstallationParameterMap(map[string]string{"visible": `"remote"`, "secret": `"retained secret"`}),
	}, actual)
}

func TestAppInstallationParameterResponseFidelity(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{`{"value":{"e\u0301":["e\u0301🌸",9007199254740993123456789]}}`, "{\"value\":{\"e\u0301\":[\"e\u0301🌸\",9007199254740993123456789]}}"} {
		model, diags := NewAppInstallationResourceModelFromResponse(cm.AppInstallation{Parameters: []byte(raw)}, AppInstallationModel{})
		require.Empty(t, diags)
		assert.Equal(t, jsontypes.NewNormalizedValue(`{"e\u0301":["e\u0301\ud83c\udf38",9007199254740993123456789]}`), model.Parameters.Elements()["value"])
		request, diags := model.ToAppInstallationData()
		require.Empty(t, diags)

		const expectedWire = `{"value":{"e\u0301":["e\u0301\ud83c\udf38",9007199254740993123456789]}}`

		//nolint:testifylint // JSONEq decodes numbers to float64 and discards the exact wire representation.
		assert.Equal(t, expectedWire, string(request.Parameters))
	}
}

func TestAppInstallationParameterResponseDiagnostics(t *testing.T) {
	t.Parallel()

	const secret = "private-response-sentinel"

	for _, test := range []struct {
		name    string
		raw     []byte
		summary string
	}{
		{name: "omitted"},
		{name: "JSON null", raw: []byte("null")},
		{name: "malformed JSON", raw: []byte(`{"token":` + testJSON(secret)), summary: "Invalid installation parameters"},
		{name: "array", raw: []byte(testJSON([]string{secret})), summary: "Invalid installation parameters"},
		{name: "non-NFC parameter ID", raw: []byte(testJSON(map[string]string{"e\u0301": secret})), summary: "Unrepresentable installation parameter ID"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, diags := NewAppInstallationResourceModelFromResponse(cm.AppInstallation{Parameters: test.raw}, AppInstallationModel{})
			assert.True(t, model.Parameters.IsNull())

			if test.summary == "" {
				require.Empty(t, diags)

				return
			}

			require.Len(t, diags.Errors(), 1)
			assert.Equal(t, []string{"parameters"}, attributeDiagnosticPaths(t, diags))
			assert.Equal(t, test.summary, diags[0].Summary())
			assert.NotContains(t, diags[0].Detail(), secret)
		})
	}
}
