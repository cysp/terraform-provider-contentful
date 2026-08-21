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
	state            ContentTypeModel
	version          int
	consistencyDiags diag.Diagnostics
}

// contentTypeDraftMutationRequired reports whether the Terraform-managed draft
// configuration differs from the prior state. Identity, publication, and
// timeout values do not belong to the Contentful draft request and are
// intentionally excluded. Configuration distinguishes omitted Optional+Computed
// metadata from metadata that is explicitly unknown during planning.
func contentTypeDraftMutationRequired(configMetadata TypedObject[ContentTypeMetadataValue], plan, state ContentTypeModel) bool {
	_, fieldsEquivalent := contentTypeFieldsEquivalentAt(path.Root("fields"), plan.Fields, state.Fields)
	metadataEquivalent := contentTypeDraftMetadataEquivalent(configMetadata, plan.Metadata, state.Metadata)

	return !plan.Name.Equal(state.Name) ||
		!plan.Description.Equal(state.Description) ||
		!plan.DisplayField.Equal(state.DisplayField) ||
		!fieldsEquivalent ||
		!metadataEquivalent
}

func contentTypeDraftMetadataEquivalent(configMetadata, planMetadata, stateMetadata TypedObject[ContentTypeMetadataValue]) bool {
	if configMetadata.IsNull() && planMetadata.IsUnknown() && stateMetadata.IsNull() {
		// Optional+Computed metadata is unknown when it is omitted from
		// configuration and no metadata exists remotely. Treat that unknown as
		// the prior null state so timeout-only updates remain operational-only.
		return true
	}

	planValue, planKnown := planMetadata.GetValue()

	stateValue, stateKnown := stateMetadata.GetValue()
	if !planKnown || !stateKnown {
		return planMetadata.Equal(stateMetadata)
	}

	return contentTypeNormalizedJSONEquivalent(planValue.Annotations, stateValue.Annotations) &&
		planValue.Taxonomy.Equal(stateValue.Taxonomy)
}

// contentTypeActivationRequired reports whether the current Contentful draft
// needs activation. Contentful's lifecycle keeps a present published version
// strictly below the current draft version; a draft one version newer than the
// published version is already activated.
func contentTypeActivationRequired(state ContentTypeModel, version int) bool {
	if state.PublishedVersion.IsNull() {
		return true
	}

	if state.PublishedVersion.IsUnknown() {
		return false
	}

	return state.PublishedVersion.ValueInt64() < int64(version-1)
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

	var plan, state ContentTypeModel
	resp.Diagnostics.Append(resp.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var version int
	resp.Diagnostics.Append(GetPrivateProviderData(ctx, req.Private, "version", &version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	//nolint:contextcheck // attr.Value.Equal and TypedObject.Equal expose no context-aware alternative.
	draftMutationRequired := contentTypeDraftMutationRequired(configMetadata, plan, state)
	activationRequired := contentTypeActivationRequired(state, version)

	plannedPublishedVersion := state.PublishedVersion
	if draftMutationRequired {
		plannedPublishedVersion = types.Int64Unknown()
	} else if activationRequired {
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

	draft, draftDiagnostics := r.putContentTypeDraft(ctx, config, plan, request, 1)
	resp.Diagnostics.Append(draftDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, draft)...)
	// Checkpoint a successful draft response before reporting a response
	// consistency error. A retry can then use its exact returned version.
	resp.Diagnostics.Append(draft.consistencyDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	activated, activationDiagnostics := r.activateContentType(ctx, config, plan, draft.version)
	resp.Diagnostics.Append(activationDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, activated)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(activated.consistencyDiags...)
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

	resp.Diagnostics.Append(plan.validateRequestConfiguration(config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var state ContentTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	//nolint:contextcheck // attr.Value.Equal and TypedObject.Equal expose no context-aware alternative.
	draftMutationRequired := contentTypeDraftMutationRequired(config.Metadata, plan, state)

	var putContentTypeRequest cm.ContentTypeRequestData

	if draftMutationRequired {
		var putContentTypeRequestDiags diag.Diagnostics

		putContentTypeRequest, putContentTypeRequestDiags = plan.ToContentTypeRequestData(ctx, config)
		resp.Diagnostics.Append(putContentTypeRequestDiags...)
	}

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

	resp.Diagnostics.Append(GetPrivateProviderData(ctx, req.Private, "version", &version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	activationRequired := contentTypeActivationRequired(state, version)

	if !draftMutationRequired && !activationRequired {
		resp.State = req.State
		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

		return
	}

	activationVersion := version
	if draftMutationRequired {
		draft, draftDiagnostics := r.putContentTypeDraft(ctx, config, plan, putContentTypeRequest, version)
		resp.Diagnostics.Append(draftDiagnostics...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, draft)...)
		resp.Diagnostics.Append(draft.consistencyDiags...)

		if resp.Diagnostics.HasError() {
			return
		}

		activationVersion = draft.version
	}

	activated, activationDiagnostics := r.activateContentType(ctx, config, plan, activationVersion)
	resp.Diagnostics.Append(activationDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, activated)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// Activation succeeded and its response state is checkpointed. The editor
	// interface now observes the post-activation version even if a consistency
	// diagnostic subsequently makes this apply fail.
	r.providerData.editorInterfaceVersionOffset.Increment(activated.state.SpaceID.ValueString(), activated.state.EnvironmentID.ValueString(), activated.state.ContentTypeID.ValueString())
	resp.Diagnostics.Append(activated.consistencyDiags...)
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
	config ContentTypeModel,
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
		mutationState, mutationStateDiagnostics, consistencyDiags := ProjectContentTypeMutationResponse(ctx, response.Response, config, appliedPlan)
		diagnostics.Append(mutationStateDiagnostics...)

		return contentTypeMutationResult{
			state: mutationState, version: response.Response.Sys.Version, consistencyDiags: consistencyDiags,
		}, diagnostics
	}

	diagnostics.AddError("Failed to save content type draft", util.ErrorDetailFromContentfulManagementResponse(response, err))

	return contentTypeMutationResult{}, diagnostics
}

func (r *contentTypeResource) activateContentType(
	ctx context.Context,
	config ContentTypeModel,
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

	mutationState, mutationStateDiagnostics, consistencyDiags := ProjectContentTypeMutationResponse(ctx, responseContentType.Response, config, appliedPlan)
	diagnostics.Append(mutationStateDiagnostics...)

	return contentTypeMutationResult{
		state: mutationState, version: responseContentType.Response.Sys.Version, consistencyDiags: consistencyDiags,
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
