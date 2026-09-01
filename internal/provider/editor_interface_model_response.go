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

// ReconcileEditorInterfaceMutationResponse projects the complete mutation
// response and verifies ordered, lossless editor_layout equivalence.
func ReconcileEditorInterfaceMutationResponse(ctx context.Context, editorInterface cm.EditorInterface, plan EditorInterfaceModel) (EditorInterfaceModel, diag.Diagnostics, diag.Diagnostics) {
	state, responseDiags, editorLayoutDiags := newEditorInterfaceResourceModelFromResponse(ctx, editorInterface)
	if plan.EditorLayout.IsUnknown() {
		return state, responseDiags, nil
	}

	consistencyDiags := diag.Diagnostics{}

	switch {
	case len(editorLayoutDiags) != 0:
		consistencyDiags.AddAttributeError(path.Root("editor_layout"), "Provider cannot fully represent Editor Interface layout", "Contentful accepted the request, but the returned Editor Interface layout contains values this provider cannot fully represent. Terraform retained the representable response values but cannot verify that they match the value Terraform applied. Review the Editor Interface in Contentful before applying again.")
	case !plan.EditorLayout.Equal(state.EditorLayout):
		consistencyDiags.AddAttributeError(path.Root("editor_layout"), "Contentful returned a different Editor Interface layout", "Contentful accepted the request but returned an Editor Interface layout that differs from the value Terraform applied. Terraform retained the returned value in state rather than substituting the planned value. Review the Editor Interface and configuration before applying again.")
	}

	return state, responseDiags, consistencyDiags
}
