package provider

import (
	"context"
	"encoding/json"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewLivePreviewVariablesResourceModelFromResponse(response cm.LivePreviewVariables) (LivePreviewVariablesModel, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	spaceID := response.Sys.Space.Sys.ID
	environmentID := response.Sys.Environment.Sys.ID
	model := LivePreviewVariablesModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID),
		LivePreviewVariablesIdentityModel: LivePreviewVariablesIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
		},
	}

	if !json.Valid(response.Variables) {
		diagnostics.AddAttributeError(path.Root("variables"), "Invalid live preview variables response", "Contentful returned missing or invalid JSON variables; Terraform cannot publish this response.")
	} else {
		model.Variables = NewNormalizedJSONValue(response.Variables)
	}

	return model, diagnostics
}

func ReconcileLivePreviewVariablesMutationResponse(ctx context.Context, response cm.LivePreviewVariables, plan LivePreviewVariablesModel) (LivePreviewVariablesModel, diag.Diagnostics, diag.Diagnostics) {
	model, diagnostics := NewLivePreviewVariablesResourceModelFromResponse(response)
	reconciler := mutationResponseReconciler{resourceName: "live preview variables"}
	reconciler.pinIdentity(path.Root("space_id"), plan.SpaceID, model.SpaceID, &model.SpaceID)
	reconciler.pinIdentity(path.Root("environment_id"), plan.EnvironmentID, model.EnvironmentID, &model.EnvironmentID)

	model.IDIdentityModel = NewIDIdentityModelFromMultipartID(model.SpaceID.ValueString(), model.EnvironmentID.ValueString())

	if !plan.ID.IsNull() && !plan.ID.IsUnknown() && !plan.ID.Equal(model.ID) {
		reconciler.diagnostics.AddAttributeError(path.Root("id"), "Live preview variables identity is inconsistent with its endpoint", "The planned legacy ID differs from the requested live preview variables endpoint identity. Terraform retained the endpoint identity as the resource target and the remaining returned values as recovery state. Review or re-import the live preview variables before applying again.")
	}

	candidateModel := model
	if !diagnostics.HasError() && reconciler.compareSemantic(
		path.Root("variables"), "live preview variables", "Contentful returned different live preview variables",
		plan.Variables.IsUnknown(), nil,
		func() (bool, diag.Diagnostics) { return normalizedJSONEquivalent(ctx, plan.Variables, model.Variables) },
	) {
		candidateModel.Variables = plan.Variables
	}

	if !reconciler.diagnostics.HasError() {
		model = candidateModel
	}

	return model, diagnostics, reconciler.diagnostics
}
