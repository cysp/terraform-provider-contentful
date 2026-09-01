package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*appInstallationResource)(nil)
	_ resource.ResourceWithConfigure   = (*appInstallationResource)(nil)
	_ resource.ResourceWithIdentity    = (*appInstallationResource)(nil)
	_ resource.ResourceWithImportState = (*appInstallationResource)(nil)
)

//nolint:ireturn
func NewAppInstallationResource() resource.Resource {
	return &appInstallationResource{}
}

type appInstallationResource struct {
	providerData ContentfulProviderData
}

func appInstallationIdentityAttributeNames() []string {
	return []string{"space_id", "environment_id", "app_definition_id"}
}

func (r *appInstallationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_installation"
}

func (r *appInstallationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AppInstallationResourceSchema(ctx)
}

func (r *appInstallationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *appInstallationResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(appInstallationIdentityAttributeNames())
}

func (r *appInstallationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, appInstallationIdentityAttributeNames(), req, resp)
}

//nolint:dupl
func (r *appInstallationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppInstallationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	xContentfulMarketplace, xContentfulMarketplaceDiags := plan.ToXContentfulMarketplaceHeaderValue()
	resp.Diagnostics.Append(xContentfulMarketplaceDiags...)

	params := cm.PutAppInstallationParams{
		SpaceID:                plan.SpaceID.ValueString(),
		EnvironmentID:          plan.EnvironmentID.ValueString(),
		AppDefinitionID:        plan.AppDefinitionID.ValueString(),
		XContentfulMarketplace: xContentfulMarketplace,
	}

	request, requestDiags := plan.ToAppInstallationData()
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppInstallation(ctx, &request, params)

	tflog.Info(ctx, "app_installation.create", map[string]any{
		"params": params,
		// "request": request, omitted to avoid logging sensitive values
		// "response": response, omitted to avoid logging sensitive values
		"err": err,
	})

	var data AppInstallationModel

	switch response := response.(type) {
	case *cm.AppInstallationStatusCode:
		responseModel, responseModelDiags := NewAppInstallationResourceModelFromResponse(response.Response, plan.Marketplace)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to create app installation", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appInstallationIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *appInstallationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppInstallationModel

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

	params := cm.GetAppInstallationParams{
		SpaceID:         state.SpaceID.ValueString(),
		EnvironmentID:   state.EnvironmentID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.GetAppInstallation(ctx, params)

	tflog.Info(ctx, "app_installation.read", map[string]any{
		"params": params,
		// "response": response, omitted to avoid logging sensitive values
		"err": err,
	})

	var data AppInstallationModel

	switch response := response.(type) {
	case *cm.AppInstallation:
		responseModel, responseModelDiags := NewAppInstallationResourceModelFromResponse(*response, state.Marketplace)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read app installation", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read app installation", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appInstallationIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:dupl
func (r *appInstallationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AppInstallationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	xContentfulMarketplace, xContentfulMarketplaceDiags := plan.ToXContentfulMarketplaceHeaderValue()
	resp.Diagnostics.Append(xContentfulMarketplaceDiags...)

	params := cm.PutAppInstallationParams{
		SpaceID:                plan.SpaceID.ValueString(),
		EnvironmentID:          plan.EnvironmentID.ValueString(),
		AppDefinitionID:        plan.AppDefinitionID.ValueString(),
		XContentfulMarketplace: xContentfulMarketplace,
	}

	request, requestDiags := plan.ToAppInstallationData()
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppInstallation(ctx, &request, params)

	tflog.Info(ctx, "app_installation.update", map[string]any{
		"params": params,
		// "request": request, omitted to avoid logging sensitive values
		// "response": response, omitted to avoid logging sensitive values
		"err": err,
	})

	var data AppInstallationModel

	switch response := response.(type) {
	case *cm.AppInstallationStatusCode:
		responseModel, responseModelDiags := NewAppInstallationResourceModelFromResponse(response.Response, plan.Marketplace)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to update app installation", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appInstallationIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:dupl
func (r *appInstallationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppInstallationModel

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

	params := cm.DeleteAppInstallationParams{
		SpaceID:         state.SpaceID.ValueString(),
		EnvironmentID:   state.EnvironmentID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.DeleteAppInstallation(ctx, params)

	tflog.Info(ctx, "app_installation.delete", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("App already uninstalled", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to uninstall app", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
