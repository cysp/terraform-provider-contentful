package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*appInstallationResource)(nil)
	_ resource.ResourceWithConfigure   = (*appInstallationResource)(nil)
	_ resource.ResourceWithImportState = (*appInstallationResource)(nil)
)

//nolint:ireturn
func NewAppInstallationResource() resource.Resource {
	return &appInstallationResource{}
}

type appInstallationResource struct {
	providerData ContentfulProviderData
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

func (r *appInstallationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("app_definition_id"),
	}, req, resp)
}

//nolint:dupl
func (r *appInstallationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AppInstallationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	xContentfulMarketplace, xContentfulMarketplaceDiags := data.ToXContentfulMarketplaceHeaderValue(ctx)
	resp.Diagnostics.Append(xContentfulMarketplaceDiags...)

	params := cm.PutAppInstallationParams{
		SpaceID:                data.SpaceID.ValueString(),
		EnvironmentID:          data.EnvironmentID.ValueString(),
		AppDefinitionID:        data.AppDefinitionID.ValueString(),
		XContentfulMarketplace: xContentfulMarketplace,
	}

	request, requestDiags := data.ToAppInstallationFields()
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppInstallation(ctx, &request, params)

	tflog.Info(ctx, "app_installation.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.AppInstallationStatusCode:
		responseModel, responseModelDiags := NewAppInstallationResourceModelFromResponse(response.Response, data.Marketplace)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to create app installation", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *appInstallationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AppInstallationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetAppInstallationParams{
		SpaceID:         data.SpaceID.ValueString(),
		EnvironmentID:   data.EnvironmentID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.GetAppInstallation(ctx, params)

	tflog.Info(ctx, "app_installation.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.AppInstallation:
		responseModel, responseModelDiags := NewAppInstallationResourceModelFromResponse(*response, data.Marketplace)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appInstallationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AppInstallationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	xContentfulMarketplace, xContentfulMarketplaceDiags := data.ToXContentfulMarketplaceHeaderValue(ctx)
	resp.Diagnostics.Append(xContentfulMarketplaceDiags...)

	params := cm.PutAppInstallationParams{
		SpaceID:                data.SpaceID.ValueString(),
		EnvironmentID:          data.EnvironmentID.ValueString(),
		AppDefinitionID:        data.AppDefinitionID.ValueString(),
		XContentfulMarketplace: xContentfulMarketplace,
	}

	request, requestDiags := data.ToAppInstallationFields()
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppInstallation(ctx, &request, params)

	tflog.Info(ctx, "app_installation.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.AppInstallationStatusCode:
		responseModel, responseModelDiags := NewAppInstallationResourceModelFromResponse(response.Response, data.Marketplace)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to update app installation", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *appInstallationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AppInstallationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.DeleteAppInstallationParams{
		SpaceID:         data.SpaceID.ValueString(),
		EnvironmentID:   data.EnvironmentID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.DeleteAppInstallation(ctx, params)

	tflog.Info(ctx, "app_installation.delete", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("App already uninstalled", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to uninstall app", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
