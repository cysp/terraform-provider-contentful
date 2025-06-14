package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*contentTypeResource)(nil)
	_ resource.ResourceWithConfigure   = (*contentTypeResource)(nil)
	_ resource.ResourceWithImportState = (*contentTypeResource)(nil)
)

//nolint:ireturn
func NewContentTypeResource() resource.Resource {
	return &contentTypeResource{}
}

type contentTypeResource struct {
	providerData ContentfulProviderData
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

func (r *contentTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("content_type_id"),
	}, req, resp)
}

func (r *contentTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ContentTypeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1

	params := cm.PutContentTypeParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		ContentTypeID:      data.ContentTypeID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := data.ToContentTypeRequestFields(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutContentType(ctx, &request, params)

	tflog.Info(ctx, "content_type.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.ContentTypeStatusCode:
		responseModel, responseModelDiags := NewContentTypeResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	activateContentTypeParams := cm.ActivateContentTypeParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		ContentTypeID:      data.ContentTypeID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	activateContentTypeResponse, err := r.providerData.client.ActivateContentType(ctx, activateContentTypeParams)

	tflog.Info(ctx, "content_type.create.activate", map[string]interface{}{
		"params":   activateContentTypeParams,
		"response": activateContentTypeResponse,
		"err":      err,
	})

	switch response := activateContentTypeResponse.(type) {
	case *cm.ContentTypeStatusCode:
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to activate content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *contentTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ContentTypeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetContentTypeParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		ContentTypeID: data.ContentTypeID.ValueString(),
	}

	response, err := r.providerData.client.GetContentType(ctx, params)

	tflog.Info(ctx, "content_type.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0

	switch response := response.(type) {
	case *cm.ContentType:
		responseModel, responseModelDiags := NewContentTypeResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Sys.Version

	default:
		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
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

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *contentTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ContentTypeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int
	currentVersionDiags := GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	putContentTypeParams := cm.PutContentTypeParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		ContentTypeID:      data.ContentTypeID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	putContentTypeRequest, putContentTypeRequestDiags := data.ToContentTypeRequestFields(ctx)
	resp.Diagnostics.Append(putContentTypeRequestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	putContentTypeResponse, err := r.providerData.client.PutContentType(ctx, &putContentTypeRequest, putContentTypeParams)

	tflog.Info(ctx, "content_type.update", map[string]interface{}{
		"params":   putContentTypeParams,
		"request":  putContentTypeRequest,
		"response": putContentTypeResponse,
		"err":      err,
	})

	switch response := putContentTypeResponse.(type) {
	case *cm.ContentTypeStatusCode:
		responseModel, responseModelDiags := NewContentTypeResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	activateContentTypeParams := cm.ActivateContentTypeParams{
		SpaceID:            data.SpaceID.ValueString(),
		EnvironmentID:      data.EnvironmentID.ValueString(),
		ContentTypeID:      data.ContentTypeID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	activateContentTypeResponse, err := r.providerData.client.ActivateContentType(ctx, activateContentTypeParams)

	tflog.Info(ctx, "content_type.update.activate", map[string]interface{}{
		"params":   activateContentTypeParams,
		"response": activateContentTypeResponse,
		"err":      err,
	})

	switch response := activateContentTypeResponse.(type) {
	case *cm.ContentTypeStatusCode:
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to activate content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	r.providerData.editorInterfaceVersionOffset.Increment(data.ContentTypeID.ValueString())
}

func (r *contentTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ContentTypeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deactivateContentTypeParams := cm.DeactivateContentTypeParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		ContentTypeID: data.ContentTypeID.ValueString(),
	}

	deactivateContentTypeResponse, err := r.providerData.client.DeactivateContentType(ctx, deactivateContentTypeParams)

	tflog.Info(ctx, "content_type.delete.deactivate", map[string]interface{}{
		"params":   deactivateContentTypeParams,
		"response": deactivateContentTypeResponse,
		"err":      err,
	})

	switch response := deactivateContentTypeResponse.(type) {
	case *cm.NoContent:
	case *cm.ContentType:

	default:
		handled := false

		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound || (response.Response.Sys.ID == "BadRequest" && response.Response.Message.Value == "Not published") {
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
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		ContentTypeID: data.ContentTypeID.ValueString(),
	}
	deleteContentTypeResponse, err := r.providerData.client.DeleteContentType(ctx, deleteContentTypeParams)

	tflog.Info(ctx, "content_type.delete", map[string]interface{}{
		"params":   deleteContentTypeParams,
		"response": deleteContentTypeResponse,
		"err":      err,
	})

	switch response := deleteContentTypeResponse.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Content type already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
