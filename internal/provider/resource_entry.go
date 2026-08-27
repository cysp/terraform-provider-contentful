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
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"space_id":       identityschema.StringAttribute{RequiredForImport: true},
			"environment_id": identityschema.StringAttribute{RequiredForImport: true},
			"entry_id":       identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *entryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("entry_id"),
	}, req, resp)
}

func (r *entryResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var state EntryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	version, versionDiags := entryPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	pendingVersion, pending, pendingDiags := entryPendingPublishVersion(ctx, req.Private)
	resp.Diagnostics.Append(pendingDiags...)
	resp.Diagnostics.Append(validateEntryStateLifecycle(version, state.PublishedVersion)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var plan EntryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	draftMutation, draftMutationDiags := entryDraftMutationRequired(ctx, plan, state)
	resp.Diagnostics.Append(draftMutationDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if draftMutation || entryPublicationRecoveryRequired(version, pendingVersion, pending, state.PublishedVersion) {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("published_version"), types.Int64Unknown())...)
	}
}

func (r *entryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EntryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

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

	r.publishAndCheckpointEntry(ctx, responseModel, version, entryResponseFieldsCreationDefaults, resp.Identity, &resp.State, resp.Private, &resp.Diagnostics)
}

func (r *entryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EntryModel

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

	var data EntryModel

	switch response := getEntryResponse.(type) {
	case *cm.Entry:
		pendingVersion, pending, pendingDiags := entryPendingPublishVersion(ctx, req.Private)
		resp.Diagnostics.Append(pendingDiags...)

		publishedVersion, published := response.Sys.PublishedVersion.Get()
		if entryPendingPublicationShouldBeCleared(
			response.Sys.Version,
			pendingVersion,
			pending,
			int64(publishedVersion),
			published,
		) {
			resp.Diagnostics.Append(clearEntryPendingPublishVersion(ctx, resp.Private)...)
		}

		resp.Diagnostics.Append(validateObservedEntryLifecycle(response.Sys.Version, response.Sys.PublishedVersion)...)

		responseModel, responseModelDiags := NewEntryResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Sys.Version

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Fields = mergeEntryResponseFieldsWithOmissionFallback(data.Fields, state.Fields)
	if entryMetadataEquivalent(data.Metadata, state.Metadata) {
		data.Metadata = state.Metadata
	}

	data.Timeouts = state.Timeouts

	var identityModel EntryIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, &identityModel, &data)...)

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

	timeout, timeoutDiagnostics := plan.Timeouts.Update(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	version, versionDiags := entryPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	pendingVersion, pending, pendingDiags := entryPendingPublishVersion(ctx, req.Private)
	resp.Diagnostics.Append(pendingDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	draftMutation, draftMutationDiags := entryDraftMutationRequired(ctx, plan, state)
	resp.Diagnostics.Append(draftMutationDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	responseModel := plan
	recoveryRequired := entryPublicationRecoveryRequired(version, pendingVersion, pending, state.PublishedVersion)

	switch {
	case draftMutation:
		expectedDraftVersion := version + 1

		responseModel, version = r.updateEntry(ctx, plan, version, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		var consistencyDiags diag.Diagnostics

		responseModel, consistencyDiags = projectEntryMutationResponse(ctx, plan, responseModel, entryResponseFieldsExact)
		resp.Diagnostics.Append(setEntryIdentityStateAndVersion(ctx, resp.Identity, &resp.State, resp.Private, responseModel, version)...)

		if version != expectedDraftVersion {
			resp.Diagnostics.AddError("Unexpected entry draft version", fmt.Sprintf("Contentful returned version %d after writing version %d.", version, expectedDraftVersion))
		}

		resp.Diagnostics.Append(validateEntryDraftResponse(version, responseModel.PublishedVersion)...)
		resp.Diagnostics.Append(consistencyDiags...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(setEntryPendingPublishVersion(ctx, resp.Private, version)...)

		if resp.Diagnostics.HasError() {
			return
		}
	case recoveryRequired:
		// Publish exactly the provider-written draft version authorized by
		// resource private state.
	default:
		// Representation-only changes use the effective plan, while publication
		// remains response truth from the prior state.
		resp.State = tfsdk.State(req.Plan)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("published_version"), state.PublishedVersion)...)

		return
	}

	r.publishAndCheckpointEntry(ctx, responseModel, version, entryResponseFieldsExact, resp.Identity, &resp.State, resp.Private, &resp.Diagnostics)
}

func (r *entryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EntryModel

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

	r.unpublishEntry(ctx, state, &resp.Diagnostics)

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

	createEntryResponse, err := r.providerData.client.CreateEntry(ctx, &createEntryRequest, createEntryParams)

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

	putEntryResponse, err := r.providerData.client.PutEntry(ctx, &putEntryRequest, putEntryParams)

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

func (r *entryResource) updateEntry(ctx context.Context, entry EntryModel, version int, diags *diag.Diagnostics) (EntryModel, int) {
	putEntryParams := cm.PutEntryParams{
		SpaceID:                entry.SpaceID.ValueString(),
		EnvironmentID:          entry.EnvironmentID.ValueString(),
		EntryID:                entry.EntryID.ValueString(),
		XContentfulContentType: cm.NewOptPointerString(entry.ContentTypeID.ValueStringPointer()),
		XContentfulVersion:     cm.NewOptInt(version),
	}

	putEntryRequest, putEntryRequestDiags := entry.ToEntryRequest(ctx)
	diags.Append(putEntryRequestDiags...)

	if diags.HasError() {
		return entry, version
	}

	putEntryResponse, err := r.providerData.client.PutEntry(ctx, &putEntryRequest, putEntryParams)

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
	}

	return entry, version
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
) {
	publishEntryParams := cm.PublishEntryParams{
		SpaceID: entry.SpaceID.ValueString(), EnvironmentID: entry.EnvironmentID.ValueString(), EntryID: entry.EntryID.ValueString(), XContentfulVersion: version,
	}
	publishEntryResponse, err := r.providerData.client.PublishEntry(ctx, publishEntryParams)
	tflog.Info(ctx, "entry.publish", map[string]any{"params": publishEntryParams, "response": publishEntryResponse, "err": err})

	response, ok := publishEntryResponse.(*cm.EntryStatusCode)
	if !ok {
		diags.AddError("Failed to publish entry", util.ErrorDetailFromContentfulManagementResponse(publishEntryResponse, err))

		return
	}

	diags.Append(clearEntryPendingPublishVersion(ctx, private)...)

	responseModel, responseDiags := NewEntryResourceModelFromResponse(ctx, response.Response)
	diags.Append(responseDiags...)

	if diags.HasError() {
		return
	}

	responseVersion := response.Response.Sys.Version

	fieldPolicy = entryPublicationResponseFieldPolicy(fieldPolicy, version, responseVersion, response.Response.Sys.PublishedVersion)

	responseModel, consistencyDiags := projectEntryMutationResponse(ctx, entry, responseModel, fieldPolicy)
	diags.Append(setEntryIdentityStateAndVersion(ctx, identity, state, private, responseModel, responseVersion)...)
	diags.Append(consistencyDiags...)

	diags.Append(validateEntryPublicationResponse(version, response.Response.Sys.Version, response.Response.Sys.PublishedVersion)...)
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

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				diags.AddWarning("Entry already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
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
			if response.GetStatusCode() == http.StatusNotFound || (responseError.Sys.ID == "BadRequest" && responseError.Message.Value == "Not published") {
				diags.AddWarning("Entry already unpublished", "")

				handled = true
			}
		}

		if !handled {
			diags.AddError("Failed to unpublish entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
