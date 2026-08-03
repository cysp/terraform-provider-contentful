package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *EditorInterfaceModel) ToEditorInterfaceData(ctx context.Context) (cm.EditorInterfaceData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	request := cm.EditorInterfaceData{}

	editorLayout, editorLayoutSet, editorLayoutDiags := editorInterfaceOptionalObjectList(
		ctx,
		path.Root("editor_layout"),
		model.EditorLayout,
		func(ctx context.Context, valuePath path.Path, value EditorInterfaceEditorLayoutItemValue) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics) {
			return value.ToEditorInterfaceEditorLayoutItem(ctx, valuePath)
		},
	)
	diags.Append(editorLayoutDiags...)

	if editorLayoutSet && !editorLayoutDiags.HasError() {
		request.EditorLayout.SetTo(editorLayout)
	}

	controls, controlsSet, controlsDiags := editorInterfaceOptionalObjectList(
		ctx,
		path.Root("controls"),
		model.Controls,
		func(ctx context.Context, valuePath path.Path, value EditorInterfaceControlValue) (cm.EditorInterfaceDataControlsItem, diag.Diagnostics) {
			return value.ToEditorInterfaceDataControlsItem(ctx, valuePath)
		},
	)
	diags.Append(controlsDiags...)

	if controlsSet && !controlsDiags.HasError() {
		request.Controls.SetTo(controls)
	}

	groupControls, groupControlsSet, groupControlsDiags := editorInterfaceOptionalObjectList(
		ctx,
		path.Root("group_controls"),
		model.GroupControls,
		func(ctx context.Context, valuePath path.Path, value EditorInterfaceGroupControlValue) (cm.EditorInterfaceDataGroupControlsItem, diag.Diagnostics) {
			return value.ToEditorInterfaceDataGroupControlsItem(ctx, valuePath)
		},
	)
	diags.Append(groupControlsDiags...)

	if groupControlsSet && !groupControlsDiags.HasError() {
		request.GroupControls.SetTo(groupControls)
	}

	sidebar, sidebarSet, sidebarDiags := editorInterfaceOptionalObjectList(
		ctx,
		path.Root("sidebar"),
		model.Sidebar,
		func(ctx context.Context, valuePath path.Path, value EditorInterfaceSidebarValue) (cm.EditorInterfaceDataSidebarItem, diag.Diagnostics) {
			return value.ToEditorInterfaceDataSidebarItem(ctx, valuePath)
		},
	)
	diags.Append(sidebarDiags...)

	if sidebarSet && !sidebarDiags.HasError() {
		request.Sidebar.SetTo(sidebar)
	}

	if diags.HasError() {
		return cm.EditorInterfaceData{}, diags
	}

	return request, diags
}
