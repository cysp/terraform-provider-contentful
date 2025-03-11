package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *EditorInterfaceModel) ReadFromResponse(ctx context.Context, editorInterface *cm.EditorInterface) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// SpaceId, EnvironmentId and ContentTypeId are all already known

	if editorInterfaceEditorLayout, ok := editorInterface.EditorLayout.Get(); ok {
		editorLayout, editorLayoutDiags := NewEditorInterfaceEditorLayoutListValueFromResponse(ctx, path.Root("editor_layout"), editorInterfaceEditorLayout)
		diags.Append(editorLayoutDiags...)

		model.EditorLayout = editorLayout
	} else {
		model.EditorLayout = NewEditorInterfaceEditorLayoutListValueNull(ctx)
	}

	if editorInterfaceControls, ok := editorInterface.Controls.Get(); ok {
		controlsListValue, controlsListValueDiags := NewEditorInterfaceControlListValueFromResponse(ctx, path.Root("controls"), editorInterfaceControls)
		diags.Append(controlsListValueDiags...)

		model.Controls = controlsListValue
	} else {
		model.Controls = NewEditorInterfaceControlListValueNull(ctx)
	}

	if editorInterfaceGroupControls, ok := editorInterface.GroupControls.Get(); ok {
		groupControlsListValue, groupControlsListValueDiags := NewEditorInterfaceGroupControlListValueFromResponse(ctx, path.Root("group_controls"), editorInterfaceGroupControls)
		diags.Append(groupControlsListValueDiags...)

		model.GroupControls = groupControlsListValue
	} else {
		model.GroupControls = NewEditorInterfaceGroupControlListValueNull(ctx)
	}

	if editorInterfaceSidebar, ok := editorInterface.Sidebar.Get(); ok {
		sidebarListValue, sidebarListValueDiags := NewEditorInterfaceSidebarListValueFromResponse(ctx, path.Root("sidebar"), editorInterfaceSidebar)
		diags.Append(sidebarListValueDiags...)

		model.Sidebar = sidebarListValue
	} else {
		model.Sidebar = NewEditorInterfaceSidebarListValueNull(ctx)
	}

	return diags
}
