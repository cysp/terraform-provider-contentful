package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func contentTypeValidationField() map[string]any {
	return map[string]any{"id": "field", "name": "Field", "type": "Symbol", "localized": false, "required": false, "disabled": false, "omitted": false, "validations": []any{}}
}

func contentTypeValidationPlan(t *testing.T, fields []any) tfsdk.Plan {
	t.Helper()
	schema := ContentTypeResourceSchema(t.Context())
	raw, err := json.Marshal(map[string]any{"space_id": "space", "environment_id": "master", "content_type_id": "type", "name": "Type", "description": "", "display_field": "field", "fields": fields})
	require.NoError(t, err)
	value, err := tftypes.ValueFromJSONWithOpts(raw, schema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
	require.NoError(t, err)

	return tfsdk.Plan{Schema: schema, Raw: value}
}

func TestContentTypeJSONValidation(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)

	for _, attribute := range []string{"default_value", "validations", "items.validations"} {
		t.Run(attribute, func(t *testing.T) {
			t.Parallel()

			for _, test := range []struct {
				name, raw string
				invalid   bool
			}{
				{"empty object", "{}", false},
				{"opaque object", `{ "en-US":null,"value":9007199254740993,"future":{"enabled":false} }`, false},
				{"JSON null", "null", true},
				{"string", `"value"`, true},
				{"number", "42", true},
				{"Boolean", "false", true},
				{"array", "[]", true},
			} {
				t.Run(test.name, func(t *testing.T) {
					t.Parallel()

					field := contentTypeValidationField()
					valuePath := path.Root("fields").AtListIndex(0)
					diagnosticPath := tftypes.NewAttributePath().WithAttributeName("fields").WithElementKeyInt(0)

					switch attribute {
					case "default_value":
						field[attribute] = test.raw
						valuePath = valuePath.AtName(attribute)
						diagnosticPath = diagnosticPath.WithAttributeName(attribute)
					case "validations":
						field[attribute] = []any{test.raw}
						valuePath = valuePath.AtName(attribute).AtListIndex(0)
						diagnosticPath = diagnosticPath.WithAttributeName(attribute).WithElementKeyInt(0)
					case "items.validations":
						field["type"] = "Array"
						field["items"] = map[string]any{"type": "Symbol", "validations": []any{test.raw}}
						valuePath = valuePath.AtName("items").AtName("validations").AtListIndex(0)
						diagnosticPath = diagnosticPath.WithAttributeName("items").WithAttributeName("validations").WithElementKeyInt(0)
					}

					plan := contentTypeValidationPlan(t, []any{field})
					dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
					require.NoError(t, err)
					response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_content_type", Config: &dynamic})
					require.NoError(t, err)

					var model ContentTypeModel
					require.Empty(t, plan.Get(t.Context(), &model))

					converted, diags := model.ToContentTypeRequestData(t.Context(), model)
					if test.invalid {
						require.Len(t, response.Diagnostics, 1)
						assert.Equal(t, diagnosticPath, response.Diagnostics[0].Attribute)
						assert.Equal(t, "Invalid Content Type JSON object", response.Diagnostics[0].Summary)
						require.Equal(t, []string{valuePath.String()}, mutationDiagnosticPaths(t, diags))
						require.Nil(t, converted.Fields)
					} else {
						require.Empty(t, response.Diagnostics)
						require.Empty(t, diags)
						require.Len(t, converted.Fields, 1)

						var raw []byte

						switch attribute {
						case "default_value":
							raw = converted.Fields[0].DefaultValue
						case "validations":
							raw = converted.Fields[0].Validations[0]
						case "items.validations":
							raw = converted.Fields[0].Items.Value.Validations[0]
						}

						assert.Equal(t, test.raw, string(raw))
					}
				})
			}
		})
	}
}

func TestContentTypeDuplicateFieldValidation(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)
	plan := contentTypeValidationPlan(t, []any{contentTypeValidationField(), contentTypeValidationField()})
	dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
	require.NoError(t, err)
	response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_content_type", Config: &dynamic})
	require.NoError(t, err)
	require.Len(t, response.Diagnostics, 1)
	assert.Equal(t, tftypes.NewAttributePath().WithAttributeName("fields").WithElementKeyInt(1).WithAttributeName("id"), response.Diagnostics[0].Attribute)

	var model ContentTypeModel
	require.Empty(t, plan.Get(t.Context(), &model))
	_, diags := model.ToContentTypeRequestData(t.Context(), model)
	require.Equal(t, []string{"fields[1].id"}, mutationDiagnosticPaths(t, diags))
	require.Empty(t, plan.SetAttribute(t.Context(), path.Root("fields").AtListIndex(1).AtName("id"), types.StringUnknown()))
	dynamic, err = tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
	require.NoError(t, err)
	response, err = server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_content_type", Config: &dynamic})
	require.NoError(t, err)
	require.Empty(t, response.Diagnostics)
}

func TestContentTypeJSONRequestRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"", "{", "{} {}", "invalid json"} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			field := contentTypeValidationField()
			field["default_value"] = raw
			plan := contentTypeValidationPlan(t, []any{field})

			var model ContentTypeModel
			require.Empty(t, plan.Get(t.Context(), &model))
			converted, diags := model.ToContentTypeRequestData(t.Context(), model)
			require.Nil(t, converted.Fields)
			require.Equal(t, []string{"fields[0].default_value"}, mutationDiagnosticPaths(t, diags))
			validations, diags := ValidationsListToContentTypeRequestDataFieldValidations(path.Root("validations"), NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue(raw)}))
			require.Nil(t, validations)
			require.Equal(t, []string{"validations[0]"}, mutationDiagnosticPaths(t, diags))
		})
	}
}

func TestContentTypeValidationChecksKnownSiblings(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)
	plan := contentTypeValidationPlan(t, []any{contentTypeValidationField(), contentTypeValidationField(), contentTypeValidationField()})
	require.Empty(t, plan.SetAttribute(t.Context(), path.Root("fields").AtListIndex(0).AtName("id"), types.StringUnknown()))
	require.Empty(t, plan.SetAttribute(t.Context(), path.Root("fields").AtListIndex(0).AtName("default_value"), jsontypes.NewNormalizedUnknown()))
	require.Empty(t, plan.SetAttribute(t.Context(), path.Root("fields").AtListIndex(1).AtName("default_value"), jsontypes.NewNormalizedValue("[]")))
	dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
	require.NoError(t, err)
	response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_content_type", Config: &dynamic})
	require.NoError(t, err)
	require.Len(t, response.Diagnostics, 2)

	paths := make([]string, 0, len(response.Diagnostics))
	for _, diagnostic := range response.Diagnostics {
		paths = append(paths, diagnostic.Attribute.String())
	}

	require.ElementsMatch(t, []string{
		tftypes.NewAttributePath().WithAttributeName("fields").WithElementKeyInt(1).WithAttributeName("default_value").String(),
		tftypes.NewAttributePath().WithAttributeName("fields").WithElementKeyInt(2).WithAttributeName("id").String(),
	}, paths)
}
