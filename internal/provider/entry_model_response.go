package provider

import (
	"context"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEntryResourceModelFromResponse(ctx context.Context, entry cm.Entry) (EntryModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := entry.Sys.Space.Sys.ID
	environmentID := entry.Sys.Environment.Sys.ID
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
		ContentTypeID: types.StringValue(entry.Sys.ContentType.Sys.ID),
	}

	fields, fieldsDiags := jsontypes.NewNormalizedValue(string(entry.Fields))
	diags.Append(fieldsDiags...)
	model.Fields = DiagsMust(NewTypedObjectFromValue[jsontypes.Normalized](ctx, fields))

	metadata, metadataDiags := NewEntryMetadataFromResponse(ctx, path.Root("metadata"), entry.Metadata)
	diags.Append(metadataDiags...)
	model.Metadata = metadata

	return model, diags
}

func NewEntryMetadataFromResponse(ctx context.Context, path path.Path, metadata cm.OptEntryMetadata) (TypedObject[EntryMetadataValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if !metadata.IsSet() {
		return NewTypedObjectNull[EntryMetadataValue](), diags
	}

	tags := []types.String{}
	for _, tag := range metadata.Value.Tags {
		tags = append(tags, types.StringValue(tag.Sys.ID))
	}

	return DiagsMust(NewTypedObjectFromAttributes[EntryMetadataValue](ctx, map[string]types.Value{
		"tags": NewTypedList(tags),
	})), diags
}
