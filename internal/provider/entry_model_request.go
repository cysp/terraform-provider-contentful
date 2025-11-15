package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func (m EntryModel) ToEntryRequest(ctx context.Context) (cm.EntryRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields, fieldsDiags := entryModelToOptEntryFields(ctx, m)
	diags.Append(fieldsDiags...)

	metadata, metadataDiags := entryModelToOptEntryMetadata(ctx, m)
	diags.Append(metadataDiags...)

	return cm.EntryRequest{
		Fields:   fields,
		Metadata: metadata,
	}, diags
}

func entryModelToOptEntryFields(_ context.Context, model EntryModel) (cm.OptEntryFields, diag.Diagnostics) {
	if model.Fields.IsNull() || model.Fields.IsUnknown() {
		return cm.OptEntryFields{}, nil
	}

	diags := diag.Diagnostics{}

	fields := make(cm.EntryFields)

	attrs := model.Fields.Elements()
	for k, v := range attrs {
		if v.IsNull() {
			continue
		}

		fields[k] = jx.Raw(v.ValueString())
	}

	return cm.NewOptEntryFields(fields), diags
}

func entryModelToOptEntryMetadata(_ context.Context, model EntryModel) (cm.OptEntryMetadata, diag.Diagnostics) {
	if model.Metadata.IsNull() || model.Metadata.IsUnknown() {
		return cm.OptEntryMetadata{}, nil
	}

	diags := diag.Diagnostics{}

	metadata := cm.EntryMetadata{}

	modelTags := model.Metadata.Value().Tags.Elements()
	tags := make([]cm.TagLink, 0, len(modelTags))

	for _, tag := range modelTags {
		tagValue := tag.ValueString()
		if tagValue == "" {
			continue
		}

		tags = append(tags, cm.NewTagLink(tagValue))
	}

	metadata.Tags = tags

	return cm.NewOptEntryMetadata(metadata), diags
}
