package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = (*appActionResource)(nil)
	_ resource.ResourceWithConfigure      = (*appActionResource)(nil)
	_ resource.ResourceWithIdentity       = (*appActionResource)(nil)
	_ resource.ResourceWithImportState    = (*appActionResource)(nil)
	_ resource.ResourceWithValidateConfig = (*appActionResource)(nil)
)

//nolint:ireturn
func NewAppActionResource() resource.Resource { return &appActionResource{} }

type appActionResource struct{ providerData ContentfulProviderData }

func appActionIdentityAttributeNames() []string {
	return []string{"organization_id", "app_definition_id", "app_action_id"}
}

func (r *appActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_action"
}

func (r *appActionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AppActionResourceSchema(ctx)
}

func (r *appActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *appActionResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(appActionIdentityAttributeNames())
}

func (r *appActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, appActionIdentityAttributeNames(), req, resp)
}

func (r *appActionResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config AppActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateAppActionTarget(config.Type, config.URL, config.FunctionID)...)

	resp.Diagnostics.Append(config.validateInput()...)
}

func (r *appActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan         AppActionModel
		configuredID types.String
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("app_action_id"), &configuredID)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(appActionScope(plan.AppActionBaseModel)...)

	if configuredID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("app_action_id"), "Unknown App Action ID", "The configured App Action ID must be known before creating the action.")
	}

	body, diags := plan.ToAppActionData()
	resp.Diagnostics.Append(diags...)

	createBody := appActionCreateData(body)

	if !configuredID.IsNull() {
		resp.Diagnostics.Append(requireDiscoveryID(path.Root("app_action_id"), plan.AppActionID)...)

		if !resp.Diagnostics.HasError() {
			createBody.ID = cm.NewOptString(plan.AppActionID.ValueString())
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diags := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	response, err := r.providerData.client.CreateAppAction(withContentfulRequestNoRedirect(ctx), &createBody, cm.CreateAppActionParams{OrganizationID: plan.OrganizationID.ValueString(), AppDefinitionID: plan.AppDefinitionID.ValueString()})

	result, ok := response.(*cm.AppAction)
	if err != nil || !ok || result == nil {
		resp.Diagnostics.AddError("Failed to create app action", util.ErrorDetailFromContentfulManagementResponse(response, err)+" Before retrying, check this App Definition in Contentful. Import the action if the request created one.")

		return
	}

	if !validDiscoveryID(result.Sys.ID) {
		resp.Diagnostics.AddAttributeError(path.Root("app_action_id"), "Invalid App Action response identity", "Contentful returned an App Action ID that Terraform cannot use. Check this App Definition in Contentful before attempting to create the action again.")

		return
	}

	data, consistencyDiags := reconcileAppActionResponse(ctx, *result, plan, AppActionBaseModel{OrganizationID: plan.OrganizationID, AppDefinitionID: plan.AppDefinitionID})
	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appActionIdentityAttributeNames(), &data)...)
	resp.Diagnostics.Append(consistencyDiags...)
}

func (r *appActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(appActionScope(state.AppActionBaseModel)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("app_action_id"), state.AppActionID)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diags := resourceReadContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	response, err := r.providerData.client.GetAppAction(ctx, cm.GetAppActionParams{OrganizationID: state.OrganizationID.ValueString(), AppDefinitionID: state.AppDefinitionID.ValueString(), AppActionID: state.AppActionID.ValueString()})
	if err == nil && contentfulResponseIsNotFound(response) {
		resp.State.RemoveResource(ctx)

		return
	}

	result, ok := response.(*cm.AppAction)
	if err != nil || !ok || result == nil {
		resp.Diagnostics.AddError("Failed to read app action", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	base, identityDiags := appActionResourceResponse(*result, state.AppActionBaseModel)
	state.AppActionBaseModel = base
	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appActionIdentityAttributeNames(), &state)...)
	resp.Diagnostics.Append(identityDiags...)
}

func (r *appActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state AppActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if appActionRemoteValuesEqual(ctx, plan, state) {
		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appActionIdentityAttributeNames(), &state)...)

		return
	}

	resp.Diagnostics.Append(appActionScope(plan.AppActionBaseModel)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("app_action_id"), plan.AppActionID)...)
	body, diags := plan.ToAppActionData()
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diags := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	response, err := r.providerData.client.UpdateAppAction(withContentfulRequestNoRedirect(ctx), &body, cm.UpdateAppActionParams{OrganizationID: plan.OrganizationID.ValueString(), AppDefinitionID: plan.AppDefinitionID.ValueString(), AppActionID: plan.AppActionID.ValueString()})

	result, ok := response.(*cm.AppAction)
	if err != nil || !ok || result == nil {
		resp.Diagnostics.AddError("Failed to update app action", util.ErrorDetailFromContentfulManagementResponse(response, err)+" The action may have been updated even though the request failed. Refresh its state and review the plan before applying again.")

		return
	}

	data, consistencyDiags := reconcileAppActionResponse(ctx, *result, plan, plan.AppActionBaseModel)
	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appActionIdentityAttributeNames(), &data)...)
	resp.Diagnostics.Append(consistencyDiags...)
}

func (r *appActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(appActionScope(state.AppActionBaseModel)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("app_action_id"), state.AppActionID)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diags := resourceDeleteContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	response, err := r.providerData.client.DeleteAppAction(withContentfulRequestNoRedirect(ctx), cm.DeleteAppActionParams{OrganizationID: state.OrganizationID.ValueString(), AppDefinitionID: state.AppDefinitionID.ValueString(), AppActionID: state.AppActionID.ValueString()})
	if err == nil {
		if _, ok := response.(*cm.NoContent); ok {
			return
		}

		if contentfulResponseIsNotFound(response) {
			return
		}
	}

	resp.Diagnostics.AddError("Failed to delete app action", util.ErrorDetailFromContentfulManagementResponse(response, err))
}

func appActionRemoteValuesEqual(ctx context.Context, plan, state AppActionModel) bool {
	for _, field := range []struct {
		plan  *jsontypes.Normalized
		state jsontypes.Normalized
	}{
		{&plan.Parameters, state.Parameters}, {&plan.ParametersSchema, state.ParametersSchema}, {&plan.ResultSchema, state.ResultSchema},
	} {
		equal, diags := normalizedJSONEquivalent(ctx, *field.plan, field.state)
		if !equal || diags.HasError() {
			return false
		}

		*field.plan = field.state
	}

	return NewTypedObject(plan.AppActionBaseModel).Equal(NewTypedObject(state.AppActionBaseModel))
}
