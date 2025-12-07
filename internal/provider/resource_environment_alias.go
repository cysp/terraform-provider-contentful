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
	_ resource.Resource                = (*environmentAliasResource)(nil)
	_ resource.ResourceWithConfigure   = (*environmentAliasResource)(nil)
	_ resource.ResourceWithIdentity    = (*environmentAliasResource)(nil)
	_ resource.ResourceWithImportState = (*environmentAliasResource)(nil)
)

//nolint:ireturn
func NewEnvironmentAliasResource() resource.Resource {
	return &environmentAliasResource{}
}

type environmentAliasResource struct {
	providerData ContentfulProviderData
}

func (r *environmentAliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_alias"
}

func (r *environmentAliasResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = EnvironmentAliasResourceSchema(ctx)
}

func (r *environmentAliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *environmentAliasResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"space_id":             identityschema.StringAttribute{RequiredForImport: true},
			"environment_alias_id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *environmentAliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_alias_id"),
	}, req, resp)
}

func (r *environmentAliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentAliasModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1

	params := cm.CreateOrUpdateEnvironmentAliasParams{
		SpaceID:            plan.SpaceID.ValueString(),
		EnvironmentAliasID: plan.EnvironmentAliasID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := plan.ToEnvironmentAliasData(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreateOrUpdateEnvironmentAlias(ctx, &request, params)

	tflog.Info(ctx, "environment_alias.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data EnvironmentAliasModel

	switch response := response.(type) {
	case *cm.EnvironmentAliasStatusCode:
		responseModel, responseModelDiags := NewEnvironmentAliasResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel EnvironmentAliasIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

//nolint:dupl
func (r *environmentAliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentAliasModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetEnvironmentAliasParams{
		SpaceID:            state.SpaceID.ValueString(),
		EnvironmentAliasID: state.EnvironmentAliasID.ValueString(),
	}

	response, err := r.providerData.client.GetEnvironmentAlias(ctx, params)

	tflog.Info(ctx, "environment_alias.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0

	var data EnvironmentAliasModel

	switch response := response.(type) {
	case *cm.EnvironmentAlias:
		responseModel, responseModelDiags := NewEnvironmentAliasResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Sys.Version

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel EnvironmentAliasIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

func (r *environmentAliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentAliasModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int

	currentVersionDiags := GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	params := cm.CreateOrUpdateEnvironmentAliasParams{
		SpaceID:            plan.SpaceID.ValueString(),
		EnvironmentAliasID: plan.EnvironmentAliasID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := plan.ToEnvironmentAliasData(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreateOrUpdateEnvironmentAlias(ctx, &request, params)

	tflog.Info(ctx, "environment_alias.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data EnvironmentAliasModel

	switch response := response.(type) {
	case *cm.EnvironmentAliasStatusCode:
		responseModel, responseModelDiags := NewEnvironmentAliasResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel EnvironmentAliasIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

//nolint:dupl
func (r *environmentAliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentAliasModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.DeleteEnvironmentAliasParams{
		SpaceID:            state.SpaceID.ValueString(),
		EnvironmentAliasID: state.EnvironmentAliasID.ValueString(),
	}

	response, err := r.providerData.client.DeleteEnvironmentAlias(ctx, params)

	tflog.Info(ctx, "environment_alias.delete", map[string]any{
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
				resp.Diagnostics.AddWarning("Environment alias already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
