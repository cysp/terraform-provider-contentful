package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func (m *EntryModel) ToEntryRequest(ctx context.Context) (*cm.PutEntryReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Treat fields as a map of opaque JSON blobs
	fields := make(cm.EntryFields)

	attrs := m.Fields.Elements()
	for k, v := range attrs {
		if v.IsNull() {
			continue
		}
		var val any
		if err := v.UnmarshalTo(&val); err != nil {
			continue
		}
		fields[k] = val
	}

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

	return &cm.PutEntryReq{
		Fields:   fields,
		Metadata: metadata,
	}, diags
}
