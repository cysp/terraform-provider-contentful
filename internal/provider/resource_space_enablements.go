package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*spaceEnablementsResource)(nil)
	_ resource.ResourceWithConfigure   = (*spaceEnablementsResource)(nil)
	_ resource.ResourceWithIdentity    = (*spaceEnablementsResource)(nil)
	_ resource.ResourceWithImportState = (*spaceEnablementsResource)(nil)
)

//nolint:ireturn
func NewSpaceEnablementsResource() resource.Resource {
	return &spaceEnablementsResource{}
}

type spaceEnablementsResource struct {
	providerData ContentfulProviderData
}

func spaceEnablementsIdentityAttributeNames() []string { return []string{"space_id"} }

func (r *spaceEnablementsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space_enablements"
}

func (r *spaceEnablementsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = SpaceEnablementsResourceSchema(ctx)
}

func (r *spaceEnablementsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *spaceEnablementsResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(spaceEnablementsIdentityAttributeNames())
}

func (r *spaceEnablementsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, spaceEnablementsIdentityAttributeNames(), req, resp)
}

func (r *spaceEnablementsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan   SpaceEnablementsModel
		config SpaceEnablementsModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToSpaceEnablementData(ctx, config)
	resp.Diagnostics.Append(requestDiags...)

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

	params := cm.PutSpaceEnablementsParams{
		SpaceID:            plan.SpaceID.ValueString(),
		XContentfulVersion: version,
	}

	response, err := r.providerData.client.PutSpaceEnablements(ctx, &request, params)

	tflog.Info(ctx, "space_enablements.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data SpaceEnablementsModel

	switch response := response.(type) {
	case *cm.SpaceEnablementStatusCode:
		data = NewSpaceEnablementsResourceModelFromResponse(response.Response)
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create space enablements", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, spaceEnablementsIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *spaceEnablementsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SpaceEnablementsModel

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

	params := cm.GetSpaceEnablementsParams{
		SpaceID: state.SpaceID.ValueString(),
	}

	response, err := r.providerData.client.GetSpaceEnablements(ctx, params)

	tflog.Info(ctx, "space_enablements.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	version := 0

	var data SpaceEnablementsModel

	switch response := response.(type) {
	case *cm.SpaceEnablement:
		data = NewSpaceEnablementsResourceModelFromResponse(*response)
		version = response.Sys.Version

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read space enablements", util.ErrorDetailFromContentfulManagementResponse(response, err))
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read space enablements", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, spaceEnablementsIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *spaceEnablementsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan   SpaceEnablementsModel
		config SpaceEnablementsModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToSpaceEnablementData(ctx, config)
	resp.Diagnostics.Append(requestDiags...)

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

	params := cm.PutSpaceEnablementsParams{
		SpaceID:            plan.SpaceID.ValueString(),
		XContentfulVersion: version,
	}

	response, err := r.providerData.client.PutSpaceEnablements(ctx, &request, params)

	tflog.Info(ctx, "space_enablements.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data SpaceEnablementsModel

	switch response := response.(type) {
	case *cm.SpaceEnablementStatusCode:
		data = NewSpaceEnablementsResourceModelFromResponse(response.Response)
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update space enablements", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, spaceEnablementsIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *spaceEnablementsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Cannot delete space enablements
}
