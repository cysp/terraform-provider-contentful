//nolint:dupl
package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*appDefinitionResource)(nil)
	_ resource.ResourceWithConfigure   = (*appDefinitionResource)(nil)
	_ resource.ResourceWithIdentity    = (*appDefinitionResource)(nil)
	_ resource.ResourceWithImportState = (*appDefinitionResource)(nil)
)

//nolint:ireturn
func NewAppDefinitionResource() resource.Resource {
	return &appDefinitionResource{}
}

type appDefinitionResource struct {
	providerData ContentfulProviderData
}

func appDefinitionIdentityAttributeNames() []string {
	return []string{"organization_id", "app_definition_id"}
}

func (r *appDefinitionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_definition"
}

func (r *appDefinitionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AppDefinitionResourceSchema(ctx)
}

func (r *appDefinitionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *appDefinitionResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(appDefinitionIdentityAttributeNames())
}

func (r *appDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, appDefinitionIdentityAttributeNames(), req, resp)
}

func (r *appDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan   AppDefinitionResourceModel
		config AppDefinitionResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToAppDefinitionData(config.AppDefinitionBaseModel, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	params := cm.CreateAppDefinitionParams{
		OrganizationID: plan.OrganizationID.ValueString(),
	}

	response, err := r.providerData.client.CreateAppDefinition(ctx, &request, params)

	tflog.Info(ctx, "app_definition.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data AppDefinitionResourceModel

	switch response := response.(type) {
	case *cm.AppDefinitionStatusCode:
		responseModel, responseModelDiags := NewAppDefinitionResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to create app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appDefinitionIdentityAttributeNames(), &data)...)
}

//nolint:dupl
func (r *appDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppDefinitionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceReadContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	params := cm.GetAppDefinitionParams{
		OrganizationID:  state.OrganizationID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.GetAppDefinition(ctx, params)

	tflog.Info(ctx, "app_definition.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	var data AppDefinitionResourceModel

	switch response := response.(type) {
	case *cm.AppDefinition:
		responseModel, responseModelDiags := NewAppDefinitionResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appDefinitionIdentityAttributeNames(), &data)...)
}

//nolint:dupl
func (r *appDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan   AppDefinitionResourceModel
		config AppDefinitionResourceModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToAppDefinitionData(config.AppDefinitionBaseModel, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	params := cm.PutAppDefinitionParams{
		OrganizationID:  plan.OrganizationID.ValueString(),
		AppDefinitionID: plan.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.PutAppDefinition(ctx, &request, params)

	tflog.Info(ctx, "app_definition.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data AppDefinitionResourceModel

	switch response := response.(type) {
	case *cm.AppDefinitionStatusCode:
		responseModel, responseModelDiags := NewAppDefinitionResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to update app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appDefinitionIdentityAttributeNames(), &data)...)
}

//nolint:dupl
func (r *appDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppDefinitionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceDeleteContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	response, err := r.providerData.client.DeleteAppDefinition(ctx, cm.DeleteAppDefinitionParams{
		OrganizationID:  state.OrganizationID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Resource type definition already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

			return
		}

		resp.Diagnostics.AddError("Failed to delete app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}
}
