package provider

import (
	"context"
	"encoding/json"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// entryFieldsRequestProjection omits Terraform-null map elements. A known
// jsontypes.Normalized value containing JSON null remains a request value.
func entryFieldsRequestProjection(fields TypedMap[TypedMap[jsontypes.Normalized]]) TypedMap[TypedMap[jsontypes.Normalized]] {
	if fields.IsNull() || fields.IsUnknown() {
		return fields
	}

	elements := make(map[string]TypedMap[jsontypes.Normalized], len(fields.Elements()))
	for fieldID, value := range fields.Elements() {
		if !value.IsNull() {
			elements[fieldID] = value
		}
	}

	return NewTypedMap(elements)
}

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

	attrs := entryFieldsRequestProjection(model.Fields).Elements()
	for fieldID, localizedValues := range attrs {
		if localizedValues.IsUnknown() {
			diags.AddAttributeError(
				path.Root("fields").AtMapKey(fieldID),
				"Unexpected unknown entry field",
				"Entry field values must be known before they can be sent to Contentful.",
			)

			continue
		}

		if localizedValues.IsNull() {
			// Terraform null omits the field. A configured JSON null remains a
			// known jsontypes.Normalized value in a locale and is sent as JSON null.
			continue
		}

		fieldValue, fieldValueDiags := entryLocalizedFieldToRaw(path.Root("fields").AtMapKey(fieldID), localizedValues)
		diags.Append(fieldValueDiags...)

		if fieldValueDiags.HasError() {
			continue
		}

		fields[fieldID] = fieldValue
	}

	if diags.HasError() {
		return cm.OptEntryFields{}, diags
	}

	return cm.NewOptEntryFields(fields), diags
}

func entryLocalizedFieldToRaw(path path.Path, localizedValues TypedMap[jsontypes.Normalized]) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	values := map[string]json.RawMessage{}

	for locale, value := range localizedValues.Elements() {
		if value.IsNull() {
			continue
		}

		if value.IsUnknown() {
			diags.AddAttributeError(
				path.AtMapKey(locale),
				"Unexpected unknown entry field value",
				"Entry field values must be known before they can be sent to Contentful.",
			)

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

func isRawJSONNull(raw []byte) bool {
	return strings.TrimSpace(string(raw)) == "null"
}

func entryModelToOptEntryMetadata(_ context.Context, model EntryModel) (cm.OptEntryMetadata, diag.Diagnostics) {
	if model.Metadata.IsUnknown() {
		diags := diag.Diagnostics{}
		diags.AddAttributeError(
			path.Root("metadata"),
			"Unexpected unknown entry metadata",
			"Entry metadata must be known before it can be sent to Contentful.",
		)

		return cm.OptEntryMetadata{}, diags
	}

	if model.Metadata.IsNull() {
		return cm.OptEntryMetadata{}, nil
	}

	diags := diag.Diagnostics{}

	metadata := cm.EntryMetadata{}

	modelConcepts := model.Metadata.Value().Concepts
	if modelConcepts.IsUnknown() {
		diags.AddAttributeError(
			path.Root("metadata").AtName("concepts"),
			"Unexpected unknown entry concepts",
			"Entry concepts must be known before they can be sent to Contentful.",
		)
	} else if !modelConcepts.IsNull() {
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
	if modelTags.IsUnknown() {
		diags.AddAttributeError(
			path.Root("metadata").AtName("tags"),
			"Unexpected unknown entry tags",
			"Entry tags must be known before they can be sent to Contentful.",
		)
	} else if !modelTags.IsNull() {
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
