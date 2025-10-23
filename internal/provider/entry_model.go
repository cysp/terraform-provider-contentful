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

type EntryModel struct {
	IDIdentityModel
	EntryIdentityModel

	ContentTypeID types.String `tfsdk:"content_type_id"`

	Fields   TypedMap[jsontypes.Normalized]  `tfsdk:"fields"`
	Metadata TypedObject[EntryMetadataValue] `tfsdk:"metadata"`
}

type EntryMetadataValue struct {
	Tags TypedList[types.String] `tfsdk:"tags"`
}
