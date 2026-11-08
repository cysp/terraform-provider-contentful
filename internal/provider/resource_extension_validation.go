package provider

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type extensionConfigurationValidator struct{}

var _ validator.Object = extensionConfigurationValidator{}

func (extensionConfigurationValidator) Description(context.Context) string {
	return "Recognized Extension field types and parameter definitions must use their corresponding nested attributes and JSON values."
}

func (v extensionConfigurationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (extensionConfigurationValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	resp.Diagnostics.Append(validateExtensionConfiguration(req.ConfigValue, req.Path)...)
}

func validateExtensionPlannedConfiguration(ctx context.Context, plan tfsdk.Plan) diag.Diagnostics {
	var extension types.Object

	valuePath := path.Root("extension")

	diags := plan.GetAttribute(ctx, valuePath, &extension)
	if !diags.HasError() {
		diags.Append(validateExtensionConfiguration(extension, valuePath)...)
	}

	return diags
}

func validateExtensionConfiguration(extension types.Object, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if extension.IsNull() || extension.IsUnknown() {
		return diags
	}

	attributes := extension.Attributes()
	if fields, ok := attributes["field_types"].(types.List); ok {
		for index, value := range fields.Elements() {
			if field, ok := value.(types.Object); ok {
				fieldPath := valuePath.AtName("field_types").AtListIndex(index)
				diags.Append(validateFieldTypeAttributes(field, fieldPath, "Invalid Extension field type configuration")...)
				diags.Append(validateArrayFieldItemType(field, fieldPath, "Invalid Extension field type configuration")...)
			}
		}
	}

	parameters, ok := attributes["parameters"].(types.Object)
	if !ok || parameters.IsNull() || parameters.IsUnknown() {
		return diags
	}

	for _, scope := range []string{"installation", "instance"} {
		list, ok := parameters.Attributes()[scope].(types.List)
		if !ok {
			continue
		}

		for index, value := range list.Elements() {
			if parameter, ok := value.(types.Object); ok {
				diags.Append(validateExtensionParameter(parameter, valuePath.AtName("parameters").AtName(scope).AtListIndex(index))...)
			}
		}
	}

	return diags
}

func validateExtensionParameter(parameter types.Object, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if parameter.IsNull() || parameter.IsUnknown() {
		return diags
	}

	attributes := parameter.Attributes()

	kind, ok := attributes["type"].(types.String)
	if !ok || kind.IsNull() || kind.IsUnknown() {
		return diags
	}

	switch kind.ValueString() {
	case "Symbol", "Enum", "Secret", "Number", "Boolean":
	default:
		return diags
	}

	defaultValue, _ := attributes["default"].(jsontypes.Normalized)
	diags.Append(validateExtensionParameterDefault(kind.ValueString(), defaultValue, valuePath.AtName("default"))...)

	options, ok := attributes["options"].(TypedList[jsontypes.Normalized])
	if !ok || options.IsUnknown() {
		return diags
	}

	if kind.ValueString() != "Enum" {
		if !options.IsNull() {
			diags.AddAttributeError(valuePath.AtName("options"), "Invalid Extension parameter configuration", "Only Enum parameter definitions accept options. Omit options for this parameter type.")
		}

		return diags
	}

	values, complete, optionDiags := validateExtensionEnumOptions(options, valuePath.AtName("options"))
	diags.Append(optionDiags...)

	if complete && !diags.HasError() && !defaultValue.IsNull() && !defaultValue.IsUnknown() {
		var value string

		err := json.Unmarshal([]byte(defaultValue.ValueString()), &value)
		if err == nil && !values[value] {
			diags.AddAttributeError(valuePath.AtName("default"), "Invalid Extension parameter configuration", "An Enum default must match an option value. For labeled options, use the object key, not its label.")
		}
	}

	return diags
}

func validateExtensionParameterDefault(kind string, defaultValue jsontypes.Normalized, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !defaultValue.IsNull() && !defaultValue.IsUnknown() {
		raw := strings.TrimSpace(defaultValue.ValueString())

		valid := json.Valid([]byte(raw))
		if valid {
			switch kind {
			case "Symbol", "Enum", "Secret":
				valid = raw[0] == '"'
			case "Number":
				valid = raw[0] == '-' || (raw[0] >= '0' && raw[0] <= '9')
			case "Boolean":
				valid = raw == "true" || raw == "false"
			}
		}

		if !valid {
			diags.AddAttributeError(valuePath, "Invalid Extension parameter configuration", "The default must contain one JSON value matching the parameter type: string for Symbol, Enum, or Secret; number for Number; boolean for Boolean. Omit default to leave it unspecified.")
		}
	}

	return diags
}

func validateExtensionEnumOptions(options TypedList[jsontypes.Normalized], valuePath path.Path) (map[string]bool, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	values := map[string]bool{}
	complete := true

	if options.IsNull() || len(options.Elements()) == 0 {
		diags.AddAttributeError(valuePath, "Invalid Extension parameter configuration", "An Enum parameter requires at least one option.")

		return values, false, diags
	}

	shape := ""

	for index, option := range options.Elements() {
		optionPath := valuePath.AtListIndex(index)

		if option.IsUnknown() {
			complete = false

			continue
		}

		value, kind, valid := extensionEnumOption(option)
		if !valid {
			diags.AddAttributeError(optionPath, "Invalid Extension parameter configuration", "An Enum option must be a nonempty JSON string or a JSON object with exactly one member whose value is a string label.")

			complete = false

			continue
		}

		values[value] = true

		if shape != "" && shape != kind {
			diags.AddAttributeError(optionPath, "Invalid Extension parameter configuration", "Enum options must be all JSON strings or all labeled JSON objects; do not mix the two forms.")
		}

		shape = kind
	}

	return values, complete, diags
}

func extensionEnumOption(option jsontypes.Normalized) (string, string, bool) {
	if option.IsNull() {
		return "", "", false
	}

	raw := strings.TrimSpace(option.ValueString())
	if !json.Valid([]byte(raw)) {
		return "", "", false
	}

	if raw[0] == '"' {
		var value string

		err := json.Unmarshal([]byte(raw), &value)
		if err != nil {
			return "", "", false
		}

		return value, "string", value != ""
	}

	var object map[string]json.RawMessage

	err := json.Unmarshal([]byte(raw), &object)
	if err != nil || len(object) != 1 {
		return "", "", false
	}

	for key, label := range object {
		var value *string

		err := json.Unmarshal(label, &value)
		if err != nil || value == nil {
			return "", "", false
		}

		return key, "object", true
	}

	return "", "", false
}
