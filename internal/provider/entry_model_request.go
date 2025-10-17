package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func (m *EntryModel) ToEntryRequestFields(ctx context.Context) (cm.EntryFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var fields map[string]any
	diags.Append(tfsdk.ValueAs(ctx, m.Fields.Value(), &fields)...)
	if diags.HasError() {
		return cm.EntryFields{}, diags
	}

	metadata, metadataDiags := ToOptEntryMetadata(ctx, path.Root("metadata"), m.Metadata)
	diags.Append(metadataDiags...)

	request := cm.EntryFields{
		Fields:   fields,
		Metadata: metadata,
	}

	return request, diags
}

func ToOptEntryMetadata(ctx context.Context, path path.Path, metadataObject TypedObject[EntryMetadataValue]) (cm.OptEntryMetadata, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	optMetadata := cm.OptEntryMetadata{}

	metadataValue, ok := metadataObject.GetValue()
	if !ok {
		return optMetadata, diags
	}

	tags := []cm.EntryMetadataTagsItem{}
	if !metadataValue.Tags.IsNull() && !metadataValue.Tags.IsUnknown() {
		tagValues := metadataValue.Tags.Elements()
		for _, tagValue := range tagValues {
			tags = append(tags, cm.EntryMetadataTagsItem{
				Sys: cm.EntryMetadataTagsItemSys{
					Type:     "Link",
					LinkType: "Tag",
					ID:       tagValue.ValueString(),
				},
			})
		}
	}

	optMetadata.SetTo(cm.EntryMetadata{
		Tags: tags,
	})

	return optMetadata, diags
}
