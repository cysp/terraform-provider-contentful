package provider

import (
	"context"
	"encoding/json"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEntryResourceModelFromResponse(ctx context.Context, entry cm.Entry) (EntryModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := entry.Sys.Space.Sys.ID
	environmentID := entry.Sys.Environment.Sys.ID
	contentTypeID := entry.Sys.ContentType.Sys.ID
	entryID := entry.Sys.ID

	model := EntryModel{
		IDIdentityModel:    NewIDIdentityModelFromMultipartID(spaceID, environmentID, entryID),
		EntryIdentityModel: NewEntryIdentityModel(spaceID, environmentID, entryID),
		ContentTypeID:      types.StringValue(contentTypeID),
	}

	fields, fieldsDiags := NewEntryFieldsFromResponse(ctx, path.Root("fields"), entry.Fields)
	diags.Append(fieldsDiags...)

	model.Fields = fields

	metadata, metadataDiags := NewEntryMetadataFromResponse(ctx, path.Root("metadata"), entry.Metadata)
	diags.Append(metadataDiags...)

	model.Metadata = metadata
	model.Timeouts = TimeoutsNull()

	return model, diags
}

func NewEntryFieldsFromResponse(_ context.Context, path path.Path, fields cm.OptEntryFields) (TypedMap[TypedMap[jsontypes.Normalized]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if !fields.IsSet() {
		return NewTypedMapNull[TypedMap[jsontypes.Normalized]](), diags
	}

	elements := map[string]TypedMap[jsontypes.Normalized]{}

	for fieldID, fieldValue := range fields.Value {
		localizedValues, localizedValuesDiags := NewEntryLocalizedFieldFromRaw(path.AtMapKey(fieldID), fieldValue)
		diags.Append(localizedValuesDiags...)

		if localizedValuesDiags.HasError() {
			continue
		}

		elements[fieldID] = localizedValues
	}

	return NewTypedMap(elements), diags
}

func NewEntryLocalizedFieldFromRaw(path path.Path, raw []byte) (TypedMap[jsontypes.Normalized], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if isRawJSONNull(raw) {
		return NewTypedMapNull[jsontypes.Normalized](), diags
	}

	var localizedValues map[string]json.RawMessage

	err := json.Unmarshal(raw, &localizedValues)
	if err != nil {
		diags.AddAttributeError(path, "Invalid Entry Field Value", fmt.Sprintf("Expected a JSON object keyed by locale: %s", err))

		return NewTypedMapNull[jsontypes.Normalized](), diags
	}

	elements := make(map[string]jsontypes.Normalized, len(localizedValues))
	for locale, value := range localizedValues {
		elements[locale] = NewNormalizedJSONTypesNormalizedValue(value)
	}

	return NewTypedMap(elements), diags
}

func NewEntryMetadataFromResponse(ctx context.Context, _ path.Path, metadata cm.OptEntryMetadata) (TypedObject[EntryMetadataValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if !metadata.IsSet() {
		return NewTypedObjectNull[EntryMetadataValue](), diags
	}

	conceptsValue := NewTypedListNull[types.String]()

	if metadata.Value.Concepts != nil {
		concepts := []types.String{}

		for _, concept := range metadata.Value.Concepts {
			concepts = append(concepts, types.StringValue(concept.Sys.ID))
		}

		conceptsValue = NewTypedList(concepts)
	}

	tagsValue := NewTypedListNull[types.String]()

	if metadata.Value.Tags != nil {
		tags := []types.String{}

		for _, tag := range metadata.Value.Tags {
			tags = append(tags, types.StringValue(tag.Sys.ID))
		}

		tagsValue = NewTypedList(tags)
	}

	obj, objDiags := NewTypedObjectFromAttributes[EntryMetadataValue](ctx, map[string]attr.Value{
		"concepts": conceptsValue,
		"tags":     tagsValue,
	})
	diags.Append(objDiags...)

	return obj, diags
}
