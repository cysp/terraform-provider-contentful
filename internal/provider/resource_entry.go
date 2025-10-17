package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &EntryResource{}
	_ resource.ResourceWithConfigure   = &EntryResource{}
	_ resource.ResourceWithIdentity    = &EntryResource{}
	_ resource.ResourceWithImportState = &EntryResource{}
)

func NewEntryResource() resource.Resource {
	return &EntryResource{}
}

type EntryResource struct {
	providerData ContentfulProviderData
}

func (r *EntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entry"
}

func (r *EntryResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = EntryResourceSchema(ctx)
}

func (r *EntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *EntryResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"space_id":       identityschema.StringAttribute{RequiredForImport: true},
			"environment_id": identityschema.StringAttribute{RequiredForImport: true},
			"entry_id":       identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *EntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("entry_id"),
	}, req, resp)
}

func (r *EntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EntryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := data.ToEntryRequestFields(ctx)
	resp.Diagnostics.Append(requestDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1
	params := cm.PutEntryParams{
		SpaceID:                data.SpaceID.ValueString(),
		EnvironmentID:          data.EnvironmentID.ValueString(),
		EntryID:                data.EntryID.ValueString(),
		XContentfulContentType: data.ContentTypeID.ValueString(),
		XContentfulVersion:     currentVersion,
	}

	response, err := r.providerData.client.PutEntry(ctx, request, params)

	tflog.Info(ctx, "entry.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.EntryStatusCode:
		responseModel, responseModelDiags := NewEntryResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)
		data = responseModel
		currentVersion = response.Sys.Version
	default:
		resp.Diagnostics.AddError("Failed to create entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	publishParams := cm.PublishEntryParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		EntryID:            data.EntryID.ValueString(),
		XContentfulVersion: currentVersion,
	}
	publishResponse, err := r.providerData.client.PublishEntry(ctx, publishParams)

	tflog.Info(ctx, "entry.create.publish", map[string]interface{}{
		"params":   publishParams,
		"response": publishResponse,
		"err":      err,
	})

	switch response := publishResponse.(type) {
	case *cm.EntryStatusCode:
		currentVersion = response.Sys.Version
	default:
		resp.Diagnostics.AddError("Failed to publish entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel EntryIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

func (r *EntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EntryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetEntryParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		EntryID:       data.EntryID.ValueString(),
	}

	response, err := r.providerData.client.GetEntry(ctx, params)

	tflog.Info(ctx, "entry.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0
	switch response := response.(type) {
	case *cm.Entry:
		responseModel, responseModelDiags := NewEntryResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)
		data = responseModel
		currentVersion = response.Sys.Version
	default:
		if res, ok := response.(cm.StatusCodeResponse); ok && res.GetStatusCode() == http.StatusNotFound {
			resp.Diagnostics.AddWarning("Failed to read entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var identityModel EntryIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

func (r *EntryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EntryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int
	resp.Diagnostics.Append(GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)...)

	request, requestDiags := data.ToEntryRequestFields(ctx)
	resp.Diagnostics.Append(requestDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutEntryParams{
		SpaceID:                data.SpaceID.ValueString(),
		EnvironmentID:          data.EnvironmentID.ValueString(),
		EntryID:                data.EntryID.ValueString(),
		XContentfulContentType: data.ContentTypeID.ValueString(),
		XContentfulVersion:     currentVersion,
	}

	response, err := r.providerData.client.PutEntry(ctx, request, params)

	tflog.Info(ctx, "entry.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.EntryStatusCode:
		responseModel, responseModelDiags := NewEntryResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)
		data = responseModel
		currentVersion = response.Sys.Version
	default:
		resp.Diagnostics.AddError("Failed to update entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	publishParams := cm.PublishEntryParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		EntryID:            data.EntryID.ValueString(),
		XContentfulVersion: currentVersion,
	}
	publishResponse, err := r.providerData.client.PublishEntry(ctx, publishParams)

	tflog.Info(ctx, "entry.update.publish", map[string]interface{}{
		"params":   publishParams,
		"response": publishResponse,
		"err":      err,
	})

	switch response := publishResponse.(type) {
	case *cm.EntryStatusCode:
		currentVersion = response.Sys.Version
	default:
		resp.Diagnostics.AddError("Failed to publish entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel EntryIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

func (r *EntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EntryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int
	resp.Diagnostics.Append(GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)...)

	unpublishParams := cm.UnpublishEntryParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		EntryID:            data.EntryID.ValueString(),
		XContentfulVersion: currentVersion,
	}
	unpublishResponse, err := r.providerData.client.UnpublishEntry(ctx, unpublishParams)

	tflog.Info(ctx, "entry.delete.unpublish", map[string]interface{}{
		"params":   unpublishParams,
		"response": unpublishResponse,
		"err":      err,
	})

	switch unpublishResponse.(type) {
	case *cm.Entry:
		// Success
	default:
		if res, ok := unpublishResponse.(cm.StatusCodeResponse); ok && res.GetStatusCode() == http.StatusNotFound {
			// Already unpublished
		} else {
			resp.Diagnostics.AddError("Failed to unpublish entry", util.ErrorDetailFromContentfulManagementResponse(unpublishResponse, err))
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.DeleteEntryParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		EntryID:       data.EntryID.ValueString(),
	}

	response, err := r.providerData.client.DeleteEntry(ctx, params)

	tflog.Info(ctx, "entry.delete", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.NoContent:
		// Success
	default:
		if res, ok := response.(cm.StatusCodeResponse); ok && res.GetStatusCode() == http.StatusNotFound {
			resp.Diagnostics.AddWarning("Entry already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))
			return
		}
		resp.Diagnostics.AddError("Failed to delete entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}
}
