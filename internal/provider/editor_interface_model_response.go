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
// response and restores known effective Plan values only after proving that
// every API-backed value is losslessly, semantically equivalent. Endpoint
// identity always remains the requested Terraform target.
func ReconcileEditorInterfaceMutationResponse(ctx context.Context, editorInterface cm.EditorInterface, plan EditorInterfaceModel) (EditorInterfaceModel, diag.Diagnostics, diag.Diagnostics) {
	state, responseDiags, editorLayoutDiags := newEditorInterfaceResourceModelFromResponse(ctx, editorInterface)
	reconciler := mutationResponseReconciler{resourceName: "Editor Interface"}
	reconciler.pinIdentity(path.Root("space_id"), plan.SpaceID, state.SpaceID, &state.SpaceID)
	reconciler.pinIdentity(path.Root("environment_id"), plan.EnvironmentID, state.EnvironmentID, &state.EnvironmentID)
	reconciler.pinIdentity(path.Root("content_type_id"), plan.ContentTypeID, state.ContentTypeID, &state.ContentTypeID)

	state.IDIdentityModel = NewIDIdentityModelFromMultipartID(state.SpaceID.ValueString(), state.EnvironmentID.ValueString(), state.ContentTypeID.ValueString())

	if !plan.ID.IsNull() && !plan.ID.IsUnknown() && !plan.ID.Equal(state.ID) {
		reconciler.diagnostics.AddAttributeError(path.Root("id"), "Editor Interface identity is inconsistent with its endpoint", "The planned legacy ID differs from the requested Editor Interface endpoint identity. Terraform retained the endpoint identity as the resource target and the remaining returned values as recovery state. Review or re-import the Editor Interface before applying again.")
	}

	candidateState := state
	if reconciler.compareSemantic(
		path.Root("editor_layout"), "Editor Interface layout", "Contentful returned a different Editor Interface layout",
		plan.EditorLayout.IsUnknown(), editorLayoutDiags,
		func() (bool, diag.Diagnostics) { return plan.EditorLayout.Equal(state.EditorLayout), nil },
	) {
		candidateState.EditorLayout = plan.EditorLayout
	}

	if reconciler.compareSemantic(
		path.Root("controls"), "Editor Interface controls", "Contentful returned different Editor Interface controls",
		plan.Controls.IsUnknown(), diag.Diagnostics{},
		func() (bool, diag.Diagnostics) {
			return editorInterfaceControlsEquivalent(ctx, plan.Controls, state.Controls)
		},
	) {
		candidateState.Controls = plan.Controls
	}

	if reconciler.compareSemantic(
		path.Root("group_controls"), "Editor Interface group controls", "Contentful returned different Editor Interface group controls",
		plan.GroupControls.IsUnknown(), diag.Diagnostics{},
		func() (bool, diag.Diagnostics) {
			return editorInterfaceGroupControlsEquivalent(ctx, plan.GroupControls, state.GroupControls)
		},
	) {
		candidateState.GroupControls = plan.GroupControls
	}

	if reconciler.compareSemantic(
		path.Root("sidebar"), "Editor Interface sidebar", "Contentful returned a different Editor Interface sidebar",
		plan.Sidebar.IsUnknown(), diag.Diagnostics{},
		func() (bool, diag.Diagnostics) {
			return editorInterfaceSidebarEquivalent(ctx, plan.Sidebar, state.Sidebar)
		},
	) {
		candidateState.Sidebar = plan.Sidebar
	}

	if !reconciler.diagnostics.HasError() {
		state = candidateState
	}

	return state, responseDiags, reconciler.diagnostics
}
