package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EditorInterfaceIdentityModel struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ContentTypeID types.String `tfsdk:"content_type_id"`
}

type EditorInterfaceModel struct {
	EditorInterfaceIdentityModel

	ID            types.String                                                 `tfsdk:"id"`
	EditorLayout  TypedList[TypedObject[EditorInterfaceEditorLayoutItemValue]] `tfsdk:"editor_layout"`
	Controls      TypedList[TypedObject[EditorInterfaceControlValue]]          `tfsdk:"controls"`
	GroupControls TypedList[TypedObject[EditorInterfaceGroupControlValue]]     `tfsdk:"group_controls"`
	Sidebar       TypedList[TypedObject[EditorInterfaceSidebarValue]]          `tfsdk:"sidebar"`
}

type EditorInterfaceEditorLayoutItemValue struct {
	Group TypedObject[EditorInterfaceEditorLayoutItemGroupValue] `tfsdk:"group"`
}

type EditorInterfaceEditorLayoutItemGroupValue struct {
	GroupID types.String                                                          `tfsdk:"group_id"`
	Name    types.String                                                          `tfsdk:"name"`
	Items   TypedList[TypedObject[EditorInterfaceEditorLayoutItemGroupItemValue]] `tfsdk:"items"`
}

type EditorInterfaceEditorLayoutItemGroupItemValue struct {
	Field TypedObject[EditorInterfaceEditorLayoutItemGroupItemFieldValue] `tfsdk:"field"`
	Group TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupValue] `tfsdk:"group"`
}

type EditorInterfaceEditorLayoutItemGroupItemFieldValue struct {
	FieldID types.String `tfsdk:"field_id"`
}

type EditorInterfaceEditorLayoutItemGroupItemGroupValue struct {
	GroupID types.String                                                                   `tfsdk:"group_id"`
	Name    types.String                                                                   `tfsdk:"name"`
	Items   TypedList[TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]] `tfsdk:"items"`
}

type EditorInterfaceEditorLayoutItemGroupItemGroupItemValue struct {
	Field TypedObject[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue] `tfsdk:"field"`
}

type EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue struct {
	FieldID types.String `tfsdk:"field_id"`
}

type EditorInterfaceControlValue struct {
	FieldID         types.String         `tfsdk:"field_id"`
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	WidgetID        types.String         `tfsdk:"widget_id"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
}

type EditorInterfaceGroupControlValue struct {
	GroupID         types.String         `tfsdk:"group_id"`
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	WidgetID        types.String         `tfsdk:"widget_id"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
}

type EditorInterfaceSidebarValue struct {
	WidgetNamespace types.String         `tfsdk:"widget_namespace"`
	WidgetID        types.String         `tfsdk:"widget_id"`
	Settings        jsontypes.Normalized `tfsdk:"settings"`
	Disabled        types.Bool           `tfsdk:"disabled"`
}
