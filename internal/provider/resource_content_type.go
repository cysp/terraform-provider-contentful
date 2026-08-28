package provider

import (
	"context"
	"fmt"
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

type contentTypePublicationState uint8

const (
	contentTypePublicationUnknown contentTypePublicationState = iota
	contentTypePublicationUnpublished
	contentTypePublicationActive
	contentTypePublicationPendingDraft
)

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

// classifyContentTypePublicationState interprets the current optimistic-lock
// version and the last published version as one lifecycle value. Existing
// Content Types have positive current versions, and a present publication must
// precede that current version. Other tuples cannot be used safely as lifecycle
// state or lock tokens.
func classifyContentTypePublicationState(version int, publishedVersion types.Int64) (contentTypePublicationState, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if version <= 0 {
		diagnostics.AddError(
			"Invalid Contentful content type version",
			fmt.Sprintf("Contentful content type sys.version must be positive; got %d.", version),
		)

		return contentTypePublicationUnknown, diagnostics
	}

	if publishedVersion.IsUnknown() {
		return contentTypePublicationUnknown, diagnostics
	}

	if publishedVersion.IsNull() {
		return contentTypePublicationUnpublished, diagnostics
	}

	published := publishedVersion.ValueInt64()
	if published < 0 || published >= int64(version) {
		diagnostics.AddError(
			"Invalid Contentful content type publication version",
			fmt.Sprintf("Contentful content type sys.publishedVersion must be non-negative and less than sys.version; got publishedVersion %d and version %d.", published, version),
		)

		return contentTypePublicationUnknown, diagnostics
	}

	if published == int64(version-1) {
		return contentTypePublicationActive, diagnostics
	}

	return contentTypePublicationPendingDraft, diagnostics
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

	if resp.Diagnostics.HasError() {
		return
	}

	plannedPublishedVersion := state.PublishedVersion
	//nolint:contextcheck // attr.Value.Equal and TypedObject.Equal expose no context-aware alternative.
	if contentTypeDraftMutationRequired(configMetadata, plan, state) {
		_, versionDiags := requiredPrivateVersion(ctx, req.Private)
		resp.Diagnostics.Append(versionDiags...)

		if resp.Diagnostics.HasError() {
			return
		}

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

	timeout, timeoutDiagnostics := plan.Timeouts.Create(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
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

	if !draftMutationRequired {
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

	updateContentTypeRequest, updateContentTypeRequestDiags := plan.ToContentTypeRequestData(ctx, config)
	resp.Diagnostics.Append(updateContentTypeRequestDiags...)

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

	draft, draftDiagnostics := r.updateContentType(ctx, config, plan, updateContentTypeRequest, version)
	resp.Diagnostics.Append(draftDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setContentTypeIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, draft)...)
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

	response, err := r.providerData.client.PutContentType(ctx, &request, params)

	tflog.Info(ctx, "content_type.create_with_id", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	if response, ok := response.(*cm.ContentTypeStatusCode); ok {
		return projectContentTypeDraftMutationResponse(ctx, config, appliedPlan, response.Response, 1)
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
	currentVersion int,
) (contentTypeMutationResult, diag.Diagnostics) {
	params := cm.PutContentTypeParams{
		SpaceID:            appliedPlan.SpaceID.ValueString(),
		EnvironmentID:      appliedPlan.EnvironmentID.ValueString(),
		ContentTypeID:      appliedPlan.ContentTypeID.ValueString(),
		XContentfulVersion: cm.NewOptInt(currentVersion),
	}

	response, err := r.providerData.client.PutContentType(ctx, &request, params)

	tflog.Info(ctx, "content_type.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	if response, ok := response.(*cm.ContentTypeStatusCode); ok {
		return projectContentTypeDraftMutationResponse(ctx, config, appliedPlan, response.Response, currentVersion+1)
	}

	var diagnostics diag.Diagnostics
	diagnostics.AddError("Failed to update content type", util.ErrorDetailFromContentfulManagementResponse(response, err))

	return contentTypeMutationResult{}, diagnostics
}

func projectContentTypeDraftMutationResponse(
	ctx context.Context,
	config ContentTypeModel,
	appliedPlan ContentTypeModel,
	response cm.ContentType,
	expectedVersion int,
) (contentTypeMutationResult, diag.Diagnostics) {
	mutationState, diagnostics, consistencyDiags := ProjectContentTypeMutationResponse(ctx, response, config, appliedPlan)
	consistencyDiags.Append(validateContentTypeDraftResponse(expectedVersion, response)...)

	return contentTypeMutationResult{
		state: mutationState, version: response.Sys.Version, consistencyDiags: consistencyDiags,
	}, diagnostics
}

func validateContentTypeDraftResponse(expectedVersion int, response cm.ContentType) diag.Diagnostics {
	publishedVersion := types.Int64Null()
	if published, ok := response.Sys.PublishedVersion.Get(); ok {
		publishedVersion = types.Int64Value(int64(published))
	}

	publicationState, diagnostics := classifyContentTypePublicationState(response.Sys.Version, publishedVersion)
	if diagnostics.HasError() {
		return diagnostics
	}

	if response.Sys.Version != expectedVersion {
		diagnostics.AddError(
			"Unexpected Contentful content type draft response",
			fmt.Sprintf("Contentful accepted the draft update for version %d but returned version %d.", expectedVersion, response.Sys.Version),
		)
	}

	if publicationState == contentTypePublicationActive {
		diagnostics.AddAttributeError(
			path.Root("published_version"),
			"Unexpected Contentful content type draft response",
			"Contentful accepted the draft update but returned an active Content Type instead of an unpublished current draft.",
		)
	}

	return diagnostics
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
		diagnostics.AddError(
			"Failed to activate content type",
			fmt.Sprintf(
				"%s\n\nContentful accepted Content Type draft version %d, but activation was not confirmed. Terraform preserved the draft response and an unchanged later apply will not retry activation. Inspect the Content Type in Contentful, then activate it manually if needed or make another Terraform-managed draft change.",
				util.ErrorDetailFromContentfulManagementResponse(response, err),
				expectedVersion,
			),
		)

		return contentTypeMutationResult{}, diagnostics
	}

	mutationState, mutationStateDiagnostics, consistencyDiags := ProjectContentTypeMutationResponse(ctx, responseContentType.Response, config, appliedPlan)
	diagnostics.Append(mutationStateDiagnostics...)
	consistencyDiags.Append(validateContentTypeActivationResponse(expectedVersion, responseContentType.Response)...)

	return contentTypeMutationResult{
		state: mutationState, version: responseContentType.Response.Sys.Version, consistencyDiags: consistencyDiags,
	}, diagnostics
}

func validateContentTypeActivationResponse(expectedVersion int, response cm.ContentType) diag.Diagnostics {
	publishedVersion := types.Int64Null()
	if published, ok := response.Sys.PublishedVersion.Get(); ok {
		publishedVersion = types.Int64Value(int64(published))
	}

	publicationState, diagnostics := classifyContentTypePublicationState(response.Sys.Version, publishedVersion)
	if diagnostics.HasError() {
		return diagnostics
	}

	if publicationState != contentTypePublicationActive {
		diagnostics.AddAttributeError(
			path.Root("published_version"),
			"Unexpected Contentful content type activation response",
			"Contentful accepted the activation request but did not return an active Content Type.",
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
