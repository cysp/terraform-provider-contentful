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

		editorLayoutElementValues := []EditorLayoutValue{}
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

		controlsElementValues := []ControlsValue{}
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

		groupControlsElementValues := []GroupControlsValue{}
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

		sidebarElementValues := []SidebarValue{}
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

func (model *EditorInterfaceModel) ReadFromResponse(ctx context.Context, editorInterface *cm.EditorInterface) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// SpaceId, EnvironmentId and ContentTypeId are all already known

	if editorInterfaceEditorLayout, ok := editorInterface.EditorLayout.Get(); ok {
		editorLayout, editorLayoutDiags := NewEditorLayoutListValueFromResponse(ctx, path.Root("editor_layout"), editorInterfaceEditorLayout)
		diags.Append(editorLayoutDiags...)

		model.EditorLayout = editorLayout
	} else {
		model.EditorLayout = NewEditorLayoutListValueNull(ctx)
	}

	if editorInterfaceControls, ok := editorInterface.Controls.Get(); ok {
		controlsListValue, controlsListValueDiags := NewControlsListValueFromResponse(ctx, path.Root("controls"), editorInterfaceControls)
		diags.Append(controlsListValueDiags...)

		model.Controls = controlsListValue
	} else {
		model.Controls = NewControlsListValueNull(ctx)
	}

	if editorInterfaceGroupControls, ok := editorInterface.GroupControls.Get(); ok {
		groupControlsListValue, groupControlsListValueDiags := NewGroupControlsListValueFromResponse(ctx, path.Root("group_controls"), editorInterfaceGroupControls)
		diags.Append(groupControlsListValueDiags...)

		model.GroupControls = groupControlsListValue
	} else {
		model.GroupControls = NewGroupControlsListValueNull(ctx)
	}

	if editorInterfaceSidebar, ok := editorInterface.Sidebar.Get(); ok {
		sidebarListValue, sidebarListValueDiags := NewSidebarListValueFromResponse(ctx, path.Root("sidebar"), editorInterfaceSidebar)
		diags.Append(sidebarListValueDiags...)

		model.Sidebar = sidebarListValue
	} else {
		model.Sidebar = NewSidebarListValueNull(ctx)
	}

	return diags
}
