package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEditorInterfaceResourceModelFromResponse(ctx context.Context, editorInterface cm.EditorInterface) (EditorInterfaceModel, diag.Diagnostics) {
	model, diags, _ := newEditorInterfaceResourceModelFromResponse(ctx, editorInterface)

	return model, diags
}

func newEditorInterfaceResourceModelFromResponse(ctx context.Context, editorInterface cm.EditorInterface) (EditorInterfaceModel, diag.Diagnostics, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	editorLayoutDiags := diag.Diagnostics{}

	spaceID := editorInterface.Sys.Space.Sys.ID
	environmentID := editorInterface.Sys.Environment.Sys.ID
	contentTypeID := editorInterface.Sys.ContentType.Sys.ID

	model := EditorInterfaceModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID, contentTypeID),
		EditorInterfaceIdentityModel: EditorInterfaceIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
			ContentTypeID: types.StringValue(contentTypeID),
		},
	}

	if editorInterfaceEditorLayout, ok := editorInterface.EditorLayout.Get(); ok {
		editorLayout, projectedEditorLayoutDiags := NewEditorInterfaceEditorLayoutListValueFromResponse(ctx, path.Root("editor_layout"), editorInterfaceEditorLayout)
		diags.Append(projectedEditorLayoutDiags...)
		editorLayoutDiags.Append(projectedEditorLayoutDiags...)

		model.EditorLayout = editorLayout
	} else {
		model.EditorLayout = NewTypedListNull[TypedObject[EditorInterfaceEditorLayoutItemValue]]()
	}

	if editorInterfaceControls, ok := editorInterface.Controls.Get(); ok {
		controlsListValue, controlsListValueDiags := NewEditorInterfaceControlListValueFromResponse(ctx, path.Root("controls"), editorInterfaceControls)
		diags.Append(controlsListValueDiags...)

		model.Controls = controlsListValue
	} else {
		model.Controls = NewTypedListNull[TypedObject[EditorInterfaceControlValue]]()
	}

	if editorInterfaceGroupControls, ok := editorInterface.GroupControls.Get(); ok {
		groupControlsListValue, groupControlsListValueDiags := NewEditorInterfaceGroupControlListValueFromResponse(ctx, path.Root("group_controls"), editorInterfaceGroupControls)
		diags.Append(groupControlsListValueDiags...)

		model.GroupControls = groupControlsListValue
	} else {
		model.GroupControls = NewTypedListNull[TypedObject[EditorInterfaceGroupControlValue]]()
	}

	if editorInterfaceSidebar, ok := editorInterface.Sidebar.Get(); ok {
		sidebarListValue, sidebarListValueDiags := NewEditorInterfaceSidebarListValueFromResponse(ctx, path.Root("sidebar"), editorInterfaceSidebar)
		diags.Append(sidebarListValueDiags...)

		model.Sidebar = sidebarListValue
	} else {
		model.Sidebar = NewTypedListNull[TypedObject[EditorInterfaceSidebarValue]]()
	}

	return model, diags, editorLayoutDiags
}

// NewEditorInterfaceResourceModelForMutationState starts with the complete
// response projection. It restores the exact known Plan representation only
// after ordered editor_layout equality is proven; lossy or contradictory
// responses remain available as recovery state with a consistency diagnostic.
// Unknown Plan values resolve from the response. Read skips reconciliation.
func NewEditorInterfaceResourceModelForMutationState(ctx context.Context, editorInterface cm.EditorInterface, appliedPlan EditorInterfaceModel) (EditorInterfaceModel, diag.Diagnostics, diag.Diagnostics) {
	mutationState, responseDiags, editorLayoutDiags := newEditorInterfaceResourceModelFromResponse(ctx, editorInterface)
	if appliedPlan.EditorLayout.IsUnknown() {
		return mutationState, responseDiags, nil
	}

	consistencyDiags := diag.Diagnostics{}

	switch {
	case len(editorLayoutDiags) != 0:
		consistencyDiags.AddAttributeError(path.Root("editor_layout"), "Unexpected Contentful editor interface response", "The editor_layout response could not be projected without loss, so equivalence with the Terraform plan could not be established.")
	case appliedPlan.EditorLayout.Equal(mutationState.EditorLayout):
		mutationState.EditorLayout = appliedPlan.EditorLayout
	default:
		consistencyDiags.AddAttributeError(path.Root("editor_layout"), "Unexpected Contentful editor interface response", "The editor_layout response differed meaningfully from the Terraform plan.")
	}

	return mutationState, responseDiags, consistencyDiags
}
