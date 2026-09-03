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

func teamSpaceMembershipIdentityAttributeNames() []string {
	return []string{"space_id", "team_space_membership_id"}
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
	resp.IdentitySchema = resourceIdentitySchema(teamSpaceMembershipIdentityAttributeNames())
}

func (r *teamSpaceMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, teamSpaceMembershipIdentityAttributeNames(), req, resp)
}

//nolint:dupl
func (r *teamSpaceMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TeamSpaceMembershipModel

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
		data = NewTeamSpaceMembershipResourceModelFromResponse(response.Response)

	default:
		resp.Diagnostics.AddError("Failed to create team space membership", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, teamSpaceMembershipIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:dupl
func (r *teamSpaceMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TeamSpaceMembershipModel

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
		data = NewTeamSpaceMembershipResourceModelFromResponse(*response)

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read team space membership", util.ErrorDetailFromContentfulManagementResponse(response, err))
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read team space membership", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, teamSpaceMembershipIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:dupl
func (r *teamSpaceMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TeamSpaceMembershipModel

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
		data = NewTeamSpaceMembershipResourceModelFromResponse(response.Response)

	default:
		resp.Diagnostics.AddError("Failed to update team space membership", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, teamSpaceMembershipIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

//nolint:dupl
func (r *teamSpaceMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TeamSpaceMembershipModel

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

	response, err := r.providerData.client.DeleteTeamSpaceMembership(ctx, cm.DeleteTeamSpaceMembershipParams{
		SpaceID:               state.SpaceID.ValueString(),
		TeamSpaceMembershipID: state.TeamSpaceMembershipID.ValueString(),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Team space membership already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

			handled = true
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete team space membership", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
