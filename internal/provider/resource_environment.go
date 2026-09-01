package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*environmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*environmentResource)(nil)
	_ resource.ResourceWithIdentity    = (*environmentResource)(nil)
	_ resource.ResourceWithImportState = (*environmentResource)(nil)
)

//nolint:ireturn
func NewEnvironmentResource() resource.Resource {
	return &environmentResource{}
}

type environmentResource struct {
	providerData ContentfulProviderData
}

func environmentIdentityAttributeNames() []string { return []string{"space_id", "environment_id"} }

func (r *environmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *environmentResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = EnvironmentResourceSchema(ctx)
}

func (r *environmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *environmentResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(environmentIdentityAttributeNames())
}

func (r *environmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, environmentIdentityAttributeNames(), req, resp)
}

func (r *environmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentModel

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

	params := cm.CreateOrUpdateEnvironmentParams{
		SpaceID:       plan.SpaceID.ValueString(),
		EnvironmentID: plan.EnvironmentID.ValueString(),
	}

	sourceEnvironment, sourceEnvironmentDiags := plan.ToSourceEnvironmentHeader()
	resp.Diagnostics.Append(sourceEnvironmentDiags...)

	params.XContentfulSourceEnvironment = sourceEnvironment

	request, requestDiags := plan.ToEnvironmentData(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreateOrUpdateEnvironment(ctx, &request, params)

	tflog.Info(ctx, "environment.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data EnvironmentModel

	switch response := response.(type) {
	case *cm.EnvironmentStatusCode:
		responseModel, responseModelDiags := NewEnvironmentResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create environment", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts
	data.SourceEnvironmentID = plan.SourceEnvironmentID

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, environmentIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *environmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentModel

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

	params := cm.GetEnvironmentParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
	}

	response, err := r.providerData.client.GetEnvironment(ctx, params)

	tflog.Info(ctx, "environment.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	version := 0

	var data EnvironmentModel

	switch response := response.(type) {
	case *cm.Environment:
		responseModel, responseModelDiags := NewEnvironmentResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Sys.Version

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read environment", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read environment", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts
	data.SourceEnvironmentID = state.SourceEnvironmentID

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, environmentIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *environmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentModel

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

	request, requestDiags := plan.ToEnvironmentData(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	version, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.CreateOrUpdateEnvironmentParams{
		SpaceID:            plan.SpaceID.ValueString(),
		EnvironmentID:      plan.EnvironmentID.ValueString(),
		XContentfulVersion: cm.NewOptInt(version),
	}

	response, err := r.providerData.client.CreateOrUpdateEnvironment(ctx, &request, params)

	tflog.Info(ctx, "environment.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data EnvironmentModel

	switch response := response.(type) {
	case *cm.EnvironmentStatusCode:
		responseModel, responseModelDiags := NewEnvironmentResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update environment", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts
	data.SourceEnvironmentID = plan.SourceEnvironmentID

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, environmentIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

//nolint:dupl
func (r *environmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentModel

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

	params := cm.DeleteEnvironmentParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
	}

	response, err := r.providerData.client.DeleteEnvironment(ctx, params)

	tflog.Info(ctx, "environment.delete", map[string]any{
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
				resp.Diagnostics.AddWarning("Environment already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete environment", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
