package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// requestRequiredString converts a Terraform string whose request contract is:
// null and unknown are invalid, while every known non-null value, including the
// empty string, is returned unchanged. Required describes the request value
// contract, not whether the Terraform schema attribute is Required.
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

// requestNullableString converts a Terraform string whose wire contract is:
// null sends an explicit null property, unknown is invalid at the request
// boundary, and every known value (including the empty string) is sent explicitly.
func requestNullableString(value types.String, valuePath path.Path) (cm.OptNilString, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown string",
			"The string value must be known before it can be sent to Contentful.",
		)

		return cm.OptNilString{}, diags
	}

	if value.IsNull() {
		return cm.NewOptNilStringNull(), diags
	}

	return cm.NewOptNilString(value.ValueString()), diags
}

// requestOmittableString converts a Terraform string whose wire contract is:
// null omits the property, unknown is invalid at the request boundary, and every
// known value (including the empty string) is sent explicitly.
func requestOmittableString(value types.String, valuePath path.Path) (cm.OptString, diag.Diagnostics) {
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

// requestOmittableBool converts a Terraform boolean whose wire contract is:
// null omits the property, unknown is invalid at the request boundary, and every
// known value (including false) is sent explicitly.
func requestOmittableBool(value types.Bool, valuePath path.Path) (cm.OptBool, diag.Diagnostics) {
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
