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
	_ resource.Resource                = (*entryResource)(nil)
	_ resource.ResourceWithConfigure   = (*entryResource)(nil)
	_ resource.ResourceWithIdentity    = (*entryResource)(nil)
	_ resource.ResourceWithImportState = (*entryResource)(nil)
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

func (r *entryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EntryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

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

	request, requestDiags := data.ToEntryRequestFields(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutEntry(ctx, &request, params)

	tflog.Info(ctx, "entry.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.EntryStatusCode:
		responseModel, responseModelDiags := NewEntryResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	publishEntryParams := cm.PublishEntryParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		EntryID:            data.EntryID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	publishEntryResponse, err := r.providerData.client.PublishEntry(ctx, publishEntryParams)

	tflog.Info(ctx, "entry.create.publish", map[string]interface{}{
		"params":   publishEntryParams,
		"response": publishEntryResponse,
		"err":      err,
	})

	switch response := publishEntryResponse.(type) {
	case *cm.EntryStatusCode:
		currentVersion = response.Response.Sys.Version

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

func (r *entryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
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

func (r *entryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EntryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int

	currentVersionDiags := GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	params := cm.PutEntryParams{
		SpaceID:                data.SpaceID.ValueString(),
		EnvironmentID:          data.EnvironmentID.ValueString(),
		EntryID:                data.EntryID.ValueString(),
		XContentfulContentType: data.ContentTypeID.ValueString(),
		XContentfulVersion:     currentVersion,
	}

	request, requestDiags := data.ToEntryRequestFields(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutEntry(ctx, &request, params)

	tflog.Info(ctx, "entry.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.EntryStatusCode:
		responseModel, responseModelDiags := NewEntryResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	publishEntryParams := cm.PublishEntryParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		EntryID:            data.EntryID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	publishEntryResponse, err := r.providerData.client.PublishEntry(ctx, publishEntryParams)

	tflog.Info(ctx, "entry.update.publish", map[string]interface{}{
		"params":   publishEntryParams,
		"response": publishEntryResponse,
		"err":      err,
	})

	switch response := publishEntryResponse.(type) {
	case *cm.EntryStatusCode:
		currentVersion = response.Response.Sys.Version

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

func (r *entryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EntryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int
	resp.Diagnostics.Append(GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)...)

	unpublishEntryParams := cm.UnpublishEntryParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		EntryID:            data.EntryID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	unpublishEntryResponse, err := r.providerData.client.UnpublishEntry(ctx, unpublishEntryParams)

	tflog.Info(ctx, "entry.delete.unpublish", map[string]interface{}{
		"params":   unpublishEntryParams,
		"response": unpublishEntryResponse,
		"err":      err,
	})

	switch unpublishEntryResponse.(type) {
	case *cm.NoContent:
	case *cm.Entry:

	default:
		handled := false

		if response, ok := unpublishEntryResponse.(cm.ErrorStatusCodeResponse); ok {
			responseError, _ := response.GetError()
			if response.GetStatusCode() == http.StatusNotFound || (responseError.Sys.ID == "BadRequest" && responseError.Message.Value == "Not published") {
				resp.Diagnostics.AddWarning("Entry already unpublished", "")

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to unpublish entry", util.ErrorDetailFromContentfulManagementResponse(unpublishEntryResponse, err))
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	deleteEntryParams := cm.DeleteEntryParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		EntryID:       data.EntryID.ValueString(),
	}

	deleteEntryResponse, err := r.providerData.client.DeleteEntry(ctx, deleteEntryParams)

	tflog.Info(ctx, "entry.delete", map[string]interface{}{
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
				resp.Diagnostics.AddWarning("Entry already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
