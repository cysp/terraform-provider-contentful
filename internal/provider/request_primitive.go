package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func requestRequiredString(value types.String, valuePath path.Path) (string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown string",
			"The string value must be known before it can be sent to Contentful.",
		)
	case value.IsNull():
		diags.AddAttributeError(
			valuePath,
			"Unexpected null string",
			"The required string value cannot be null.",
		)
	default:
		return value.ValueString(), diags
	}

	return "", diags
}

func requestRequiredBool(value types.Bool, valuePath path.Path) (bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown boolean",
			"The boolean value must be known before it can be sent to Contentful.",
		)
	case value.IsNull():
		diags.AddAttributeError(
			valuePath,
			"Unexpected null boolean",
			"The required boolean value cannot be null.",
		)
	default:
		return value.ValueBool(), diags
	}

	return false, diags
}

func requestOptionalString(value types.String, valuePath path.Path) (cm.OptString, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown string",
			"The string value must be known before it can be sent to Contentful.",
		)

		return cm.OptString{}, diags
	}

	if value.IsNull() {
		return cm.OptString{}, diags
	}

	return cm.NewOptString(value.ValueString()), diags
}

func requestOptionalBool(value types.Bool, valuePath path.Path) (cm.OptBool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown boolean",
			"The boolean value must be known before it can be sent to Contentful.",
		)

		return cm.OptBool{}, diags
	}

	if value.IsNull() {
		return cm.OptBool{}, diags
	}

	return cm.NewOptBool(value.ValueBool()), diags
}
