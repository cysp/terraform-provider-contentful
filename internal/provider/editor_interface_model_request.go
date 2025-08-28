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

		editorLayoutElementValues := model.EditorLayout.Elements()

		requestEditorLayoutItems := make([]cm.EditorInterfaceEditorLayoutItem, len(editorLayoutElementValues))

		for index, editorLayoutElement := range editorLayoutElementValues {
			path := editorLayoutPath.AtListIndex(index)

			requestEditorLayoutItem, requestEditorLayoutItemDiags := editorLayoutElement.Value().ToEditorInterfaceEditorLayoutItem(ctx, path)
			diags.Append(requestEditorLayoutItemDiags...)

			requestEditorLayoutItems[index] = requestEditorLayoutItem
		}

		request.EditorLayout.SetTo(requestEditorLayoutItems)
	}

	if model.Controls.IsNull() || model.Controls.IsUnknown() {
		request.Controls.Reset()
	} else {
		controlsPath := path.Root("controls")

		controlsElementValues := model.Controls.Elements()

		requestControlsItems := make([]cm.EditorInterfaceFieldsControlsItem, len(controlsElementValues))

		for index, controlsElement := range controlsElementValues {
			path := controlsPath.AtListIndex(index)

			requestControlsItem, requestControlsItemDiags := controlsElement.Value().ToEditorInterfaceFieldsControlsItem(ctx, path)
			diags.Append(requestControlsItemDiags...)

			requestControlsItems[index] = requestControlsItem
		}

		request.Controls.SetTo(requestControlsItems)
	}

	if model.GroupControls.IsNull() || model.GroupControls.IsUnknown() {
		request.GroupControls.Reset()
	} else {
		controlsPath := path.Root("group_controls")

		groupControlsElementValues := model.GroupControls.Elements()

		requestGroupControlsItems := make([]cm.EditorInterfaceFieldsGroupControlsItem, len(groupControlsElementValues))

		for index, groupControlsElement := range groupControlsElementValues {
			path := controlsPath.AtListIndex(index)

			requestGroupControlsItem, requestGroupControlsItemDiags := groupControlsElement.Value().ToEditorInterfaceFieldsGroupControlsItem(ctx, path)
			diags.Append(requestGroupControlsItemDiags...)

			requestGroupControlsItems[index] = requestGroupControlsItem
		}

		request.GroupControls.SetTo(requestGroupControlsItems)
	}

	if model.Sidebar.IsNull() || model.Sidebar.IsUnknown() {
		request.Sidebar.Reset()
	} else {
		sidebarPath := path.Root("sidebar")

		sidebarElementValues := model.Sidebar.Elements()

		requestSidebarItems := make([]cm.EditorInterfaceFieldsSidebarItem, len(sidebarElementValues))

		for index, sidebarElement := range sidebarElementValues {
			path := sidebarPath.AtListIndex(index)

			requestSidebarItem, requestSidebarItemDiags := sidebarElement.Value().ToEditorInterfaceFieldsSidebarItem(ctx, path)
			diags.Append(requestSidebarItemDiags...)

			requestSidebarItems[index] = requestSidebarItem
		}

		request.Sidebar.SetTo(requestSidebarItems)
	}

	return request, diags
}
