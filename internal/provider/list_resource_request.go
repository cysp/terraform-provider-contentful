package provider

import (
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func knownListResourceString(valuePath path.Path, value types.String) (string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown list configuration value",
			"The string value must be known before Contentful resources can be listed.",
		)
	case value.IsNull():
		diags.AddAttributeError(
			valuePath,
			"Unexpected null list configuration value",
			"The string value cannot be null when Contentful resources are listed.",
		)
	default:
		return value.ValueString(), diags
	}

	return "", diags
}

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

		switch {
		case element.IsUnknown():
			diags.AddAttributeError(
				elementPath,
				"Unexpected unknown list configuration value",
				"The map value must be known before Contentful resources can be listed.",
			)
		case element.IsNull():
			diags.AddAttributeError(
				elementPath,
				"Unexpected null list configuration value",
				"The map value cannot be null when Contentful resources are listed.",
			)
		default:
			values[key] = element.ValueString()
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return values, diags
}
