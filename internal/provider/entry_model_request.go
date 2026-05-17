package provider

import (
	"context"
	"encoding/json"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	for fieldID, localizedValues := range attrs {
		if localizedValues.IsNull() || localizedValues.IsUnknown() {
			continue
		}

		fieldValue, fieldValueDiags := entryLocalizedFieldToRaw(path.Root("fields").AtMapKey(fieldID), localizedValues)
		diags.Append(fieldValueDiags...)

		if fieldValueDiags.HasError() {
			continue
		}

		fields[fieldID] = fieldValue
	}

	return cm.NewOptEntryFields(fields), diags
}

func entryLocalizedFieldToRaw(path path.Path, localizedValues TypedMap[jsontypes.Normalized]) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	values := map[string]json.RawMessage{}

	for locale, value := range localizedValues.Elements() {
		if value.IsNull() || value.IsUnknown() {
			continue
		}

		raw := []byte(value.ValueString())
		if !json.Valid(raw) {
			diags.AddAttributeError(path.AtMapKey(locale), "Invalid Entry Field Value", "Expected a valid JSON value.")

			continue
		}

		values[locale] = json.RawMessage(raw)
	}

	encoded, err := json.Marshal(values)
	if err != nil {
		diags.AddAttributeError(path, "Invalid Entry Field Value", err.Error())
	}

	return jx.Raw(encoded), diags
}

func entryModelToOptEntryMetadata(_ context.Context, model EntryModel) (cm.OptEntryMetadata, diag.Diagnostics) {
	if model.Metadata.IsNull() || model.Metadata.IsUnknown() {
		return cm.OptEntryMetadata{}, nil
	}

	diags := diag.Diagnostics{}

	metadata := cm.EntryMetadata{}

	modelConcepts := model.Metadata.Value().Concepts
	if !modelConcepts.IsNull() && !modelConcepts.IsUnknown() {
		concepts := make([]cm.TaxonomyConceptLink, 0, len(modelConcepts.Elements()))

		for _, concept := range modelConcepts.Elements() {
			conceptValue := concept.ValueString()
			concepts = append(concepts, cm.NewTaxonomyConceptLink(conceptValue))
		}

		metadata.Concepts = concepts
	}

	modelTags := model.Metadata.Value().Tags
	if !modelTags.IsNull() && !modelTags.IsUnknown() {
		tags := make([]cm.TagLink, 0, len(modelTags.Elements()))

		for _, tag := range modelTags.Elements() {
			tagValue := tag.ValueString()
			tags = append(tags, cm.NewTagLink(tagValue))
		}

		metadata.Tags = tags
	}

	return cm.NewOptEntryMetadata(metadata), diags
}
