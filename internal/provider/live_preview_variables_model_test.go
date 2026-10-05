package provider_test

import (
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func livePreviewVariablesModel(variables string) LivePreviewVariablesModel {
	return LivePreviewVariablesModel{
		IDIdentityModel:                   NewIDIdentityModelFromMultipartID("space", "environment"),
		LivePreviewVariablesIdentityModel: LivePreviewVariablesIdentityModel{SpaceID: types.StringValue("space"), EnvironmentID: types.StringValue("environment")},
		Variables:                         jsontypes.NewNormalizedValue(variables),
		Timeouts:                          TimeoutsNull(),
	}
}

//nolint:gosec // Synthetic sentinel verifies diagnostic redaction.
func TestLivePreviewVariablesValidationAndRequest(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value    jsontypes.Normalized
		invalid  bool
		deferred bool
		location string
	}{
		"mixed stable values": {value: jsontypes.NewNormalizedValue(testJSON(map[string]any{
			"global":    "value",
			"localized": map[string]any{"en-US": "", "fr-FR": nil},
			"null":      nil,
			"empty":     map[string]any{},
		}))},
		"empty object": {value: jsontypes.NewNormalizedValue(testJSON(map[string]any{}))},
		"unrestricted names and locale validation delegated": {value: jsontypes.NewNormalizedValue(testJSON(map[string]any{
			"":            "",
			"a.b[c]/d":    "",
			"café":        "",
			"__proto__":   "",
			"constructor": "",
			"locale":      map[string]any{"not-enabled": "value"},
		}))},
		"long string left to CMA":             {value: jsontypes.NewNormalizedValue(testJSON(map[string]any{"text": strings.Repeat("x", 50001)}))},
		"literal whitespace and placeholders": {value: jsontypes.NewNormalizedValue(` { "host": " https://preview.invalid/{entry.sys.id}\n" } `)},
		"unknown":                             {value: jsontypes.NewNormalizedUnknown(), invalid: true, deferred: true},
		"Terraform null":                      {value: jsontypes.NewNormalizedNull(), invalid: true, deferred: true},
		"invalid JSON":                        {value: jsontypes.NewNormalizedValue(`{"secret":"VALUE_DO_NOT_ECHO"`), invalid: true},
		"trailing JSON":                       {value: jsontypes.NewNormalizedValue(`{} {}`), invalid: true},
		"JSON null root":                      {value: jsontypes.NewNormalizedValue(testJSON(nil)), invalid: true},
		"array root":                          {value: jsontypes.NewNormalizedValue(testJSON([]any{})), invalid: true},
		"string root":                         {value: jsontypes.NewNormalizedValue(testJSON("VALUE_DO_NOT_ECHO")), invalid: true},
		"number root":                         {value: jsontypes.NewNormalizedValue(testJSON(123)), invalid: true},
		"boolean root":                        {value: jsontypes.NewNormalizedValue(testJSON(true)), invalid: true},
		"variable array":                      {value: jsontypes.NewNormalizedValue(testJSON(map[string]any{"bad": []any{}})), invalid: true, location: `$["bad"]`},
		"variable number":                     {value: jsontypes.NewNormalizedValue(testJSON(map[string]any{"bad": 1})), invalid: true},
		"variable boolean":                    {value: jsontypes.NewNormalizedValue(testJSON(map[string]any{"bad": false})), invalid: true},
		"localized number":                    {value: jsontypes.NewNormalizedValue(testJSON(map[string]any{"bad": map[string]any{"en-US": 1}})), invalid: true, location: `$["bad"]["en-US"]`},
		"localized boolean":                   {value: jsontypes.NewNormalizedValue(testJSON(map[string]any{"bad": map[string]any{"en-US": false}})), invalid: true},
		"localized array":                     {value: jsontypes.NewNormalizedValue(testJSON(map[string]any{"bad": map[string]any{"en-US": []any{}}})), invalid: true},
		"localized object":                    {value: jsontypes.NewNormalizedValue(testJSON(map[string]any{"a\"b": map[string]any{"en-US": map[string]any{"secret": "VALUE_DO_NOT_ECHO"}}})), invalid: true, location: `$["a\"b"]["en-US"]`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			attribute, ok := LivePreviewVariablesResourceSchema(t.Context()).Attributes["variables"].(schema.StringAttribute)
			require.True(t, ok)
			require.True(t, attribute.Required)
			require.False(t, attribute.Optional)
			require.False(t, attribute.Computed)
			require.False(t, attribute.Sensitive)

			validation := validator.StringResponse{}
			for _, valueValidator := range attribute.Validators {
				valueValidator.ValidateString(t.Context(), validator.StringRequest{Path: path.Root("variables"), ConfigValue: test.value.StringValue}, &validation)
			}

			require.Equal(t, test.invalid && !test.deferred, validation.Diagnostics.HasError())

			model := livePreviewVariablesModel(testJSON(map[string]any{}))
			model.Variables = test.value
			request, diagnostics := model.ToLivePreviewVariablesData()
			require.Equal(t, test.invalid, diagnostics.HasError(), diagnostics)

			if !test.invalid {
				require.Equal(t, test.value.ValueString(), string(request.Variables))

				return
			}

			require.Empty(t, request.Variables)
			require.Equal(t, []string{"variables"}, attributeDiagnosticPaths(t, diagnostics))

			for _, diagnostic := range diagnostics {
				assert.NotContains(t, diagnostic.Detail(), "VALUE_DO_NOT_ECHO")

				if test.location != "" {
					assert.Contains(t, diagnostic.Detail(), test.location)
				}
			}
		})
	}
}

