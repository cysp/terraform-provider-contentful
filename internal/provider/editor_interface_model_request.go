package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *EditorInterfaceModel) ToEditorInterfaceFields(ctx context.Context) (cm.EditorInterfaceFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	request := cm.EditorInterfaceFields{}

	if model.EditorLayout.IsNull() || model.EditorLayout.IsUnknown() {
		request.EditorLayout.Reset()
	} else {
		editorLayoutPath := path.Root("editor_layout")

		editorLayoutElementValues := []EditorInterfaceEditorLayoutValue{}
		diags.Append(model.EditorLayout.ElementsAs(ctx, &editorLayoutElementValues, false)...)

		requestEditorLayoutItems := make([]cm.EditorInterfaceFieldsEditorLayoutItem, len(editorLayoutElementValues))

		for index, editorLayoutElement := range editorLayoutElementValues {
			path := editorLayoutPath.AtListIndex(index)

			requestEditorLayoutItem, requestEditorLayoutItemDiags := editorLayoutElement.ToEditorInterfaceFieldsEditorLayoutItem(ctx, path)
			diags.Append(requestEditorLayoutItemDiags...)

			requestEditorLayoutItems[index] = requestEditorLayoutItem
		}

		request.EditorLayout.SetTo(requestEditorLayoutItems)
	}

	if model.Controls.IsNull() || model.Controls.IsUnknown() {
		request.Controls.Reset()
	} else {
		controlsPath := path.Root("controls")

		controlsElementValues := []EditorInterfaceControlValue{}
		diags.Append(model.Controls.ElementsAs(ctx, &controlsElementValues, false)...)

		requestControlsItems := make([]cm.EditorInterfaceFieldsControlsItem, len(controlsElementValues))

		for index, controlsElement := range controlsElementValues {
			path := controlsPath.AtListIndex(index)

			requestControlsItem, requestControlsItemDiags := controlsElement.ToEditorInterfaceFieldsControlsItem(ctx, path)
			diags.Append(requestControlsItemDiags...)

			requestControlsItems[index] = requestControlsItem
		}

		request.Controls.SetTo(requestControlsItems)
	}

	if model.GroupControls.IsNull() || model.GroupControls.IsUnknown() {
		request.GroupControls.Reset()
	} else {
		controlsPath := path.Root("group_controls")

		groupControlsElementValues := []EditorInterfaceGroupControlValue{}
		diags.Append(model.GroupControls.ElementsAs(ctx, &groupControlsElementValues, false)...)

		requestGroupControlsItems := make([]cm.EditorInterfaceFieldsGroupControlsItem, len(groupControlsElementValues))

		for index, groupControlsElement := range groupControlsElementValues {
			path := controlsPath.AtListIndex(index)

			requestGroupControlsItem, requestGroupControlsItemDiags := groupControlsElement.ToEditorInterfaceFieldsGroupControlsItem(ctx, path)
			diags.Append(requestGroupControlsItemDiags...)

			requestGroupControlsItems[index] = requestGroupControlsItem
		}

		request.GroupControls.SetTo(requestGroupControlsItems)
	}

	if model.Sidebar.IsNull() || model.Sidebar.IsUnknown() {
		request.Sidebar.Reset()
	} else {
		sidebarPath := path.Root("sidebar")

		sidebarElementValues := []EditorInterfaceSidebarValue{}
		diags.Append(model.Sidebar.ElementsAs(ctx, &sidebarElementValues, false)...)

		requestSidebarItems := make([]cm.EditorInterfaceFieldsSidebarItem, len(sidebarElementValues))

		for index, sidebarElement := range sidebarElementValues {
			path := sidebarPath.AtListIndex(index)

			requestSidebarItem, requestSidebarItemDiags := sidebarElement.ToEditorInterfaceFieldsSidebarItem(ctx, path)
			diags.Append(requestSidebarItemDiags...)

			requestSidebarItems[index] = requestSidebarItem
		}

		request.Sidebar.SetTo(requestSidebarItems)
	}

	return request, diags
}
