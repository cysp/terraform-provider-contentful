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
	_ resource.Resource                = (*teamSpaceMembershipResource)(nil)
	_ resource.ResourceWithConfigure   = (*teamSpaceMembershipResource)(nil)
	_ resource.ResourceWithIdentity    = (*teamSpaceMembershipResource)(nil)
	_ resource.ResourceWithImportState = (*teamSpaceMembershipResource)(nil)
)

//nolint:ireturn
func NewTeamSpaceMembershipResource() resource.Resource {
	return &teamSpaceMembershipResource{}
}

type teamSpaceMembershipResource struct {
	providerData ContentfulProviderData
}

func (r *teamSpaceMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_space_membership"
}

func (r *teamSpaceMembershipResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = TeamSpaceMembershipResourceSchema(ctx)
}

func (r *teamSpaceMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *teamSpaceMembershipResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"space_id":                 identityschema.StringAttribute{RequiredForImport: true},
			"team_space_membership_id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *teamSpaceMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("team_space_membership_id"),
	}, req, resp)
}

//nolint:dupl
func (r *teamSpaceMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TeamSpaceMembershipModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.CreateTeamSpaceMembershipParams{
		SpaceID:         plan.SpaceID.ValueString(),
		XContentfulTeam: plan.TeamID.ValueString(),
	}

	request, requestDiags := plan.ToTeamSpaceMembershipData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreateTeamSpaceMembership(ctx, &request, params)

	tflog.Info(ctx, "team_space_membership.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data TeamSpaceMembershipModel

	switch response := response.(type) {
	case *cm.TeamSpaceMembershipStatusCode:
		responseModel, responseModelDiags := NewTeamSpaceMembershipResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to create team space membership", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel TeamSpaceMembershipIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *teamSpaceMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TeamSpaceMembershipModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetTeamSpaceMembershipParams{
		SpaceID:               state.SpaceID.ValueString(),
		TeamSpaceMembershipID: state.TeamSpaceMembershipID.ValueString(),
	}

	response, err := r.providerData.client.GetTeamSpaceMembership(ctx, params)

	tflog.Info(ctx, "team_space_membership.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	var data TeamSpaceMembershipModel

	switch response := response.(type) {
	case *cm.TeamSpaceMembership:
		responseModel, responseModelDiags := NewTeamSpaceMembershipResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read team space membership", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read team space membership", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel TeamSpaceMembershipIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *teamSpaceMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TeamSpaceMembershipModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutTeamSpaceMembershipParams{
		SpaceID:               plan.SpaceID.ValueString(),
		TeamSpaceMembershipID: plan.TeamSpaceMembershipID.ValueString(),
	}

	request, requestDiags := plan.ToTeamSpaceMembershipData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutTeamSpaceMembership(ctx, &request, params)

	tflog.Info(ctx, "team_space_membership.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data TeamSpaceMembershipModel

	switch response := response.(type) {
	case *cm.TeamSpaceMembershipStatusCode:
		responseModel, responseModelDiags := NewTeamSpaceMembershipResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to update team space membership", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel TeamSpaceMembershipIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *teamSpaceMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TeamSpaceMembershipModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.DeleteTeamSpaceMembership(ctx, cm.DeleteTeamSpaceMembershipParams{
		SpaceID:               state.SpaceID.ValueString(),
		TeamSpaceMembershipID: state.TeamSpaceMembershipID.ValueString(),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Team space membership already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete team space membership", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
