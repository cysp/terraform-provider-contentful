package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*contentTypeResource)(nil)
	_ resource.ResourceWithConfigure   = (*contentTypeResource)(nil)
	_ resource.ResourceWithIdentity    = (*contentTypeResource)(nil)
	_ resource.ResourceWithImportState = (*contentTypeResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*contentTypeResource)(nil)
)

//nolint:ireturn
func NewContentTypeResource() resource.Resource {
	return &contentTypeResource{}
}

type contentTypeResource struct {
	providerData ContentfulProviderData
}

type contentTypeMutationResult struct {
	state   ContentTypeModel
	version int
}

func (r *contentTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_type"
}

func (r *contentTypeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ContentTypeResourceSchema(ctx)
}

func (r *contentTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *contentTypeResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"space_id":        identityschema.StringAttribute{RequiredForImport: true},
			"environment_id":  identityschema.StringAttribute{RequiredForImport: true},
			"content_type_id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *contentTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("content_type_id"),
	}, req, resp)
}

func (r *contentTypeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var configMetadata, stateMetadata TypedObject[ContentTypeMetadataValue]
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("metadata"), &configMetadata)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("metadata"), &stateMetadata)...)

	if resp.Diagnostics.HasError() {
		return
	}

	plannedMetadata, modified := reconcileContentTypeMetadataPlan(configMetadata, stateMetadata)
	if modified {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("metadata"), plannedMetadata)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var state ContentTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var version int
	resp.Diagnostics.Append(GetPrivateProviderData(ctx, req.Private, "version", &version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Substitute the prior computed publication result so equality reflects
	// whether any other planned value requires a mutation.
	planWithStatePublishedVersion := resp.Plan
	resp.Diagnostics.Append(planWithStatePublishedVersion.SetAttribute(ctx, path.Root("published_version"), state.PublishedVersion)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configurationMatchesState := planWithStatePublishedVersion.Raw.Equal(req.State.Raw)

	plannedPublishedVersion := state.PublishedVersion
	if !configurationMatchesState {
		plannedPublishedVersion = types.Int64Unknown()
	} else if state.PublishedVersion.IsNull() || state.PublishedVersion.ValueInt64() < int64(version-1) {
		plannedPublishedVersion = types.Int64Value(int64(version))
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("published_version"), plannedPublishedVersion)...)
}

func (r *contentTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan   ContentTypeModel
		config ContentTypeModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToContentTypeRequestData(ctx, config)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := plan.Timeouts.Create(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	draft, draftDiagnostics := r.putContentTypeDraft(ctx, plan, request, 1)
	resp.Diagnostics.Append(draftDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, draft)...)

	if resp.Diagnostics.HasError() {
		return
	}

	activated, activationDiagnostics := r.activateContentType(ctx, plan, draft.version)
	resp.Diagnostics.Append(activationDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, activated)...)
}

