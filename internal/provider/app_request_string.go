package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func appRequestOptionalString(value types.String, valuePath path.Path) (*string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown request value",
			"This value must be known before it can be sent to Contentful.",
		)
	case value.IsNull():
		return nil, diags
	default:
		return value.ValueStringPointer(), diags
	}

	return nil, diags
}
