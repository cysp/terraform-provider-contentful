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

	// Store fields as an opaque JSON blob
	fieldsObj := jsontypes.Wrap(entry.Fields)
	model.Fields = fieldsObj

	metadata, metadataDiags := NewEntryMetadataFromResponse(ctx, path.Root("metadata"), entry.Metadata)
	diags.Append(metadataDiags...)
	model.Metadata = metadata

	return model, diags
}

// convertToAttrValue converts supported types to attr.Value
func convertToAttrValue(v any) attr.Value {
	switch val := v.(type) {
	case string:
		return types.StringValue(val)
	case bool:
		return types.BoolValue(val)
	case int:
		return types.Int64Value(int64(val))
	case int64:
		return types.Int64Value(val)
	case float64:
		return types.Float64Value(val)
	default:
		return types.StringNull() // fallback for unsupported types
	}
}

func NewEntryMetadataFromResponse(ctx context.Context, path path.Path, metadata cm.OptEntryMetadata) (TypedObject[EntryMetadataValue], diag.Diagnostics) {
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

	attrs := map[string]attr.Value{"tags": NewTypedList(tags)}
	obj, objDiags := NewTypedObjectFromAttributes[EntryMetadataValue](ctx, attrs)
	diags.Append(objDiags...)
	return obj, diags
}
