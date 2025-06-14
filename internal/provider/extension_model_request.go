package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *ExtensionResourceModel) ToExtensionFields() (cm.ExtensionFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.ExtensionFields{}

	switch {
	case model.Parameters.IsUnknown():
		diags.AddAttributeWarning(path.Root("parameters"), "Failed to update extension parameters", "Parameters are unknown")
	case model.Parameters.IsNull():
	default:
		fields.Parameters = []byte(model.Parameters.ValueString())
	}

	return fields, diags
}
