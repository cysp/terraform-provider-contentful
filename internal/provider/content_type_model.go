package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ContentTypeModel struct {
	ContentTypeId types.String `tfsdk:"content_type_id"`
	Description   types.String `tfsdk:"description"`
	DisplayField  types.String `tfsdk:"display_field"`
	EnvironmentId types.String `tfsdk:"environment_id"`
	Fields        types.List   `tfsdk:"fields"`
	Name          types.String `tfsdk:"name"`
	SpaceId       types.String `tfsdk:"space_id"`
}
