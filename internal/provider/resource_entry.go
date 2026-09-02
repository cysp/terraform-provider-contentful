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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*entryResource)(nil)
	_ resource.ResourceWithConfigure   = (*entryResource)(nil)
	_ resource.ResourceWithIdentity    = (*entryResource)(nil)
	_ resource.ResourceWithImportState = (*entryResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*entryResource)(nil)
)

//nolint:ireturn
func NewEntryResource() resource.Resource {
	return &entryResource{}
}

type entryResource struct {
	providerData ContentfulProviderData
}

func entryIdentityAttributeNames() []string {
	return []string{"space_id", "environment_id", "entry_id"}
}

const entryPendingPublicationVersionPrivateKey = "pending_publication_version"

func (r *entryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entry"
}

func (r *entryResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = EntryResourceSchema(ctx)
}

func (r *entryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *entryResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(entryIdentityAttributeNames())
}

func (r *entryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, entryIdentityAttributeNames(), req, resp)
}

func (r *entryResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var plan, state EntryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	draftMutation, draftMutationDiags := entryDraftMutationRequired(ctx, plan, state)
	resp.Diagnostics.Append(draftMutationDiags...)

	pendingVersion, pending, pendingDiags := optionalPendingLifecycleVersion(
		ctx, req.Private, entryPendingPublicationVersionPrivateKey, "Entry publication",
	)
	resp.Diagnostics.Append(pendingDiags...)

	if pending {
		pending, pendingDiags = reconcilePendingLifecycleAuthorityCheckpoint(
			ctx, resp.Private, entryPendingPublicationVersionPrivateKey, pendingVersion,
		)
		resp.Diagnostics.Append(pendingDiags...)
	}

	if pending && validateEntryDraftResponse(pendingVersion, state.PublishedVersion).HasError() {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, entryPendingPublicationVersionPrivateKey, nil)...)

		pending = false
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if !draftMutation && !pending {
		return
	}

	version, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if pending && version != pendingVersion {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, entryPendingPublicationVersionPrivateKey, nil)...)

		if !draftMutation {
			return
		}
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("published_version"), types.Int64Unknown())...)
}

