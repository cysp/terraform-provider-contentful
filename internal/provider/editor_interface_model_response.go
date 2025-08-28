package provider

import (
	"context"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorInterfaceResourceModelFromResponse(ctx context.Context, editorInterface cm.EditorInterface) (EditorInterfaceModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := editorInterface.Sys.Space.Sys.ID
	environmentID := editorInterface.Sys.Environment.Sys.ID
	contentTypeID := editorInterface.Sys.ContentType.Sys.ID

	model := EditorInterfaceModel{
		ID:            types.StringValue(strings.Join([]string{spaceID, environmentID, contentTypeID}, "/")),
		SpaceID:       types.StringValue(spaceID),
		EnvironmentID: types.StringValue(environmentID),
		ContentTypeID: types.StringValue(contentTypeID),
	}

	if editorInterfaceEditorLayout, ok := editorInterface.EditorLayout.Get(); ok {
		editorLayout, editorLayoutDiags := NewEditorInterfaceEditorLayoutListValueFromResponse(ctx, path.Root("editor_layout"), editorInterfaceEditorLayout)
		diags.Append(editorLayoutDiags...)

		model.EditorLayout = editorLayout
	} else {
		model.EditorLayout = NewEditorInterfaceEditorLayoutListValueNull()
	}

	if editorInterfaceControls, ok := editorInterface.Controls.Get(); ok {
		controlsListValue, controlsListValueDiags := NewEditorInterfaceControlListValueFromResponse(ctx, path.Root("controls"), editorInterfaceControls)
		diags.Append(controlsListValueDiags...)

		model.Controls = controlsListValue
	} else {
		model.Controls = NewEditorInterfaceControlListValueNull()
	}

	if editorInterfaceGroupControls, ok := editorInterface.GroupControls.Get(); ok {
		groupControlsListValue, groupControlsListValueDiags := NewEditorInterfaceGroupControlListValueFromResponse(ctx, path.Root("group_controls"), editorInterfaceGroupControls)
		diags.Append(groupControlsListValueDiags...)

		model.GroupControls = groupControlsListValue
	} else {
		model.GroupControls = NewEditorInterfaceGroupControlListValueNull()
	}

	if editorInterfaceSidebar, ok := editorInterface.Sidebar.Get(); ok {
		sidebarListValue, sidebarListValueDiags := NewEditorInterfaceSidebarListValueFromResponse(ctx, path.Root("sidebar"), editorInterfaceSidebar)
		diags.Append(sidebarListValueDiags...)

		model.Sidebar = sidebarListValue
	} else {
		model.Sidebar = NewEditorInterfaceSidebarListValueNull()
	}

	return model, diags
}
