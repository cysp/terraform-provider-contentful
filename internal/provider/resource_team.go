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
	_ resource.Resource                = (*teamResource)(nil)
	_ resource.ResourceWithConfigure   = (*teamResource)(nil)
	_ resource.ResourceWithIdentity    = (*teamResource)(nil)
	_ resource.ResourceWithImportState = (*teamResource)(nil)
)

//nolint:ireturn
func NewTeamResource() resource.Resource {
	return &teamResource{}
}

type teamResource struct {
	providerData ContentfulProviderData
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = TeamResourceSchema(ctx)
}

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *teamResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"organization_id": identityschema.StringAttribute{RequiredForImport: true},
			"team_id":         identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("organization_id"),
		path.Root("team_id"),
	}, req, resp)
}

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TeamModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.CreateTeamParams{
		OrganizationID: data.OrganizationID.ValueString(),
	}

	request, requestDiags := data.ToTeamData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreateTeam(ctx, &request, params)

	tflog.Info(ctx, "team.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var currentVersion int

	switch response := response.(type) {
	case *cm.TeamStatusCode:
		responseModel, responseModelDiags := NewTeamResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create team", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel TeamIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

//nolint:dupl
func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TeamModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetTeamParams{
		OrganizationID: data.OrganizationID.ValueString(),
		TeamID:         data.TeamID.ValueString(),
	}

	response, err := r.providerData.client.GetTeam(ctx, params)

	tflog.Info(ctx, "team.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	var currentVersion int

	switch response := response.(type) {
	case *cm.Team:
		responseModel, responseModelDiags := NewTeamResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Sys.Version

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read team", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read team", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel TeamIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

//nolint:dupl
func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TeamModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int

	currentVersionDiags := GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	params := cm.PutTeamParams{
		OrganizationID:     data.OrganizationID.ValueString(),
		TeamID:             data.TeamID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := data.ToTeamData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutTeam(ctx, &request, params)

	tflog.Info(ctx, "team.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.TeamStatusCode:
		responseModel, responseModelDiags := NewTeamResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update team", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel TeamIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

//nolint:dupl
func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TeamModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.DeleteTeam(ctx, cm.DeleteTeamParams{
		OrganizationID: data.OrganizationID.ValueString(),
		TeamID:         data.TeamID.ValueString(),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Team already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete team", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
