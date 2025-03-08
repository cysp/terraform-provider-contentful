package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ContentTypeModel struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ContentTypeID types.String `tfsdk:"content_type_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	DisplayField  types.String `tfsdk:"display_field"`
	Fields        types.List   `tfsdk:"fields"`
}