func (r *entryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EntryModel

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

	var (
		responseModel EntryModel
		version       int
	)

	if plan.EntryID.IsNull() || plan.EntryID.IsUnknown() {
		responseModel, version = r.createEntry(ctx, plan, &resp.Diagnostics)
	} else {
		responseModel, version = r.createEntryWithID(ctx, plan, &resp.Diagnostics)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, consistencyDiags := projectEntryMutationResponse(ctx, plan, responseModel, entryResponseFieldsCreationDefaults)
	resp.Diagnostics.Append(setEntryIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, responseModel, version)...)

	resp.Diagnostics.Append(validateEntryDraftResponse(version, responseModel.PublishedVersion)...)
	resp.Diagnostics.Append(consistencyDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, entryPendingPublicationVersionPrivateKey, version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	outcome, failure := r.publishAndCheckpointEntry(
		ctx, responseModel, version, entryResponseFieldsCreationDefaults,
		resp.Identity, &resp.State, resp.Private, &resp.Diagnostics,
	)
	resp.Diagnostics.Append(applyPendingLifecycleAuthorityOutcome(
		ctx, resp.Private, entryPendingPublicationVersionPrivateKey, outcome,
	)...)

	if failure != nil {
		resp.Diagnostics.AddWarning(failure.summary, failure.detail)
	}
}

func (r *entryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EntryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pendingVersion, pending, pendingDiags := optionalPendingLifecycleVersion(
		ctx, req.Private, entryPendingPublicationVersionPrivateKey, "Entry publication",
	)
	resp.Diagnostics.Append(pendingDiags...)

	if pending {
		pending, pendingDiags = reconcilePendingLifecycleAuthorityCheckpoint(
			ctx, resp.Private, entryPendingPublicationVersionPrivateKey, pendingVersion,
		)
		resp.Diagnostics.Append(pendingDiags...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceReadContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	getEntryParams := cm.GetEntryParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
		EntryID:       state.EntryID.ValueString(),
	}

	getEntryResponse, err := r.providerData.client.GetEntry(ctx, getEntryParams)

	tflog.Info(ctx, "entry.read", map[string]any{
		"params":   getEntryParams,
		"response": getEntryResponse,
		"err":      err,
	})

	version := 0

	var (
		data                     EntryModel
		observedPublishedVersion cm.OptInt
	)

	switch response := getEntryResponse.(type) {
	case *cm.Entry:
		responseModel, responseModelDiags := NewEntryResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Sys.Version
		observedPublishedVersion = response.Sys.PublishedVersion

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if pending && !pendingLifecycleDraftMatchesCheckpoint(
		pendingVersion, version, observedPublishedVersion, state.PublishedVersion,
	) {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, entryPendingPublicationVersionPrivateKey, nil)...)
	}

	data.Fields = mergeEntryResponseFieldsWithOmissionFallback(data.Fields, state.Fields)
	if entryMetadataEquivalent(data.Metadata, state.Metadata) {
		data.Metadata = state.Metadata
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, entryIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *entryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state EntryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	draftMutation, draftMutationDiags := entryDraftMutationRequired(ctx, plan, state)
	resp.Diagnostics.Append(draftMutationDiags...)

	pendingVersion, pending, pendingDiags := optionalPendingLifecycleVersion(
		ctx, req.Private, entryPendingPublicationVersionPrivateKey, "Entry publication",
	)
	resp.Diagnostics.Append(pendingDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !draftMutation && !pending {
		// Representation-only changes use the effective plan, while publication
		// remains response truth from the prior state.
		resp.State = tfsdk.State(req.Plan)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("published_version"), state.PublishedVersion)...)

		return
	}

	version, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if pending && version != pendingVersion {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, entryPendingPublicationVersionPrivateKey, nil)...)

		if !draftMutation {
			resp.State = tfsdk.State(req.Plan)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("published_version"), state.PublishedVersion)...)

			return
		}
	}

	ctx, cancel, timeoutDiagnostics := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	if !draftMutation {
		outcome, failure := r.publishAndCheckpointEntry(
			ctx, plan, pendingVersion, entryResponseFieldsExact,
			resp.Identity, &resp.State, resp.Private, &resp.Diagnostics,
		)
		resp.Diagnostics.Append(applyPendingLifecycleAuthorityOutcome(
			ctx, resp.Private, entryPendingPublicationVersionPrivateKey, outcome,
		)...)

		if failure != nil {
			resp.Diagnostics.AddError(failure.summary, failure.detail)
		}

		return
	}

	responseModel, version, versionMismatch := r.updateEntry(ctx, plan, version, &resp.Diagnostics)

	if versionMismatch {
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, entryPendingPublicationVersionPrivateKey, nil)...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, consistencyDiags := projectEntryMutationResponse(ctx, plan, responseModel, entryResponseFieldsExact)
	resp.Diagnostics.Append(setEntryIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, responseModel, version)...)

	resp.Diagnostics.Append(validateEntryDraftResponse(version, responseModel.PublishedVersion)...)
	resp.Diagnostics.Append(consistencyDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, entryPendingPublicationVersionPrivateKey, version)...)

	if resp.Diagnostics.HasError() {
		return
	}

	outcome, failure := r.publishAndCheckpointEntry(
		ctx, responseModel, version, entryResponseFieldsExact,
		resp.Identity, &resp.State, resp.Private, &resp.Diagnostics,
	)
	resp.Diagnostics.Append(applyPendingLifecycleAuthorityOutcome(
		ctx, resp.Private, entryPendingPublicationVersionPrivateKey, outcome,
	)...)

	if failure != nil {
		resp.Diagnostics.AddError(failure.summary, failure.detail)
	}
}

