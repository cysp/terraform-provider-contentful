package provider

import (
	"context"
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*livePreviewVariablesResource)(nil)
	_ resource.ResourceWithConfigure   = (*livePreviewVariablesResource)(nil)
	_ resource.ResourceWithIdentity    = (*livePreviewVariablesResource)(nil)
	_ resource.ResourceWithImportState = (*livePreviewVariablesResource)(nil)
)

//nolint:ireturn
func NewLivePreviewVariablesResource() resource.Resource {
	return &livePreviewVariablesResource{}
}

type livePreviewVariablesResource struct {
	providerData ContentfulProviderData
}

func livePreviewVariablesIdentityAttributeNames() []string {
	return []string{"space_id", "environment_id"}
}

func (r *livePreviewVariablesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_live_preview_variables"
}

func (r *livePreviewVariablesResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = LivePreviewVariablesResourceSchema(ctx)
}

func (r *livePreviewVariablesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *livePreviewVariablesResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(livePreviewVariablesIdentityAttributeNames())
}

func (r *livePreviewVariablesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, livePreviewVariablesIdentityAttributeNames(), req, resp)
}

func (r *livePreviewVariablesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LivePreviewVariablesModel
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

	request, requestDiagnostics := plan.ToLivePreviewVariablesData()
	resp.Diagnostics.Append(requestDiagnostics...)

	spaceID, spaceDiagnostics := requestRequiredString(plan.SpaceID, path.Root("space_id"))
	resp.Diagnostics.Append(spaceDiagnostics...)

	environmentID, environmentDiagnostics := requestRequiredString(plan.EnvironmentID, path.Root("environment_id"))
	resp.Diagnostics.Append(environmentDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutLivePreviewVariablesParams{SpaceID: spaceID, EnvironmentID: environmentID, XContentfulVersion: 0}
	response, err := r.providerData.client.PutLivePreviewVariables(ctx, &request, params)
	tflog.Info(ctx, "live_preview_variables.create", map[string]any{"params": params})

	variables, ok := response.(*cm.LivePreviewVariables)
	if !ok {
		detail := livePreviewVariablesMutationErrorDetail(response, err)
		if livePreviewVariablesResponseIsError(response, http.StatusConflict, cm.ErrorSysIDVersionMismatch) {
			detail += "\n\nAn existing live preview variables document cannot be overwritten during creation. Inspect it and import it using the space_id/environment_id identifier before managing it with Terraform."
		}

		resp.Diagnostics.AddError("Failed to create live preview variables", detail)

		return
	}

	data, dataDiagnostics, consistencyDiagnostics := ReconcileLivePreviewVariablesMutationResponse(ctx, *variables, plan)
	resp.Diagnostics.Append(dataDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts
	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, livePreviewVariablesIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", variables.Sys.Version)...)
	resp.Diagnostics.Append(consistencyDiagnostics...)
}

func (r *livePreviewVariablesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LivePreviewVariablesModel
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

	params := cm.GetLivePreviewVariablesParams{SpaceID: state.SpaceID.ValueString(), EnvironmentID: state.EnvironmentID.ValueString()}
	response, err := r.providerData.client.GetLivePreviewVariables(ctx, params)
	tflog.Info(ctx, "live_preview_variables.read", map[string]any{"params": params})

	if livePreviewVariablesResponseIsError(response, http.StatusNotFound, cm.ErrorSysIDNotFound) {
		resp.Diagnostics.AddWarning("Live preview variables not found", livePreviewVariablesErrorDetail(response, err))
		resp.State.RemoveResource(ctx)

		return
	}

	variables, ok := response.(*cm.LivePreviewVariables)
	if !ok {
		resp.Diagnostics.AddError("Failed to read live preview variables", livePreviewVariablesErrorDetail(response, err))

		return
	}

	data, dataDiagnostics := NewLivePreviewVariablesResourceModelFromResponse(*variables)
	resp.Diagnostics.Append(dataDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts
	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, livePreviewVariablesIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", variables.Sys.Version)...)
}

func (r *livePreviewVariablesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LivePreviewVariablesModel
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

	request, requestDiagnostics := plan.ToLivePreviewVariablesData()
	resp.Diagnostics.Append(requestDiagnostics...)

	spaceID, spaceDiagnostics := requestRequiredString(plan.SpaceID, path.Root("space_id"))
	resp.Diagnostics.Append(spaceDiagnostics...)

	environmentID, environmentDiagnostics := requestRequiredString(plan.EnvironmentID, path.Root("environment_id"))
	resp.Diagnostics.Append(environmentDiagnostics...)

	version, versionDiagnostics := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutLivePreviewVariablesParams{SpaceID: spaceID, EnvironmentID: environmentID, XContentfulVersion: version}
	response, err := r.providerData.client.PutLivePreviewVariables(ctx, &request, params)
	tflog.Info(ctx, "live_preview_variables.update", map[string]any{"params": params})

	variables, ok := response.(*cm.LivePreviewVariables)
	if !ok {
		resp.Diagnostics.AddError("Failed to update live preview variables", livePreviewVariablesMutationErrorDetail(response, err))

		return
	}

	data, dataDiagnostics, consistencyDiagnostics := ReconcileLivePreviewVariablesMutationResponse(ctx, *variables, plan)
	resp.Diagnostics.Append(dataDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts
	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, livePreviewVariablesIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", variables.Sys.Version)...)
	resp.Diagnostics.Append(consistencyDiagnostics...)
}

func (r *livePreviewVariablesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LivePreviewVariablesModel
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

	params := cm.DeleteLivePreviewVariablesParams{SpaceID: state.SpaceID.ValueString(), EnvironmentID: state.EnvironmentID.ValueString()}
	response, err := r.providerData.client.DeleteLivePreviewVariables(ctx, params)
	tflog.Info(ctx, "live_preview_variables.delete", map[string]any{"params": params})

	if _, ok := response.(*cm.NoContent); ok {
		return
	}

	if livePreviewVariablesResponseIsError(response, http.StatusNotFound, cm.ErrorSysIDNotFound) {
		return
	}

	resp.Diagnostics.AddError("Failed to delete live preview variables", livePreviewVariablesMutationErrorDetail(response, err))
}

func livePreviewVariablesResponseIsError(response any, status int, errorID string) bool {
	failure, ok := response.(*cm.LivePreviewVariablesErrorStatusCode)
	if !ok || failure.StatusCode != status {
		return false
	}

	cmaError, ok := failure.GetError()

	return ok && cmaError.Sys.ID == errorID
}

func livePreviewVariablesErrorDetail(response any, err error) string {
	if failure, ok := response.(*cm.LivePreviewVariablesErrorStatusCode); ok {
		if cmaError, isCMAError := failure.GetError(); isCMAError {
			return fmt.Sprintf("HTTP %d: %s", failure.StatusCode, util.ErrorDetailFromContentfulManagementError(cmaError))
		}

		if serviceError, isServiceError := failure.Response.GetLivePreviewVariablesServiceError(); isServiceError {
			return fmt.Sprintf("HTTP %d: %s: %s", failure.StatusCode, serviceError.Error, serviceError.Message)
		}
	}

	if err != nil {
		return err.Error()
	}

	return "Contentful returned an unrecognized live preview variables response."
}

func livePreviewVariablesMutationErrorDetail(response any, err error) string {
	detail := livePreviewVariablesErrorDetail(response, err)

	failure, isHTTPError := response.(*cm.LivePreviewVariablesErrorStatusCode)
	if err != nil || (isHTTPError && failure.StatusCode >= http.StatusInternalServerError) {
		detail += "\n\nThe write may have committed. Refresh and inspect the remote document before applying again. If creation succeeded remotely but no Terraform state was saved, import the document using space_id/environment_id."
	}

	return detail
}