//nolint:dupl
func (r *contentTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContentTypeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := state.Timeouts.Read(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout = max(timeout, minimumStoredResourceOperationTimeout)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	params := cm.GetContentTypeParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
		ContentTypeID: state.ContentTypeID.ValueString(),
	}

	response, err := r.providerData.client.GetContentType(ctx, params)

	tflog.Info(ctx, "content_type.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0

	var data ContentTypeModel

	switch response := response.(type) {
	case *cm.ContentType:
		responseModel, responseModelDiags := NewContentTypeResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Sys.Version

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	var identityModel ContentTypeIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

func (r *contentTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan   ContentTypeModel
		config ContentTypeModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	putContentTypeRequest, putContentTypeRequestDiags := plan.ToContentTypeRequestData(ctx, config)
	resp.Diagnostics.Append(putContentTypeRequestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := plan.Timeouts.Update(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var version int

	versionDiags := GetPrivateProviderData(ctx, req.Private, "version", &version)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	expectedPublishedVersionAfterActivation := int64(version)
	activationOnly := !plan.PublishedVersion.IsNull() &&
		!plan.PublishedVersion.IsUnknown() &&
		plan.PublishedVersion.ValueInt64() == expectedPublishedVersionAfterActivation
	activationVersion := version

	if activationOnly {
		resp.State = req.State
	} else {
		draft, draftDiagnostics := r.putContentTypeDraft(ctx, plan, putContentTypeRequest, version)
		resp.Diagnostics.Append(draftDiagnostics...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, draft)...)

		if resp.Diagnostics.HasError() {
			return
		}

		activationVersion = draft.version
	}

	activated, activationDiagnostics := r.activateContentType(ctx, plan, activationVersion)
	resp.Diagnostics.Append(activationDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, activated)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.providerData.editorInterfaceVersionOffset.Increment(activated.state.SpaceID.ValueString(), activated.state.EnvironmentID.ValueString(), activated.state.ContentTypeID.ValueString())
}

func (r *contentTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContentTypeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := state.Timeouts.Delete(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout = max(timeout, minimumStoredResourceOperationTimeout)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	deactivateContentTypeParams := cm.DeactivateContentTypeParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
		ContentTypeID: state.ContentTypeID.ValueString(),
	}

	deactivateContentTypeResponse, err := r.providerData.client.DeactivateContentType(ctx, deactivateContentTypeParams)

	tflog.Info(ctx, "content_type.delete.deactivate", map[string]any{
		"params":   deactivateContentTypeParams,
		"response": deactivateContentTypeResponse,
		"err":      err,
	})

	switch response := deactivateContentTypeResponse.(type) {
	case *cm.NoContent:
	case *cm.ContentType:

	default:
		handled := false

		if response, ok := response.(cm.ErrorStatusCodeResponse); ok {
			responseError, _ := response.GetError()
			if response.GetStatusCode() == http.StatusNotFound || (responseError.Sys.ID == "BadRequest" && responseError.Message.Value == "Not published") {
				resp.Diagnostics.AddWarning("Content type already deactivated", "")

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to deactivate content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	deleteContentTypeParams := cm.DeleteContentTypeParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
		ContentTypeID: state.ContentTypeID.ValueString(),
	}

	deleteContentTypeResponse, err := r.providerData.client.DeleteContentType(ctx, deleteContentTypeParams)

	tflog.Info(ctx, "content_type.delete", map[string]any{
		"params":   deleteContentTypeParams,
		"response": deleteContentTypeResponse,
		"err":      err,
	})

	switch response := deleteContentTypeResponse.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Content type already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}

func (r *contentTypeResource) putContentTypeDraft(
	ctx context.Context,
	appliedPlan ContentTypeModel,
	request cm.ContentTypeRequestData,
	expectedVersion int,
) (contentTypeMutationResult, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	params := cm.PutContentTypeParams{
		SpaceID:            appliedPlan.SpaceID.ValueString(),
		EnvironmentID:      appliedPlan.EnvironmentID.ValueString(),
		ContentTypeID:      appliedPlan.ContentTypeID.ValueString(),
		XContentfulVersion: expectedVersion,
	}

	response, err := r.providerData.client.PutContentType(ctx, &request, params)

	tflog.Info(ctx, "content_type.put", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	if response, ok := response.(*cm.ContentTypeStatusCode); ok {
		mutationState, mutationStateDiagnostics := NewContentTypeResourceModelForMutationState(ctx, response.Response, appliedPlan)
		diagnostics.Append(mutationStateDiagnostics...)

		mutationState.Timeouts = appliedPlan.Timeouts

		return contentTypeMutationResult{
			state:   mutationState,
			version: response.Response.Sys.Version,
		}, diagnostics
	}

	diagnostics.AddError("Failed to save content type draft", util.ErrorDetailFromContentfulManagementResponse(response, err))

	return contentTypeMutationResult{}, diagnostics
}

func (r *contentTypeResource) activateContentType(
	ctx context.Context,
	appliedPlan ContentTypeModel,
	expectedVersion int,
) (contentTypeMutationResult, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	params := cm.ActivateContentTypeParams{
		SpaceID:            appliedPlan.SpaceID.ValueString(),
		EnvironmentID:      appliedPlan.EnvironmentID.ValueString(),
		ContentTypeID:      appliedPlan.ContentTypeID.ValueString(),
		XContentfulVersion: expectedVersion,
	}

	response, err := r.providerData.client.ActivateContentType(ctx, params)

	tflog.Info(ctx, "content_type.activate", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	responseContentType, ok := response.(*cm.ContentTypeStatusCode)
	if !ok {
		diagnostics.AddError("Failed to activate content type", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return contentTypeMutationResult{}, diagnostics
	}

	mutationState, mutationStateDiagnostics := NewContentTypeResourceModelForMutationState(ctx, responseContentType.Response, appliedPlan)
	diagnostics.Append(mutationStateDiagnostics...)

	mutationState.Timeouts = appliedPlan.Timeouts

	return contentTypeMutationResult{
		state:   mutationState,
		version: responseContentType.Response.Sys.Version,
	}, diagnostics
}

// setContentTypeIdentityStateAndVersion derives and sets the identity and
// state, then records the private version. Each phase stops on errors; private
// data is not transactional with identity and state.
func setContentTypeIdentityStateAndVersion(
	ctx context.Context,
	identity *tfsdk.ResourceIdentity,
	state *tfsdk.State,
	private PrivateProviderData,
	result contentTypeMutationResult,
) diag.Diagnostics {
	var identityModel ContentTypeIdentityModel

	diags := CopyAttributeValues(ctx, &identityModel, &result.state)

	if diags.HasError() {
		return diags
	}

	diags.Append(setResourceIdentityAndState(ctx, identity, state, &identityModel, &result.state)...)

	if diags.HasError() {
		return diags
	}

	diags.Append(SetPrivateProviderData(ctx, private, "version", result.version)...)

	return diags
}
