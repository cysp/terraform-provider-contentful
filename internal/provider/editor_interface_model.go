package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/resource_editor_interface"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ReadEditorInterfaceModel(ctx context.Context, model *resource_editor_interface.EditorInterfaceModel, editorInterface contentfulManagement.EditorInterface) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// SpaceId, EnvironmentId and ContentTypeId are all already known

	if editorInterfaceControls, ok := editorInterface.Controls.Get(); ok {
		model.Controls = resource_editor_interface.NewControlsListValueFromResponse(ctx, path.Root("controls"), editorInterfaceControls, &diags)
	} else {
		model.Controls = resource_editor_interface.NewControlsListValueNull(ctx)
	}

	if editorInterfaceSidebar, ok := editorInterface.Sidebar.Get(); ok {
		model.Sidebar = resource_editor_interface.NewSidebarListValueFromResponse(ctx, path.Root("sidebar"), editorInterfaceSidebar, &diags)
	} else {
		model.Sidebar = resource_editor_interface.NewSidebarListValueNull(ctx)
	}

	return diags
}
