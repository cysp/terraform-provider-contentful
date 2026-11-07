package provider

import (
	"context"
	"encoding/json"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type appDefinitionParametersValidator struct{}

var _ validator.Object = appDefinitionParametersValidator{}

func (appDefinitionParametersValidator) Description(context.Context) string {
	return "App Definition parameter defaults and options must use supported JSON shapes."
}

func (v appDefinitionParametersValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (appDefinitionParametersValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	resp.Diagnostics.Append(validateAppDefinitionParameters(req.ConfigValue, req.Path)...)
}

func validateAppDefinitionParameters(parameters types.Object, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if parameters.IsNull() || parameters.IsUnknown() {
		return diags
	}

	for _, scope := range []string{"installation", "instance"} {
		list, ok := parameters.Attributes()[scope].(types.List)
		if !ok || list.IsNull() || list.IsUnknown() {
			continue
		}

		if scope == "instance" && len(list.Elements()) > 8 {
			diags.AddAttributeError(valuePath.AtName(scope), "Invalid App Definition parameter configuration", "At most 8 instance parameter definitions are supported. Reduce the number of declarations.")
		}

		for i, element := range list.Elements() {
			parameter, ok := element.(types.Object)
			if !ok || parameter.IsNull() || parameter.IsUnknown() {
				continue
			}

			diags.Append(validateAppDefinitionParameter(parameter, valuePath.AtName(scope).AtListIndex(i))...)
		}
	}

	return diags
}

func validateAppDefinitionParameter(parameter types.Object, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	attributes := parameter.Attributes()

	kind, ok := attributes["type"].(types.String)
	if !ok || kind.IsNull() || kind.IsUnknown() {
		return diags
	}

	switch kind.ValueString() {
	case "Symbol", "Number", "Boolean", "Enum", "Secret":
	default:
		return diags
	}

	if value, ok := attributes["default"].(jsontypes.Normalized); ok {
		decoded, valid := appDefinitionParameterJSON(value)
		if valid {
			switch decoded.(type) {
			case string, bool, json.Number:
			default:
				diags.AddAttributeError(valuePath.AtName("default"), "Invalid App Definition parameter configuration", "A parameter default must be a JSON string, number, or Boolean. Omit default to leave it unspecified.")
			}
		}
	}

	diags.Append(validateAppDefinitionParameterOptions(attributes["options"], valuePath.AtName("options"))...)

	return diags
}

func validateAppDefinitionParameterOptions(value attr.Value, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	options, ok := value.(TypedList[jsontypes.Normalized])
	if !ok || options.IsNull() || options.IsUnknown() {
		return diags
	}

	if len(options.Elements()) == 0 {
		diags.AddAttributeError(valuePath, "Invalid App Definition parameter configuration", "When configured, options must contain at least one option. Omit options to leave them unspecified.")
	}

	for index, value := range options.Elements() {
		decoded, valid := appDefinitionParameterJSON(value)
		if !valid {
			continue
		}

		validShape := false

		switch option := decoded.(type) {
		case string:
			validShape = option != ""
		case map[string]any:
			count, unique := appDefinitionOptionMemberCount(value.ValueString())
			// Do not infer server duplicate-key semantics from map decoding.
			validShape = !unique || count == 1
		}

		if !validShape {
			diags.AddAttributeError(valuePath.AtListIndex(index), "Invalid App Definition parameter configuration", "An option must be a nonempty JSON string or a JSON object with exactly one member. The member value is preserved without coercion.")
		}
	}

	return diags
}

// Syntax diagnostics belong to the custom type during configuration validation.
// Keep original bytes and numbers; this reader never normalizes a request value.
func appDefinitionParameterJSON(value jsontypes.Normalized) (any, bool) {
	if value.IsNull() || value.IsUnknown() {
		return nil, false
	}

	decoder := json.NewDecoder(strings.NewReader(value.ValueString()))
	decoder.UseNumber()

	var decoded any
	if decoder.Decode(&decoded) != nil {
		return nil, false
	}

	if decoder.Decode(new(any)) != io.EOF {
		return nil, false
	}

	return decoded, true
}

func appDefinitionOptionMemberCount(raw string) (int, bool) {
	decoder := json.NewDecoder(strings.NewReader(raw))

	_, err := decoder.Token()
	if err != nil {
		return 0, false
	}

	seen := map[string]bool{}

	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return 0, false
		}

		key, ok := token.(string)
		if !ok || seen[key] {
			return 0, false
		}

		seen[key] = true

		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return 0, false
		}
	}

	return len(seen), true
}

func validateAppDefinitionPlannedParameters(ctx context.Context, plan tfsdk.Plan) diag.Diagnostics {
	valuePath := path.Root("parameters")

	var parameters types.Object

	diags := plan.GetAttribute(ctx, valuePath, &parameters)
	if diags.HasError() {
		return diags
	}

	diags.Append(requireKnownAppDefinitionParameterValue(parameters, valuePath)...)

	if diags.HasError() {
		return diags
	}

	diags.Append(validateAppDefinitionParameters(parameters, valuePath)...)

	return diags
}

func requireKnownAppDefinitionParameterValue(value attr.Value, valuePath path.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	if value.IsUnknown() {
		diags.AddAttributeError(valuePath, "Unknown App Definition parameter value", "This value must be known before sending the App Definition to Contentful.")

		return diags
	}

	if value.IsNull() {
		return diags
	}

	switch value := value.(type) {
	case types.Object:
		a := value.Attributes()
		for _, name := range slices.Sorted(maps.Keys(a)) {
			child := a[name]

			childPath := valuePath.AtName(name)
			if (name == "id" || name == "name" || name == "type") && child.IsNull() {
				diags.AddAttributeError(childPath, "Invalid App Definition parameter configuration", "This parameter attribute must not be null.")
			} else {
				diags.Append(requireKnownAppDefinitionParameterValue(child, childPath)...)
			}
		}
	case types.List:
		for i, child := range value.Elements() {
			p := valuePath.AtListIndex(i)
			if child.IsNull() {
				diags.AddAttributeError(p, "Invalid App Definition parameter configuration", "List elements must not be null.")
			} else {
				diags.Append(requireKnownAppDefinitionParameterValue(child, p)...)
			}
		}
	case TypedList[jsontypes.Normalized]:
		for i, child := range value.Elements() {
			p := valuePath.AtListIndex(i)
			if child.IsNull() {
				diags.AddAttributeError(p, "Invalid App Definition parameter configuration", "List elements must not be null.")
			} else {
				diags.Append(requireKnownAppDefinitionParameterValue(child, p)...)
			}
		}
	case jsontypes.Normalized:
		if _, valid := appDefinitionParameterJSON(value); !valid {
			diags.AddAttributeError(valuePath, "Invalid App Definition parameter configuration", "This value must contain exactly one valid JSON value.")
		}
	}

	return diags
}
