package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EditorInterfaceModel struct {
	ID            types.String                                    `tfsdk:"id"`
	SpaceID       types.String                                    `tfsdk:"space_id"`
	EnvironmentID types.String                                    `tfsdk:"environment_id"`
	ContentTypeID types.String                                    `tfsdk:"content_type_id"`
	EditorLayout  TypedList[EditorInterfaceEditorLayoutItemValue] `tfsdk:"editor_layout"`
	Controls      TypedList[EditorInterfaceControlValue]          `tfsdk:"controls"`
	GroupControls TypedList[EditorInterfaceGroupControlValue]     `tfsdk:"group_controls"`
	Sidebar       TypedList[EditorInterfaceSidebarValue]          `tfsdk:"sidebar"`
}
