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

		switch {
		case element.IsUnknown():
			diags.AddAttributeError(
				elementPath,
				"Unexpected unknown string list element",
				"The string value must be known before it can be sent to Contentful.",
			)
		case element.IsNull():
			diags.AddAttributeError(
				elementPath,
				"Unexpected null string list element",
				"The string value cannot be null when it is sent to Contentful.",
			)
		default:
			values = append(values, element.ValueString())
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return values, diags
}
