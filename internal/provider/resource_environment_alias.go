package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func environmentAliasIdentityAttributeNames() []string {
	return []string{"space_id", "environment_alias_id"}
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
	resp.IdentitySchema = resourceIdentitySchema(environmentAliasIdentityAttributeNames())
}

func (r *environmentAliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, environmentAliasIdentityAttributeNames(), req, resp)
}

func (r *environmentAliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentAliasModel

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

	params := cm.CreateOrUpdateEnvironmentAliasParams{
		SpaceID:            plan.SpaceID.ValueString(),
		EnvironmentAliasID: plan.EnvironmentAliasID.ValueString(),
		XContentfulVersion: version,
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
		data = NewEnvironmentAliasResourceModelFromResponse(response.Response)
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, environmentAliasIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *environmentAliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentAliasModel

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

	version := 0

	var data EnvironmentAliasModel

	switch response := response.(type) {
	case *cm.EnvironmentAlias:
		data = NewEnvironmentAliasResourceModelFromResponse(*response)
		version = response.Sys.Version

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, environmentAliasIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *environmentAliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentAliasModel

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

	request, requestDiags := plan.ToEnvironmentAliasData(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	version, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.CreateOrUpdateEnvironmentAliasParams{
		SpaceID:            plan.SpaceID.ValueString(),
		EnvironmentAliasID: plan.EnvironmentAliasID.ValueString(),
		XContentfulVersion: version,
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
		data = NewEnvironmentAliasResourceModelFromResponse(response.Response)
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, environmentAliasIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

//nolint:dupl
func (r *environmentAliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentAliasModel

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

		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Environment alias already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

			handled = true
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
