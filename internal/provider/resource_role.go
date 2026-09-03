//nolint:dupl
package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*roleResource)(nil)
	_ resource.ResourceWithConfigure   = (*roleResource)(nil)
	_ resource.ResourceWithIdentity    = (*roleResource)(nil)
	_ resource.ResourceWithImportState = (*roleResource)(nil)
)

//nolint:ireturn
func NewRoleResource() resource.Resource {
	return &roleResource{}
}

type roleResource struct {
	providerData ContentfulProviderData
}

func roleIdentityAttributeNames() []string { return []string{"space_id", "role_id"} }

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = RoleResourceSchema(ctx)
}

func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *roleResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(roleIdentityAttributeNames())
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, roleIdentityAttributeNames(), req, resp)
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RoleModel

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

	version := 1

	params := cm.CreateRoleParams{
		SpaceID: plan.SpaceID.ValueString(),
	}

	request, requestDiags := plan.ToRoleData()
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreateRole(ctx, &request, params)

	tflog.Info(ctx, "role.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var (
		data             RoleModel
		consistencyDiags diag.Diagnostics
	)

	switch response := response.(type) {
	case *cm.RoleStatusCode:
		var responseDiags diag.Diagnostics

		data, responseDiags, consistencyDiags = ReconcileRoleMutationResponse(ctx, response.Response, plan)

		resp.Diagnostics.Append(responseDiags...)

		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create role", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, roleIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(consistencyDiags...)
}

//nolint:dupl
func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RoleModel

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

	params := cm.GetRoleParams{
		SpaceID: state.SpaceID.ValueString(),
		RoleID:  state.RoleID.ValueString(),
	}

	response, err := r.providerData.client.GetRole(ctx, params)

	tflog.Info(ctx, "role.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	version := 0

	var data RoleModel

	switch response := response.(type) {
	case *cm.Role:
		responseModel, responseModelDiags := NewRoleResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Sys.Version

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read role", util.ErrorDetailFromContentfulManagementResponse(response, err))
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read role", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, roleIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RoleModel

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

	version, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.UpdateRoleParams{
		SpaceID:            plan.SpaceID.ValueString(),
		RoleID:             plan.RoleID.ValueString(),
		XContentfulVersion: version,
	}

	request, requestDiags := plan.ToRoleData()
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.UpdateRole(ctx, &request, params)

	tflog.Info(ctx, "role.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var (
		data             RoleModel
		consistencyDiags diag.Diagnostics
	)

	switch response := response.(type) {
	case *cm.RoleStatusCode:
		var responseDiags diag.Diagnostics

		data, responseDiags, consistencyDiags = ReconcileRoleMutationResponse(ctx, response.Response, plan)

		resp.Diagnostics.Append(responseDiags...)

		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update role", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, roleIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(consistencyDiags...)
}

//nolint:dupl
func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RoleModel

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

	params := cm.DeleteRoleParams{
		SpaceID: state.SpaceID.ValueString(),
		RoleID:  state.RoleID.ValueString(),
	}

	response, err := r.providerData.client.DeleteRole(ctx, params)

	tflog.Info(ctx, "role.delete", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Role already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

			return
		}

		resp.Diagnostics.AddError("Failed to delete role", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}
}
