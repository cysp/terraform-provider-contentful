package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ContentTypeDataSourceItemModel struct {
	ContentTypeID    types.String                                  `tfsdk:"content_type_id"`
	Name             types.String                                  `tfsdk:"name"`
	Description      types.String                                  `tfsdk:"description"`
	DisplayField     types.String                                  `tfsdk:"display_field"`
	PublishedVersion types.Int64                                   `tfsdk:"published_version"`
	Fields           TypedList[TypedObject[ContentTypeFieldValue]] `tfsdk:"fields"`
	Metadata         TypedObject[ContentTypeMetadataValue]         `tfsdk:"metadata"`
}

type ContentTypeDataSourceModel struct {
	IDIdentityModel
	ContentTypeDataSourceItemModel

	SpaceID       types.String   `tfsdk:"space_id"`
	EnvironmentID types.String   `tfsdk:"environment_id"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

type ContentTypesDataSourceModel struct {
	IDIdentityModel

	SpaceID       types.String                     `tfsdk:"space_id"`
	EnvironmentID types.String                     `tfsdk:"environment_id"`
	ContentTypes  []ContentTypeDataSourceItemModel `tfsdk:"content_types"`
	Timeouts      timeouts.Value                   `tfsdk:"timeouts"`
}
