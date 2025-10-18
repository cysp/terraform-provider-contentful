package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func (m *EntryModel) ToEntryRequestFields(ctx context.Context) (cm.EntryFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Treat fields as an opaque JSON blob
	var fields cm.EntryFields

	normalized := m.Fields.Value()
	// If the value is null, return empty fields
	if normalized.IsNull() {
		return cm.EntryFields{}, diags
	}

	// Use package-level jsontypes.Unwrap
	if f, ok := jsontypes.UnwrapNormalized(normalized).(map[string]any); ok {
		fields = f
	} else {
		diags.AddError("Invalid entry fields type", "Expected map[string]any for entry fields, got something else.")
		return cm.EntryFields{}, diags
	}

	// Metadata/tags are handled separately if needed by the API
	return fields, diags
}
