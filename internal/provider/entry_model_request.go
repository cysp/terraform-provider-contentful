package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (m *EntryModel) ToEntryRequestFields(ctx context.Context) (cm.EntryFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var fields cm.EntryFields
	diags.Append(tfsdk.ValueAs(ctx, m.Fields.Value(), &fields)...)
	if diags.HasError() {
		return cm.EntryFields{}, diags
	}

	// Metadata/tags are handled separately if needed by the API
	return fields, diags
}
