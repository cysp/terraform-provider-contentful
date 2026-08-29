package provider

import (
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func knownOptionalListResourceString(valuePath path.Path, value types.String) (string, bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown list configuration value",
			"The string value must be known before Contentful resources can be listed.",
		)

		return "", false, diags
	}

	if value.IsNull() {
		return "", false, diags
	}

	return value.ValueString(), true, diags
}

func knownOptionalListResourceStringList(valuePath path.Path, value TypedList[types.String]) ([]string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown list configuration value",
			"The list must be known before Contentful resources can be listed.",
		)

		return nil, diags
	}

	if value.IsNull() {
		return nil, diags
	}

	return knownStringListElements(valuePath, value.Elements())
}

func knownOptionalListResourceStringMap(valuePath path.Path, value TypedMap[types.String]) (map[string]string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown list configuration value",
			"The map must be known before Contentful resources can be listed.",
		)

		return nil, diags
	}

	if value.IsNull() {
		return nil, diags
	}

	elements := value.Elements()

	keys := make([]string, 0, len(elements))
	for key := range elements {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	values := make(map[string]string, len(elements))
	for _, key := range keys {
		element := elements[key]
		elementPath := valuePath.AtMapKey(key)
		converted, valueDiags := requestRequiredString(element, elementPath)
		diags.Append(valueDiags...)

		if !valueDiags.HasError() {
			values[key] = converted
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return values, diags
}
