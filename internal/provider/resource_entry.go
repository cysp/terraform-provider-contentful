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
	var plan EntryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1

	var responseModel EntryModel

	if plan.EntryID.IsNull() || plan.EntryID.IsUnknown() {
		responseModel, currentVersion = r.createEntry(ctx, plan, &resp.Diagnostics)
	} else {
		responseModel, currentVersion = r.updateEntry(ctx, plan, currentVersion, &resp.Diagnostics)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var identityModel EntryIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &responseModel)...)

	for fieldKey, fieldValue := range plan.Fields.Elements() {
		if !responseModel.Fields.Has(fieldKey) {
			responseModel.Fields.Set(fieldKey, fieldValue)
		}
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &responseModel)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

func (r *entryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EntryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	getEntryParams := cm.GetEntryParams{
		SpaceID:       state.SpaceID.ValueString(),
		EnvironmentID: state.EnvironmentID.ValueString(),
		EntryID:       state.EntryID.ValueString(),
	}

	getEntryResponse, err := r.providerData.client.GetEntry(ctx, getEntryParams)

	tflog.Info(ctx, "entry.read", map[string]interface{}{
		"params":   getEntryParams,
		"response": getEntryResponse,
		"err":      err,
	})

	currentVersion := 0

	var data EntryModel

	switch response := getEntryResponse.(type) {
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

	if resp.Diagnostics.HasError() {
		return
	}

	var identityModel EntryIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	for fieldKey, fieldValue := range state.Fields.Elements() {
		if !data.Fields.Has(fieldKey) {
			data.Fields.Set(fieldKey, fieldValue)
		}
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

func (r *entryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EntryModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int

	currentVersionDiags := GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	responseModel, currentVersion := r.updateEntry(ctx, plan, currentVersion, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	var identityModel EntryIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &responseModel)...)

	for fieldKey, fieldValue := range plan.Fields.Elements() {
		if !responseModel.Fields.Has(fieldKey) {
			responseModel.Fields.Set(fieldKey, fieldValue)
		}
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &responseModel)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

func (r *entryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EntryModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.deleteEntry(ctx, state, &resp.Diagnostics)
}

func (r *entryResource) createEntry(ctx context.Context, data EntryModel, diags *diag.Diagnostics) (EntryModel, int) {
	createEntryParams := cm.CreateEntryParams{
		SpaceID:                data.SpaceID.ValueString(),
		EnvironmentID:          data.EnvironmentID.ValueString(),
		XContentfulContentType: data.ContentTypeID.ValueString(),
	}

	createEntryRequest, createEntryRequestDiags := data.ToEntryRequest(ctx)
	diags.Append(createEntryRequestDiags...)

	if diags.HasError() {
		return data, 0
	}

	createEntryResponse, err := r.providerData.client.CreateEntry(ctx, &createEntryRequest, createEntryParams)

	tflog.Info(ctx, "entry.create", map[string]any{
		"params":   createEntryParams,
		"request":  createEntryRequest,
		"response": createEntryResponse,
		"err":      err,
	})

	currentVersion := 0

	switch response := createEntryResponse.(type) {
	case *cm.EntryStatusCode:
		responseModel, responseModelDiags := NewEntryResourceModelFromResponse(ctx, response.Response)
		diags.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		diags.AddError("Failed to create entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	return data, currentVersion
}

func (r *entryResource) updateEntry(ctx context.Context, data EntryModel, currentVersion int, diags *diag.Diagnostics) (EntryModel, int) {
	putEntryParams := cm.PutEntryParams{
		SpaceID:                data.SpaceID.ValueString(),
		EnvironmentID:          data.EnvironmentID.ValueString(),
		EntryID:                data.EntryID.ValueString(),
		XContentfulContentType: cm.NewOptPointerString(data.ContentTypeID.ValueStringPointer()),
		XContentfulVersion:     currentVersion,
	}

	putEntryRequest, putEntryRequestDiags := data.ToEntryRequest(ctx)
	diags.Append(putEntryRequestDiags...)

	if diags.HasError() {
		return data, currentVersion
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

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		diags.AddError("Failed to update entry", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	return data, currentVersion
}

func (r *entryResource) deleteEntry(ctx context.Context, data EntryModel, diags *diag.Diagnostics) {
	deleteEntryParams := cm.DeleteEntryParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		EntryID:       data.EntryID.ValueString(),
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
