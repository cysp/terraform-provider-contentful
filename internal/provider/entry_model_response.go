package provider

import (
	"context"
	"strings"

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
		IDIdentityModel: IDIdentityModel{
			ID: types.StringValue(strings.Join([]string{spaceID, environmentID, entryID}, "/")),
		},
		EntryIdentityModel: EntryIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
			EntryID:       types.StringValue(entryID),
		},
		ContentTypeID: types.StringValue(contentTypeID),
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
		elements[k] = jsontypes.NewNormalizedValue(string(v))
	}

	return NewTypedMap(elements), diags
}

func NewEntryMetadataFromResponse(ctx context.Context, _ path.Path, metadata cm.OptEntryMetadata) (TypedObject[EntryMetadataValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if !metadata.IsSet() {
		return NewTypedObjectNull[EntryMetadataValue](), diags
	}

	tags := []types.String{}

	for _, tag := range metadata.Value.Tags {
		if tag.Sys.ID != "" {
			tags = append(tags, types.StringValue(tag.Sys.ID))
		}
	}

	obj, objDiags := NewTypedObjectFromAttributes[EntryMetadataValue](ctx, map[string]attr.Value{
		"tags": NewTypedList(tags),
	})
	diags.Append(objDiags...)

	return obj, diags
}
