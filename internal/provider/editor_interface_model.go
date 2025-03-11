package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EditorInterfaceModel struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ContentTypeID types.String `tfsdk:"content_type_id"`
	EditorLayout  types.List   `tfsdk:"editor_layout"`
	Controls      types.List   `tfsdk:"controls"`
	GroupControls types.List   `tfsdk:"group_controls"`
	Sidebar       types.List   `tfsdk:"sidebar"`
}