func (r *entryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EntryModel

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

	r.unpublishEntry(ctx, state, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	r.deleteEntry(ctx, state, &resp.Diagnostics)
}

func (r *entryResource) createEntry(ctx context.Context, entry EntryModel, diags *diag.Diagnostics) (EntryModel, int) {
	createEntryParams := cm.CreateEntryParams{
		SpaceID:                entry.SpaceID.ValueString(),
		EnvironmentID:          entry.EnvironmentID.ValueString(),
		XContentfulContentType: entry.ContentTypeID.ValueString(),
	}

	createEntryRequest, createEntryRequestDiags := entry.ToEntryRequest(ctx)
	diags.Append(createEntryRequestDiags...)

	if diags.HasError() {
		return entry, 0
	}

	createEntryResponse, err := r.providerData.client.CreateEntry(
		withContentfulRequestNoRetry(ctx), &createEntryRequest, createEntryParams,
	)

	tflog.Info(ctx, "entry.create", map[string]any{
		"params":   createEntryParams,
		"request":  createEntryRequest,
		"response": createEntryResponse,
		"err":      err,
	})

	version := 0

	switch response := createEntryResponse.(type) {
	case *cm.EntryStatusCode:
		responseModel, responseModelDiags := NewEntryResourceModelFromResponse(ctx, response.Response)
		diags.Append(responseModelDiags...)

		entry = responseModel
		version = response.Response.Sys.Version

	default:
		diags.AddError("Failed to create entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	return entry, version
}

func (r *entryResource) createEntryWithID(ctx context.Context, entry EntryModel, diags *diag.Diagnostics) (EntryModel, int) {
	putEntryParams := cm.PutEntryParams{
		SpaceID:                entry.SpaceID.ValueString(),
		EnvironmentID:          entry.EnvironmentID.ValueString(),
		EntryID:                entry.EntryID.ValueString(),
		XContentfulContentType: cm.NewOptPointerString(entry.ContentTypeID.ValueStringPointer()),
	}

	putEntryRequest, putEntryRequestDiags := entry.ToEntryRequest(ctx)
	diags.Append(putEntryRequestDiags...)

	if diags.HasError() {
		return entry, 0
	}

	putEntryResponse, err := r.providerData.client.PutEntry(
		withContentfulRequestNoRetry(ctx), &putEntryRequest, putEntryParams,
	)

	tflog.Info(ctx, "entry.create", map[string]any{
		"params":   putEntryParams,
		"request":  putEntryRequest,
		"response": putEntryResponse,
		"err":      err,
	})

	version := 0

	switch response := putEntryResponse.(type) {
	case *cm.EntryStatusCode:
		responseModel, responseModelDiags := NewEntryResourceModelFromResponse(ctx, response.Response)
		diags.Append(responseModelDiags...)

		entry = responseModel
		version = response.Response.Sys.Version

	default:
		diags.AddError("Failed to create entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	return entry, version
}

func (r *entryResource) updateEntry(
	ctx context.Context,
	entry EntryModel,
	version int,
	diags *diag.Diagnostics,
) (EntryModel, int, bool) {
	putEntryParams := cm.PutEntryParams{
		SpaceID:            entry.SpaceID.ValueString(),
		EnvironmentID:      entry.EnvironmentID.ValueString(),
		EntryID:            entry.EntryID.ValueString(),
		XContentfulVersion: cm.NewOptInt(version),
	}

	putEntryRequest, putEntryRequestDiags := entry.ToEntryRequest(ctx)
	diags.Append(putEntryRequestDiags...)

	if diags.HasError() {
		return entry, version, false
	}

	putEntryResponse, err := r.providerData.client.PutEntry(
		withContentfulRequestNoRetry(ctx), &putEntryRequest, putEntryParams,
	)

	tflog.Info(ctx, "entry.update", map[string]any{
		"params":   putEntryParams,
		"request":  putEntryRequest,
		"response": putEntryResponse,
		"err":      err,
	})

	switch response := putEntryResponse.(type) {
	case *cm.EntryStatusCode:
		responseModel, responseModelDiags := NewEntryResourceModelFromResponse(ctx, response.Response)
		diags.Append(responseModelDiags...)

		entry = responseModel
		version = response.Response.Sys.Version

	default:
		diags.AddError("Failed to update entry", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return entry, version, contentfulResponseIsVersionMismatch(response)
	}

	return entry, version, false
}

func (r *entryResource) publishAndCheckpointEntry(
	ctx context.Context,
	entry EntryModel,
	version int,
	fieldPolicy entryResponseFieldPolicy,
	identity *tfsdk.ResourceIdentity,
	state *tfsdk.State,
	private PrivateProviderData,
	diags *diag.Diagnostics,
) (pendingLifecycleAuthorityOutcome, *pendingLifecycleFailure) {
	publishEntryParams := cm.PublishEntryParams{
		SpaceID: entry.SpaceID.ValueString(), EnvironmentID: entry.EnvironmentID.ValueString(), EntryID: entry.EntryID.ValueString(), XContentfulVersion: version,
	}
	publishEntryResponse, err := r.providerData.client.PublishEntry(
		withContentfulRequestNoRetry(ctx), publishEntryParams,
	)
	tflog.Info(ctx, "entry.publish", map[string]any{"params": publishEntryParams, "response": publishEntryResponse, "err": err})

	response, ok := publishEntryResponse.(*cm.EntryStatusCode)
	if !ok {
		recoveryDetail := fmt.Sprintf(
			"Contentful accepted Entry draft version %d, but publication was not confirmed. Terraform preserved the draft response and exact-version recovery authority for a later operation while that version remains current.",
			version,
		)

		if contentfulResponseIsVersionMismatch(publishEntryResponse) {
			recoveryDetail = fmt.Sprintf(
				"Contentful rejected publication of Entry draft version %d with VersionMismatch. Terraform revoked publication authority and will not fetch or publish a newer version.",
				version,
			)

			return pendingLifecycleAuthorityRevoked, &pendingLifecycleFailure{
				summary: "Failed to publish entry",
				detail:  fmt.Sprintf("%s\n\n%s", util.ErrorDetailFromContentfulManagementResponse(publishEntryResponse, err), recoveryDetail),
			}
		}

		return pendingLifecycleAuthorityRetained, &pendingLifecycleFailure{
			summary: "Failed to publish entry",
			detail:  fmt.Sprintf("%s\n\n%s", util.ErrorDetailFromContentfulManagementResponse(publishEntryResponse, err), recoveryDetail),
		}
	}

	responseModel, responseDiags := NewEntryResourceModelFromResponse(ctx, response.Response)
	diags.Append(responseDiags...)

	if responseDiags.HasError() {
		return pendingLifecycleAuthorityRevoked, nil
	}

	responseVersion := response.Response.Sys.Version

	responseModel, consistencyDiags := projectEntryMutationResponse(ctx, entry, responseModel, fieldPolicy)
	checkpointDiags := setEntryIdentityStateAndVersion(ctx, identity, state, private, responseModel, responseVersion)
	publicationDiags := validateEntryPublicationResponse(version, response.Response.Sys.Version, response.Response.Sys.PublishedVersion)

	diags.Append(checkpointDiags...)
	diags.Append(consistencyDiags...)
	appendNonErrorDiagnostics(diags, publicationDiags)

	if checkpointDiags.HasError() || consistencyDiags.HasError() {
		return pendingLifecycleAuthorityRevoked, nil
	}

	if publicationDiags.HasError() {
		return pendingLifecycleAuthorityRevoked, pendingLifecycleFailureFromDiagnostics(
			"Unexpected entry publication response",
			fmt.Sprintf("Contentful did not return a consistent confirmation for publication of Entry draft version %d. Terraform checkpointed the returned state and revoked exact-version publication authority.", version),
			publicationDiags,
		)
	}

	return pendingLifecycleAuthorityConfirmed, nil
}

func (r *entryResource) deleteEntry(ctx context.Context, entry EntryModel, diags *diag.Diagnostics) {
	deleteEntryParams := cm.DeleteEntryParams{
		SpaceID:       entry.SpaceID.ValueString(),
		EnvironmentID: entry.EnvironmentID.ValueString(),
		EntryID:       entry.EntryID.ValueString(),
	}

	deleteEntryResponse, err := r.providerData.client.DeleteEntry(ctx, deleteEntryParams)

	tflog.Info(ctx, "entry.delete", map[string]any{
		"params":   deleteEntryParams,
		"response": deleteEntryResponse,
		"err":      err,
	})

	switch response := deleteEntryResponse.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if contentfulResponseIsNotFound(response) {
			diags.AddWarning("Entry already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

			handled = true
		}

		if !handled {
			diags.AddError("Failed to delete entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}

func (r *entryResource) unpublishEntry(ctx context.Context, entry EntryModel, diags *diag.Diagnostics) {
	unpublishEntryParams := cm.UnpublishEntryParams{
		SpaceID:       entry.SpaceID.ValueString(),
		EnvironmentID: entry.EnvironmentID.ValueString(),
		EntryID:       entry.EntryID.ValueString(),
	}

	unpublishEntryResponse, err := r.providerData.client.UnpublishEntry(ctx, unpublishEntryParams)

	tflog.Info(ctx, "entry.unpublish", map[string]any{
		"params":   unpublishEntryParams,
		"response": unpublishEntryResponse,
		"err":      err,
	})

	switch response := unpublishEntryResponse.(type) {
	case *cm.NoContent:
	case *cm.Entry:

	default:
		handled := false

		if response, ok := response.(cm.ErrorStatusCodeResponse); ok {
			responseError, _ := response.GetError()
			if contentfulResponseIsNotFound(response) ||
				(response.GetStatusCode() == http.StatusBadRequest && responseError.Sys.ID == "BadRequest" && responseError.Message.Value == "Not published") {
				diags.AddWarning("Entry already unpublished", "")

				handled = true
			}
		}

		if !handled {
			diags.AddError("Failed to unpublish entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
