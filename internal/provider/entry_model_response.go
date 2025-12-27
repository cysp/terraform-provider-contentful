package provider

import (
	"context"

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

	return model, diags
}

func NewEntryFieldsFromResponse(_ context.Context, _ path.Path, fields cm.OptEntryFields) (TypedMap[jsontypes.Normalized], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if !fields.IsSet() {
		return NewTypedMapNull[jsontypes.Normalized](), diags
	}

	elements := map[string]jsontypes.Normalized{}
	for k, v := range fields.Value {
		elements[k] = NewNormalizedJSONTypesNormalizedValue(v)
	}

	return NewTypedMap(elements), diags
}

func NewEntryMetadataFromResponse(ctx context.Context, _ path.Path, metadata cm.OptEntryMetadata) (TypedObject[EntryMetadataValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if !metadata.IsSet() {
		return NewTypedObjectNull[EntryMetadataValue](), diags
	}

	concepts := []types.String{}

	for _, concept := range metadata.Value.Concepts {
		if concept.Sys.ID != "" {
			concepts = append(concepts, types.StringValue(concept.Sys.ID))
		}
	}

	tags := []types.String{}

	for _, tag := range metadata.Value.Tags {
		if tag.Sys.ID != "" {
			tags = append(tags, types.StringValue(tag.Sys.ID))
		}
	}

	obj, objDiags := NewTypedObjectFromAttributes[EntryMetadataValue](ctx, map[string]attr.Value{
		"concepts": NewTypedList(concepts),
		"tags":     NewTypedList(tags),
	})
	diags.Append(objDiags...)

	return obj, diags
}
