package provider

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func contentTypeIdentityAttributeNames() []string {
	return []string{"space_id", "environment_id", "content_type_id"}
}

const contentTypePendingActivationVersionPrivateKey = "pending_activation_version"

type contentTypeMutationResult struct {
	state            ContentTypeModel
	version          int
	authorityOutcome pendingLifecycleAuthorityOutcome
	consistencyDiags diag.Diagnostics
	activationDiags  diag.Diagnostics
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
	resp.IdentitySchema = resourceIdentitySchema(contentTypeIdentityAttributeNames())
}

func (r *contentTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, contentTypeIdentityAttributeNames(), req, resp)
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

	if resp.Diagnostics.HasError() {
		return
	}

	pendingVersion, pending, pendingDiags := optionalPendingLifecycleVersion(
		ctx, req.Private, contentTypePendingActivationVersionPrivateKey, "Content Type activation",
	)
	resp.Diagnostics.Append(pendingDiags...)

	if pending {
		pending, pendingDiags = reconcilePendingLifecycleAuthorityCheckpoint(
			ctx, resp.Private, contentTypePendingActivationVersionPrivateKey, pendingVersion,
		)
		resp.Diagnostics.Append(pendingDiags...)
	}

	if pending && !pendingLifecycleDraftIsValid(pendingVersion, state.PublishedVersion) {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, contentTypePendingActivationVersionPrivateKey, nil)...)

		pending = false
	}

	if resp.Diagnostics.HasError() {
		return
	}

	plannedPublishedVersion := state.PublishedVersion

	draftMutation := contentTypeDraftMutationRequired(configMetadata, plan, state)
	if draftMutation || pending {
		version, versionDiags := requiredPrivateVersion(ctx, req.Private)
		resp.Diagnostics.Append(versionDiags...)

		if resp.Diagnostics.HasError() {
			return
		}

		if pending && version != pendingVersion {
			resp.Diagnostics.Append(resp.Private.SetKey(ctx, contentTypePendingActivationVersionPrivateKey, nil)...)

			pending = false
		}
	}

	if draftMutation || pending {
		plannedPublishedVersion = types.Int64Unknown()
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

	ctx, cancel, timeoutDiagnostics := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	draft, draftDiagnostics := r.createContentTypeWithID(ctx, config, plan, request)
	resp.Diagnostics.Append(draftDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, draft)...)
	// Checkpoint a successful draft response before reporting a response
	// consistency or activation error so the recorded state remains truthful.
	resp.Diagnostics.Append(draft.consistencyDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, contentTypePendingActivationVersionPrivateKey, draft.version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	activated, failure, activationDiagnostics := r.activateContentType(
		ctx, config, plan, draft.version,
	)
	resp.Diagnostics.Append(activationDiagnostics...)
	resp.Diagnostics.Append(applyPendingLifecycleAuthorityOutcome(
		ctx, resp.Private, contentTypePendingActivationVersionPrivateKey, activated.authorityOutcome,
	)...)

	if failure != nil {
		resp.Diagnostics.AddWarning(failure.summary, failure.detail)

		return
	}

	resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, activated)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if activated.consistencyDiags.HasError() {
		resp.Diagnostics.Append(activated.consistencyDiags...)
		resp.Diagnostics.Append(activated.activationDiags...)

		return
	}

	resp.Diagnostics.Append(activated.consistencyDiags...)

	if activated.activationDiags.HasError() {
		suffix := "Terraform revoked exact-version activation authority because the returned response no longer represented the checkpointed marked draft."
		appendDiagnosticsWithErrorsAsWarnings(&resp.Diagnostics, activated.activationDiags, suffix)

		return
	}

	resp.Diagnostics.Append(activated.activationDiags...)
}

