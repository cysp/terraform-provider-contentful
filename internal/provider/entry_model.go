package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EntryIdentityModel struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	EntryID       types.String `tfsdk:"entry_id"`
}

func NewEntryIdentityModel(spaceID, environmentID, entryID string) EntryIdentityModel {
	return EntryIdentityModel{
		SpaceID:       types.StringValue(spaceID),
		EnvironmentID: types.StringValue(environmentID),
		EntryID:       types.StringValue(entryID),
	}
}

type EntryModel struct {
	IDIdentityModel
	EntryIdentityModel

	ContentTypeID types.String `tfsdk:"content_type_id"`

	Fields   TypedMap[jsontypes.Normalized]  `tfsdk:"fields"`
	Metadata TypedObject[EntryMetadataValue] `tfsdk:"metadata"`
}

type EntryMetadataValue struct {
	Concepts TypedList[types.String] `tfsdk:"concepts"`
	Tags     TypedList[types.String] `tfsdk:"tags"`
}
