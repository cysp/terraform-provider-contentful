package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *EditorInterfaceModel) ToPutEditorInterfaceReq(ctx context.Context) (contentfulManagement.PutEditorInterfaceReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	request := contentfulManagement.PutEditorInterfaceReq{}

	if model.Controls.IsNull() || model.Controls.IsUnknown() {
		request.Controls.Reset()
	} else {
		controlsPath := path.Root("controls")

		controlsElementValues := []ControlsValue{}
		diags.Append(model.Controls.ElementsAs(ctx, &controlsElementValues, false)...)

		requestControlsItems := make([]contentfulManagement.PutEditorInterfaceReqControlsItem, len(controlsElementValues))

		for index, controlsElement := range controlsElementValues {
			path := controlsPath.AtListIndex(index)

			requestControlsItem, requestControlsItemDiags := controlsElement.ToPutEditorInterfaceReqControlsItem(ctx, path)
			diags.Append(requestControlsItemDiags...)

			requestControlsItems[index] = requestControlsItem
		}

		request.Controls.SetTo(requestControlsItems)
	}

	if model.Sidebar.IsNull() || model.Sidebar.IsUnknown() {
		request.Sidebar.Reset()
	} else {
		sidebarPath := path.Root("sidebar")

		sidebarElementValues := []SidebarValue{}
		diags.Append(model.Sidebar.ElementsAs(ctx, &sidebarElementValues, false)...)

		requestSidebarItems := make([]contentfulManagement.PutEditorInterfaceReqSidebarItem, len(sidebarElementValues))

		for index, sidebarElement := range sidebarElementValues {
			path := sidebarPath.AtListIndex(index)

			requestSidebarItem, requestSidebarItemDiags := sidebarElement.ToPutEditorInterfaceReqSidebarItem(ctx, path)
			diags.Append(requestSidebarItemDiags...)

			requestSidebarItems[index] = requestSidebarItem
		}

		request.Sidebar.SetTo(requestSidebarItems)
	}

	return request, diags
}

func (model *EditorInterfaceModel) ReadFromResponse(ctx context.Context, editorInterface *contentfulManagement.EditorInterface) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// SpaceId, EnvironmentId and ContentTypeId are all already known

	if editorInterfaceControls, ok := editorInterface.Controls.Get(); ok {
		controlsListValue, controlsListValueDiags := NewControlsListValueFromResponse(ctx, path.Root("controls"), editorInterfaceControls)
		diags.Append(controlsListValueDiags...)

		model.Controls = controlsListValue
	} else {
		model.Controls = NewControlsListValueNull(ctx)
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
