package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func (m *EntryModel) ToEntryRequestFields(ctx context.Context) (cm.EntryRequest, diag.Diagnostics) {
	fields, diags := m.toEntryFields(ctx)
	if diags.HasError() {
		return cm.EntryRequest{}, diags
	}

	metadata, diags := m.toEntryMetadata(ctx)
	if diags.HasError() {
		return cm.EntryRequest{}, diags
	}

	return cm.EntryRequest{
		Fields:   cm.NewOptEntryFields(fields),
		Metadata: metadata,
	}, diags
}

func (m *EntryModel) toEntryFields(ctx context.Context) (cm.EntryFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Treat fields as a map of opaque JSON blobs
	fields := make(cm.EntryFields)

	attrs := m.Fields.Elements()
	for k, v := range attrs {
		if v.IsNull() {
			continue
		}

		fields[k] = jx.Raw(v.ValueString())
	}

	return fields, diags
}

func (m *EntryModel) toEntryMetadata(ctx context.Context) (cm.OptEntryMetadata, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	var metadata cm.OptEntryMetadata
	if !m.Metadata.IsNull() {
		tags := []cm.TagLink{}
		for _, tag := range m.Metadata.Value().Tags.Elements() {
			tags = append(tags, cm.TagLink{
				Sys: cm.TagLinkSys{
					Type:     "Link",
					LinkType: "Tag",
					ID:       tag.ValueString(),
				},
			})
		}
		metadata.SetTo(cm.EntryMetadata{
			Tags: tags,
		})
	}
	return metadata, diags
}
