package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func webhookOptionalNullableString(value types.String, valuePath path.Path) (cm.OptNilString, diag.Diagnostics) {
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
