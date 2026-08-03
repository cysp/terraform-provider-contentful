package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (m EntryModel) ToEntryRequest(ctx context.Context) (cm.EntryRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields, fieldsDiags := entryModelToOptEntryFields(ctx, m)
	diags.Append(fieldsDiags...)

	metadata, metadataDiags := entryModelToOptEntryMetadata(ctx, m)
	diags.Append(metadataDiags...)

	if diags.HasError() {
		return cm.EntryRequest{}, diags
	}

	return cm.EntryRequest{
		Fields:   fields,
		Metadata: metadata,
	}, diags
}

func entryModelToOptEntryFields(_ context.Context, model EntryModel) (cm.OptEntryFields, diag.Diagnostics) {
	if model.Fields.IsUnknown() {
		diags := diag.Diagnostics{}
		diags.AddAttributeError(
			path.Root("fields"),
			"Unexpected unknown entry fields",
			"Entry fields must be known before they can be sent to Contentful.",
		)

		return cm.OptEntryFields{}, diags
	}

	if model.Fields.IsNull() {
		diags := diag.Diagnostics{}
		diags.AddAttributeError(
			path.Root("fields"),
			"Unexpected null entry fields",
			"Entry fields are required.",
		)

		return cm.OptEntryFields{}, diags
	}

	diags := diag.Diagnostics{}

	fields := make(cm.EntryFields)

	attrs := model.Fields.Elements()
	for k, v := range attrs {
		if v.IsNull() {
			// Terraform null omits the field. A configured JSON null remains a
			// known jsontypes.Normalized value and is sent as JSON null.
			continue
		}

		if v.IsUnknown() {
			diags.AddAttributeError(
				path.Root("fields").AtMapKey(k),
				"Unexpected unknown entry field",
				"Entry field values must be known before they can be sent to Contentful.",
			)

			continue
		}

		fields[k] = jx.Raw(v.ValueString())
	}

	if diags.HasError() {
		return cm.OptEntryFields{}, diags
	}

	return cm.NewOptEntryFields(fields), diags
}

func entryModelToOptEntryMetadata(_ context.Context, model EntryModel) (cm.OptEntryMetadata, diag.Diagnostics) {
	if model.Metadata.IsNull() || model.Metadata.IsUnknown() {
		return cm.OptEntryMetadata{}, nil
	}

	diags := diag.Diagnostics{}

	metadata := cm.EntryMetadata{}

	modelConcepts := model.Metadata.Value().Concepts
	if !modelConcepts.IsNull() && !modelConcepts.IsUnknown() {
		conceptValues, conceptDiags := knownStringListElements(
			path.Root("metadata").AtName("concepts"),
			modelConcepts.Elements(),
		)
		diags.Append(conceptDiags...)

		concepts := make([]cm.TaxonomyConceptLink, 0, len(conceptValues))
		for _, conceptValue := range conceptValues {
			concepts = append(concepts, cm.NewTaxonomyConceptLink(conceptValue))
		}

		metadata.Concepts = concepts
	}

	modelTags := model.Metadata.Value().Tags
	if !modelTags.IsNull() && !modelTags.IsUnknown() {
		tagValues, tagDiags := knownStringListElements(
			path.Root("metadata").AtName("tags"),
			modelTags.Elements(),
		)
		diags.Append(tagDiags...)

		tags := make([]cm.TagLink, 0, len(tagValues))
		for _, tagValue := range tagValues {
			tags = append(tags, cm.NewTagLink(tagValue))
		}

		metadata.Tags = tags
	}

	if diags.HasError() {
		return cm.OptEntryMetadata{}, diags
	}

	return cm.NewOptEntryMetadata(metadata), diags
}
