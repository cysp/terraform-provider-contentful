package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*appDefinitionAppDefinitionResourceProviderResource)(nil)
	_ resource.ResourceWithConfigure   = (*appDefinitionAppDefinitionResourceProviderResource)(nil)
	_ resource.ResourceWithIdentity    = (*appDefinitionAppDefinitionResourceProviderResource)(nil)
	_ resource.ResourceWithImportState = (*appDefinitionAppDefinitionResourceProviderResource)(nil)
)

//nolint:ireturn
func NewAppDefinitionResourceProviderResource() resource.Resource {
	return &appDefinitionAppDefinitionResourceProviderResource{}
}

type appDefinitionAppDefinitionResourceProviderResource struct {
	providerData ContentfulProviderData
}

func (r *appDefinitionAppDefinitionResourceProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_definition_resource_provider"
}

func (r *appDefinitionAppDefinitionResourceProviderResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AppDefinitionResourceProviderResourceSchema(ctx)
}

func (r *appDefinitionAppDefinitionResourceProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *appDefinitionAppDefinitionResourceProviderResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"organization_id":   identityschema.StringAttribute{RequiredForImport: true},
			"app_definition_id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *appDefinitionAppDefinitionResourceProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("organization_id"),
		path.Root("app_definition_id"),
	}, req, resp)
}

//nolint:dupl
func (r *appDefinitionAppDefinitionResourceProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AppDefinitionResourceProviderModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutAppDefinitionResourceProviderParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	}

	request, requestDiags := data.ToResourceProviderRequest(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppDefinitionResourceProvider(ctx, &request, params)

	tflog.Info(ctx, "app_definition_resource_provider.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.ResourceProviderStatusCode:
		responseModel, responseModelDiags := NewAppDefinitionResourceProviderResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to create resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appDefinitionAppDefinitionResourceProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AppDefinitionResourceProviderModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetAppDefinitionResourceProviderParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.GetAppDefinitionResourceProvider(ctx, params)

	tflog.Info(ctx, "app_definition_resource_provider.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.ResourceProvider:
		responseModel, responseModelDiags := NewAppDefinitionResourceProviderResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appDefinitionAppDefinitionResourceProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AppDefinitionResourceProviderModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutAppDefinitionResourceProviderParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	}

	request, requestDiags := data.ToResourceProviderRequest(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppDefinitionResourceProvider(ctx, &request, params)

	tflog.Info(ctx, "app_definition_resource_provider.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.ResourceProviderStatusCode:
		responseModel, responseModelDiags := NewAppDefinitionResourceProviderResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to update resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appDefinitionAppDefinitionResourceProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AppDefinitionResourceProviderModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.DeleteAppDefinitionResourceProvider(ctx, cm.DeleteAppDefinitionResourceProviderParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Resource provider definition already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete resource provider definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