func TestLivePreviewVariablesResponseReconciliation(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		remote           string
		planned          string
		unknown          bool
		spaceID          string
		environmentID    string
		plannedID        string
		paths            []string
		projectionError  bool
		consistencyError bool
	}{
		"object order and whitespace equivalent": {planned: ` { "z": null, "a": {"en-US":""} } `, remote: `{"a":{"en-US":""},"z":null}`},
		"null is not absent":                     {planned: testJSON(map[string]any{"a": nil}), remote: testJSON(map[string]any{}), consistencyError: true},
		"empty string is not null":               {planned: testJSON(map[string]any{"a": ""}), remote: testJSON(map[string]any{"a": nil}), consistencyError: true},
		"empty map is not null":                  {planned: testJSON(map[string]any{"a": map[string]any{}}), remote: testJSON(map[string]any{"a": nil}), consistencyError: true},
		"array is not empty map":                 {planned: testJSON(map[string]any{"a": map[string]any{}}), remote: testJSON(map[string]any{"a": []any{}}), consistencyError: true},
		"different string":                       {planned: testJSON(map[string]any{"a": "planned"}), remote: testJSON(map[string]any{"a": "remote"}), consistencyError: true},
		"extra key":                              {planned: testJSON(map[string]any{}), remote: testJSON(map[string]any{"extra": "remote"}), consistencyError: true},
		"remote irregular JSON preserved":        {unknown: true, remote: testJSON(map[string]any{"future": map[string]any{"nested": []any{1, true}}, "large": int64(9007199254740993)})},
		"explicit remote null preserved":         {unknown: true, remote: testJSON(nil)},
		"missing JSON":                           {planned: testJSON(map[string]any{}), projectionError: true},
		"invalid JSON":                           {planned: testJSON(map[string]any{}), remote: `{`, projectionError: true},
		"space mismatch":                         {planned: testJSON(map[string]any{"a": "value"}), remote: testJSON(map[string]any{"a": "value"}), spaceID: "other", consistencyError: true, paths: []string{"space_id"}},
		"environment mismatch":                   {planned: testJSON(map[string]any{"a": "value"}), remote: testJSON(map[string]any{"a": "value"}), environmentID: "other", consistencyError: true, paths: []string{"environment_id"}},
		"legacy ID mismatch":                     {planned: testJSON(map[string]any{"a": "value"}), remote: testJSON(map[string]any{"a": "value"}), plannedID: "other/identity", consistencyError: true, paths: []string{"id"}},
		"identity and variables mismatch":        {planned: testJSON(map[string]any{"a": "planned"}), remote: testJSON(map[string]any{"a": "remote"}), spaceID: "other", environmentID: "other", consistencyError: true, paths: []string{"space_id", "environment_id", "variables"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			plan := livePreviewVariablesModel(test.planned)
			if test.unknown {
				plan.Variables = jsontypes.NewNormalizedUnknown()
			}

			if test.plannedID != "" {
				plan.ID = types.StringValue(test.plannedID)
			}

			response := cm.LivePreviewVariables{
				Sys:       cm.LivePreviewVariablesSys{Space: cm.NewSpaceLink("space"), Environment: cm.NewEnvironmentLink("environment"), Version: 11},
				Variables: []byte(test.remote),
			}
			if test.spaceID != "" {
				response.Sys.Space = cm.NewSpaceLink(test.spaceID)
			}

			if test.environmentID != "" {
				response.Sys.Environment = cm.NewEnvironmentLink(test.environmentID)
			}

			model, diagnostics, consistencyDiagnostics := ReconcileLivePreviewVariablesMutationResponse(t.Context(), response, plan)
			require.Equal(t, test.projectionError, diagnostics.HasError(), diagnostics)
			require.Equal(t, test.consistencyError, consistencyDiagnostics.HasError(), consistencyDiagnostics)
			require.Equal(t, plan.LivePreviewVariablesIdentityModel, model.LivePreviewVariablesIdentityModel)
			require.Equal(t, "space/environment", model.ID.ValueString())

			if test.paths != nil {
				require.Equal(t, test.paths, attributeDiagnosticPaths(t, consistencyDiagnostics))
			}

			if test.projectionError {
				return
			}

			require.False(t, model.Variables.IsUnknown())

			if !test.consistencyError && !test.unknown {
				require.Equal(t, test.planned, model.Variables.ValueString())
			} else {
				// These response fixtures are already normalized; exact comparison preserves number precision.
				require.Equal(t, test.remote, model.Variables.ValueString())
			}
		})
	}
}