func (r *contentTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContentTypeModel

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

	version := 0

	var (
		data                     ContentTypeModel
		observedPublishedVersion cm.OptInt
	)

	switch response := response.(type) {
	case *cm.ContentType:
		responseModel, responseModelDiags := NewContentTypeResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Sys.Version
		observedPublishedVersion = response.Sys.PublishedVersion

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	pendingVersion, pending, pendingDiags := optionalPendingLifecycleVersion(
		ctx, req.Private, contentTypePendingActivationVersionPrivateKey, "Content Type activation",
	)
	resp.Diagnostics.Append(pendingDiags...)

	if pending {
		pending, pendingDiags = reconcilePendingLifecycleAuthorityCheckpoint(
			ctx, resp.Private, contentTypePendingActivationVersionPrivateKey, pendingVersion,
		)
		resp.Diagnostics.Append(pendingDiags...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if pending && !pendingLifecycleDraftMatchesCheckpoint(
		pendingVersion, version, observedPublishedVersion, state.PublishedVersion,
	) {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, contentTypePendingActivationVersionPrivateKey, nil)...)
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, contentTypeIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
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

	draftMutationRequired := contentTypeDraftMutationRequired(config.Metadata, plan, state)
	pendingVersion, pending, pendingDiags := optionalPendingLifecycleVersion(
		ctx, req.Private, contentTypePendingActivationVersionPrivateKey, "Content Type activation",
	)
	resp.Diagnostics.Append(pendingDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !draftMutationRequired && !pending {
		resp.State = req.State
		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

		return
	}

	version, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if pending && version != pendingVersion {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, contentTypePendingActivationVersionPrivateKey, nil)...)

		if !draftMutationRequired {
			resp.State = req.State
			state.Timeouts = plan.Timeouts
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

			return
		}
	}

	if !draftMutationRequired {
		r.recoverContentTypeActivation(ctx, config, plan, pendingVersion, resp)

		return
	}

	r.updateAndActivateContentType(ctx, config, plan, version, resp)
}

func (r *contentTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContentTypeModel

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
			if contentfulResponseIsNotFound(response) || (responseError.Sys.ID == "BadRequest" && responseError.Message.Value == "Not published") {
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
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Content type already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

			return
		}

		resp.Diagnostics.AddError("Failed to delete content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}
}

func (r *contentTypeResource) recoverContentTypeActivation(
	ctx context.Context,
	config ContentTypeModel,
	plan ContentTypeModel,
	pendingVersion int,
	resp *resource.UpdateResponse,
) {
	ctx, cancel, timeoutDiagnostics := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	activated, failure, activationDiagnostics := r.activateContentType(
		ctx, config, plan, pendingVersion,
	)
	r.applyContentTypeActivationTransition(ctx, activated, failure, activationDiagnostics, resp)
}

func (r *contentTypeResource) updateAndActivateContentType(
	ctx context.Context,
	config ContentTypeModel,
	plan ContentTypeModel,
	version int,
	resp *resource.UpdateResponse,
) {
	updateContentTypeRequest, updateContentTypeRequestDiags := plan.ToContentTypeRequestData(ctx, config)
	resp.Diagnostics.Append(updateContentTypeRequestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	draft, draftDiagnostics, versionMismatch := r.updateContentType(ctx, config, plan, updateContentTypeRequest, version)
	resp.Diagnostics.Append(draftDiagnostics...)

	if versionMismatch {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, contentTypePendingActivationVersionPrivateKey, nil)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, draft)...)
	resp.Diagnostics.Append(draft.consistencyDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, contentTypePendingActivationVersionPrivateKey, draft.version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	activated, failure, activationDiagnostics := r.activateContentType(
		ctx, config, plan, draft.version,
	)
	r.applyContentTypeActivationTransition(ctx, activated, failure, activationDiagnostics, resp)
}

func (r *contentTypeResource) applyContentTypeActivationTransition(
	ctx context.Context,
	activated contentTypeMutationResult,
	failure *pendingLifecycleFailure,
	activationDiagnostics diag.Diagnostics,
	resp *resource.UpdateResponse,
) {
	resp.Diagnostics.Append(activationDiagnostics...)
	resp.Diagnostics.Append(applyPendingLifecycleAuthorityOutcome(
		ctx, resp.Private, contentTypePendingActivationVersionPrivateKey, activated.authorityOutcome,
	)...)

	if failure != nil {
		resp.Diagnostics.AddError(failure.summary, failure.detail)

		return
	}

	resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, activated)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// The activation response state is checkpointed. The editor interface now
	// observes that version even if a consistency diagnostic makes this apply fail.
	r.providerData.editorInterfaceVersionOffset.Increment(
		activated.state.SpaceID.ValueString(), activated.state.EnvironmentID.ValueString(), activated.state.ContentTypeID.ValueString(),
	)
	resp.Diagnostics.Append(activated.consistencyDiags...)
	resp.Diagnostics.Append(activated.activationDiags...)
}

func (r *contentTypeResource) createContentTypeWithID(
	ctx context.Context,
	config ContentTypeModel,
	appliedPlan ContentTypeModel,
	request cm.ContentTypeRequestData,
) (contentTypeMutationResult, diag.Diagnostics) {
	params := cm.PutContentTypeParams{
		SpaceID:       appliedPlan.SpaceID.ValueString(),
		EnvironmentID: appliedPlan.EnvironmentID.ValueString(),
		ContentTypeID: appliedPlan.ContentTypeID.ValueString(),
	}

	response, err := r.providerData.client.PutContentType(
		withContentfulRequestNoRetry(ctx), &request, params,
	)

	tflog.Info(ctx, "content_type.create_with_id", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	if response, ok := response.(*cm.ContentTypeStatusCode); ok {
		return projectContentTypeDraftMutationResponse(ctx, config, appliedPlan, response.Response)
	}

	var diagnostics diag.Diagnostics
	diagnostics.AddError("Failed to create content type", util.ErrorDetailFromContentfulManagementResponse(response, err))

	return contentTypeMutationResult{}, diagnostics
}

func (r *contentTypeResource) updateContentType(
	ctx context.Context,
	config ContentTypeModel,
	appliedPlan ContentTypeModel,
	request cm.ContentTypeRequestData,
	version int,
) (contentTypeMutationResult, diag.Diagnostics, bool) {
	params := cm.PutContentTypeParams{
		SpaceID:            appliedPlan.SpaceID.ValueString(),
		EnvironmentID:      appliedPlan.EnvironmentID.ValueString(),
		ContentTypeID:      appliedPlan.ContentTypeID.ValueString(),
		XContentfulVersion: cm.NewOptInt(version),
	}

	response, err := r.providerData.client.PutContentType(
		withContentfulRequestNoRetry(ctx), &request, params,
	)

	tflog.Info(ctx, "content_type.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	if response, ok := response.(*cm.ContentTypeStatusCode); ok {
		result, diagnostics := projectContentTypeDraftMutationResponse(ctx, config, appliedPlan, response.Response)

		return result, diagnostics, false
	}

	var diagnostics diag.Diagnostics
	diagnostics.AddError("Failed to update content type", util.ErrorDetailFromContentfulManagementResponse(response, err))

	return contentTypeMutationResult{}, diagnostics, contentfulResponseIsVersionMismatch(response)
}

func projectContentTypeDraftMutationResponse(
	ctx context.Context,
	config ContentTypeModel,
	appliedPlan ContentTypeModel,
	response cm.ContentType,
) (contentTypeMutationResult, diag.Diagnostics) {
	mutationState, diagnostics, consistencyDiags := ProjectContentTypeMutationResponse(ctx, response, config, appliedPlan)
	consistencyDiags.Append(validateContentTypeDraftResponse(response)...)

	return contentTypeMutationResult{
		state: mutationState, version: response.Sys.Version, consistencyDiags: consistencyDiags,
	}, diagnostics
}

func validateContentTypeDraftResponse(response cm.ContentType) diag.Diagnostics {
	publishedVersion := types.Int64Null()
	if published, ok := response.Sys.PublishedVersion.Get(); ok {
		publishedVersion = types.Int64Value(int64(published))
	}

	var diagnostics diag.Diagnostics
	if !pendingLifecycleDraftIsValid(response.Sys.Version, publishedVersion) {
		diagnostics.AddAttributeError(
			path.Root("published_version"),
			"Unexpected Contentful content type draft response",
			fmt.Sprintf("Contentful returned draft version %d with an invalid publishedVersion.", response.Sys.Version),
		)
	}

	return diagnostics
}

func (r *contentTypeResource) activateContentType(
	ctx context.Context,
	config ContentTypeModel,
	appliedPlan ContentTypeModel,
	expectedVersion int,
) (contentTypeMutationResult, *pendingLifecycleFailure, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	params := cm.ActivateContentTypeParams{
		SpaceID:            appliedPlan.SpaceID.ValueString(),
		EnvironmentID:      appliedPlan.EnvironmentID.ValueString(),
		ContentTypeID:      appliedPlan.ContentTypeID.ValueString(),
		XContentfulVersion: expectedVersion,
	}

	response, err := r.providerData.client.ActivateContentType(
		withContentfulRequestNoRetry(ctx), params,
	)

	tflog.Info(ctx, "content_type.activate", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	responseContentType, ok := response.(*cm.ContentTypeStatusCode)
	if !ok {
		recoveryDetail := fmt.Sprintf(
			"Contentful accepted Content Type draft version %d, but activation was not confirmed. Terraform preserved the draft response and exact-version recovery authority for a later operation while that version remains current.",
			expectedVersion,
		)

		if contentfulResponseIsVersionMismatch(response) {
			recoveryDetail = fmt.Sprintf(
				"Contentful rejected activation of Content Type draft version %d with VersionMismatch. Terraform revoked activation authority and will not fetch or activate a newer version.",
				expectedVersion,
			)

			return contentTypeMutationResult{authorityOutcome: pendingLifecycleAuthorityRevoked}, &pendingLifecycleFailure{
				summary: "Failed to activate content type",
				detail:  fmt.Sprintf("%s\n\n%s", util.ErrorDetailFromContentfulManagementResponse(response, err), recoveryDetail),
			}, diagnostics
		}

		return contentTypeMutationResult{authorityOutcome: pendingLifecycleAuthorityRetained}, &pendingLifecycleFailure{
			summary: "Failed to activate content type",
			detail:  fmt.Sprintf("%s\n\n%s", util.ErrorDetailFromContentfulManagementResponse(response, err), recoveryDetail),
		}, diagnostics
	}

	mutationState, mutationStateDiagnostics, consistencyDiags := ProjectContentTypeMutationResponse(ctx, responseContentType.Response, config, appliedPlan)
	diagnostics.Append(mutationStateDiagnostics...)

	activationDiags := validateContentTypeActivationResponse(expectedVersion, responseContentType.Response)
	outcome := pendingLifecycleAuthorityRevoked

	if !mutationStateDiagnostics.HasError() && !consistencyDiags.HasError() {
		if !activationDiags.HasError() {
			outcome = pendingLifecycleAuthorityConfirmed
		}
	}

	return contentTypeMutationResult{
		state: mutationState, version: responseContentType.Response.Sys.Version,
		authorityOutcome: outcome, consistencyDiags: consistencyDiags, activationDiags: activationDiags,
	}, nil, diagnostics
}

func validateContentTypeActivationResponse(expectedVersion int, response cm.ContentType) diag.Diagnostics {
	publishedVersion := types.Int64Null()
	if published, ok := response.Sys.PublishedVersion.Get(); ok {
		publishedVersion = types.Int64Value(int64(published))
	}

	var diagnostics diag.Diagnostics
	if publishedVersion.IsNull() {
		diagnostics.AddAttributeError(
			path.Root("published_version"),
			"Unexpected Contentful content type activation response",
			"Contentful accepted the activation request but omitted sys.publishedVersion.",
		)

		return diagnostics
	}

	if publishedVersion.ValueInt64() != int64(expectedVersion) {
		diagnostics.AddAttributeError(
			path.Root("published_version"),
			"Unexpected Contentful content type activation response",
			fmt.Sprintf("Contentful accepted activation of version %d but returned publishedVersion %d.", expectedVersion, publishedVersion.ValueInt64()),
		)
	}

	if response.Sys.Version <= expectedVersion {
		diagnostics.AddError(
			"Unexpected Contentful content type activation response",
			fmt.Sprintf("Contentful accepted activation of version %d but returned current version %d; the current version must be greater than the published version.", expectedVersion, response.Sys.Version),
		)
	}

	return diagnostics
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
	diags := setResourceIdentityAndState(ctx, identity, state, contentTypeIdentityAttributeNames(), &result.state)

	if diags.HasError() {
		return diags
	}

	diags.Append(SetPrivateProviderData(ctx, private, "version", result.version)...)

	return diags
}
