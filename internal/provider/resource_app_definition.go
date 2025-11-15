//nolint:dupl
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
	_ resource.Resource                = (*appDefinitionAppDefinitionResource)(nil)
	_ resource.ResourceWithConfigure   = (*appDefinitionAppDefinitionResource)(nil)
	_ resource.ResourceWithIdentity    = (*appDefinitionAppDefinitionResource)(nil)
	_ resource.ResourceWithImportState = (*appDefinitionAppDefinitionResource)(nil)
)

//nolint:ireturn
func NewAppDefinitionResource() resource.Resource {
	return &appDefinitionAppDefinitionResource{}
}

type appDefinitionAppDefinitionResource struct {
	providerData ContentfulProviderData
}

func (r *appDefinitionAppDefinitionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_definition"
}

func (r *appDefinitionAppDefinitionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AppDefinitionResourceSchema(ctx)
}

func (r *appDefinitionAppDefinitionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *appDefinitionAppDefinitionResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"organization_id":   identityschema.StringAttribute{RequiredForImport: true},
			"app_definition_id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *appDefinitionAppDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("organization_id"),
		path.Root("app_definition_id"),
	}, req, resp)
}

func (r *appDefinitionAppDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AppDefinitionModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.CreateAppDefinitionParams{
		OrganizationID: data.OrganizationID.ValueString(),
	}

	request, requestDiags := data.ToAppDefinitionData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreateAppDefinition(ctx, &request, params)

	tflog.Info(ctx, "app_definition.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.AppDefinitionStatusCode:
		responseModel, responseModelDiags := NewAppDefinitionResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to create app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel AppDefinitionIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appDefinitionAppDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AppDefinitionModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetAppDefinitionParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.GetAppDefinition(ctx, params)

	tflog.Info(ctx, "app_definition.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.AppDefinition:
		responseModel, responseModelDiags := NewAppDefinitionResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel AppDefinitionIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appDefinitionAppDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AppDefinitionModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutAppDefinitionParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	}

	request, requestDiags := data.ToAppDefinitionData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppDefinition(ctx, &request, params)

	tflog.Info(ctx, "app_definition.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.AppDefinitionStatusCode:
		responseModel, responseModelDiags := NewAppDefinitionResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to update app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel AppDefinitionIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *appDefinitionAppDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AppDefinitionModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.DeleteAppDefinition(ctx, cm.DeleteAppDefinitionParams{
		OrganizationID:  data.OrganizationID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Resource type definition already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
