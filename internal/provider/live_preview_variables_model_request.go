package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type livePreviewVariablesValidator struct{}

var _ validator.String = livePreviewVariablesValidator{}

func (livePreviewVariablesValidator) Description(context.Context) string {
	return "must be a JSON object of strings, nulls, or locale objects containing strings/nulls"
}

func (v livePreviewVariablesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (livePreviewVariablesValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.Diagnostics.Append(validateLivePreviewVariables([]byte(req.ConfigValue.ValueString()), req.Path)...)
}

func validateLivePreviewVariables(raw []byte, valuePath path.Path) diag.Diagnostics {
	var (
		value       any
		diagnostics diag.Diagnostics
	)

	err := json.Unmarshal(raw, &value)
	if err != nil {
		diagnostics.AddAttributeError(valuePath, "Invalid live preview variables", "Variables must be valid JSON.")

		return diagnostics
	}

	variables, ok := value.(map[string]any)
	if !ok {
		diagnostics.AddAttributeError(valuePath, "Invalid live preview variables", "Variables must be a JSON object; arrays and a null root are not supported.")

		return diagnostics
	}

	for _, name := range slices.Sorted(maps.Keys(variables)) {
		switch variable := variables[name].(type) {
		case nil, string:
		case map[string]any:
			for _, locale := range slices.Sorted(maps.Keys(variable)) {
				switch variable[locale].(type) {
				case nil, string:
				default:
					diagnostics.AddAttributeError(valuePath, "Invalid live preview variable", fmt.Sprintf("The value at $[%q][%q] must be a string or null.", name, locale))
				}
			}
		default:
			diagnostics.AddAttributeError(valuePath, "Invalid live preview variable", fmt.Sprintf("The value at $[%q] must be a string, null, or locale object containing strings/nulls.", name))
		}
	}

	return diagnostics
}

func (model LivePreviewVariablesModel) ToLivePreviewVariablesData() (cm.LivePreviewVariablesData, diag.Diagnostics) {
	value, diagnostics := requestRequiredString(model.Variables.StringValue, path.Root("variables"))
	if diagnostics.HasError() {
		return cm.LivePreviewVariablesData{}, diagnostics
	}

	diagnostics.Append(validateLivePreviewVariables([]byte(value), path.Root("variables"))...)

	if diagnostics.HasError() {
		return cm.LivePreviewVariablesData{}, diagnostics
	}

	return cm.LivePreviewVariablesData{Variables: []byte(value)}, diagnostics
}
