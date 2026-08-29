package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func knownStringListElements(valuePath path.Path, elements []types.String) ([]string, diag.Diagnostics) {
	values := make([]string, 0, len(elements))
	diags := diag.Diagnostics{}

	for index, element := range elements {
		elementPath := valuePath.AtListIndex(index)
		value, valueDiags := requestRequiredString(element, elementPath)
		diags.Append(valueDiags...)

		if !valueDiags.HasError() {
			values = append(values, value)
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return values, diags
}

func knownOptionalStringSetElements(valuePath path.Path, value types.Set) ([]string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown string set",
			"The string set must be known before it can be sent to Contentful.",
		)

		return nil, diags
	case value.IsNull():
		return []string{}, diags
	}

	values := make([]string, 0, len(value.Elements()))

	for _, rawElement := range value.Elements() {
		elementPath := valuePath.AtSetValue(rawElement)
		element, ok := rawElement.(types.String)

		if !ok {
			diags.AddAttributeError(
				elementPath,
				"Unexpected string set element type",
				"The set element must be a string before it can be sent to Contentful.",
			)

			continue
		}

		converted, valueDiags := requestRequiredString(element, elementPath)
		diags.Append(valueDiags...)

		if !valueDiags.HasError() {
			values = append(values, converted)
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return values, diags
}
