package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func (m *EntryModel) ToEntryRequestFields(ctx context.Context) (cm.EntryFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Treat fields as a map of opaque JSON blobs
	fields := map[string]any{}

	attrs := m.Fields.Elements()
	for k, v := range attrs {
		if v.IsNull() {
			continue
		}
		fields[k] = v.Unwrap()
	}

	// Metadata/tags are handled separately if needed by the API
	return fields, diags
}
