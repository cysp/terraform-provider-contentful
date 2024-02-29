package provider

import (
	"context"
	"net/http"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/resource_app_installation"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	resp.Schema = resource_app_installation.AppInstallationResourceSchema(ctx)
}

func (r *appInstallationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	util.ProviderDataFromResourceConfigureRequest(req, &r.providerData, resp)
}

func (r *appInstallationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	util.ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("app_definition_id"),
	}, req, resp)
}

//nolint:dupl
func (r *appInstallationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_app_installation.AppInstallationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := contentfulManagement.PutAppInstallationReq{}
	resp.Diagnostics.Append(CreatePutAppInstallationRequestBody(&request, data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppInstallation(ctx, &request, contentfulManagement.PutAppInstallationParams{
		SpaceID:         data.SpaceId.ValueString(),
		EnvironmentID:   data.EnvironmentId.ValueString(),
		AppDefinitionID: data.AppDefinitionId.ValueString(),
	})

	switch response := response.(type) {
	case *contentfulManagement.AppInstallation:
		ReadAppInstallationModel(&data, *response)
	default:
		resp.Diagnostics.AddError("Failed to create app installation", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *appInstallationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_app_installation.AppInstallationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.GetAppInstallation(ctx, contentfulManagement.GetAppInstallationParams{
		SpaceID:         data.SpaceId.ValueString(),
		EnvironmentID:   data.EnvironmentId.ValueString(),
		AppDefinitionID: data.AppDefinitionId.ValueString(),
	})

	switch response := response.(type) {
	case *contentfulManagement.AppInstallation:
		ReadAppInstallationModel(&data, *response)

	case *contentfulManagement.ErrorStatusCode:
		if response.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddWarning("Failed to read app installation", response.Response.Message)
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read app installation", response.Response.Message)

	default:
		resp.Diagnostics.AddError("Failed to read app installation", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appInstallationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_app_installation.AppInstallationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := contentfulManagement.PutAppInstallationReq{}
	resp.Diagnostics.Append(CreatePutAppInstallationRequestBody(&request, data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppInstallation(ctx, &request, contentfulManagement.PutAppInstallationParams{
		SpaceID:         data.SpaceId.ValueString(),
		EnvironmentID:   data.EnvironmentId.ValueString(),
		AppDefinitionID: data.AppDefinitionId.ValueString(),
	})

	switch response := response.(type) {
	case *contentfulManagement.AppInstallation:
		ReadAppInstallationModel(&data, *response)
	default:
		resp.Diagnostics.AddError("Failed to update app installation", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *appInstallationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_app_installation.AppInstallationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.DeleteAppInstallation(ctx, contentfulManagement.DeleteAppInstallationParams{
		SpaceID:         data.SpaceId.ValueString(),
		EnvironmentID:   data.EnvironmentId.ValueString(),
		AppDefinitionID: data.AppDefinitionId.ValueString(),
	})

	switch response := response.(type) {
	case *contentfulManagement.NoContent:

	case *contentfulManagement.ErrorStatusCode:
		if response.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddWarning("App already uninstalled", response.Response.Message)
		} else {
			resp.Diagnostics.AddError("Failed to uninstall app", response.Response.Message)
		}

	default:
		resp.Diagnostics.AddError("Failed to uninstall app", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}
}