package provider

import (
	"context"
	"net/http"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/resource_content_type"
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
	resp.Schema = resource_content_type.ContentTypeResourceSchema(ctx)
}

func (r *contentTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	util.ProviderDataFromResourceConfigureRequest(req, &r.providerData, resp)
}

func (r *contentTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("content_type_id"),
	}, req, resp)
}

func (r *contentTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_content_type.ContentTypeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1

	params := contentfulManagement.PutContentTypeParams{
		SpaceID:            data.SpaceId.ValueString(),
		EnvironmentID:      data.EnvironmentId.ValueString(),
		ContentTypeID:      data.ContentTypeId.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := data.ToPutContentTypeReq(ctx)
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
	case *contentfulManagement.PutContentTypeCreated:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, (*contentfulManagement.ContentType)(response))...)

	default:
		resp.Diagnostics.AddError("Failed to create content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	activateContentTypeParams := contentfulManagement.ActivateContentTypeParams{
		SpaceID:            data.SpaceId.ValueString(),
		EnvironmentID:      data.EnvironmentId.ValueString(),
		ContentTypeID:      data.ContentTypeId.ValueString(),
		XContentfulVersion: currentVersion,
	}

	activateContentTypeResponse, err := r.providerData.client.ActivateContentType(ctx, activateContentTypeParams)

	tflog.Info(ctx, "content_type.create.activate", map[string]interface{}{
		"params":   activateContentTypeParams,
		"response": activateContentTypeResponse,
		"err":      err,
	})

	switch response := activateContentTypeResponse.(type) {
	case *contentfulManagement.ContentType:
		currentVersion = response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to activate content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(util.PrivateDataSetValue(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *contentTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_content_type.ContentTypeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := contentfulManagement.GetContentTypeParams{
		SpaceID:       data.SpaceId.ValueString(),
		EnvironmentID: data.EnvironmentId.ValueString(),
		ContentTypeID: data.ContentTypeId.ValueString(),
	}

	response, err := r.providerData.client.GetContentType(ctx, params)

	tflog.Info(ctx, "content_type.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0

	switch response := response.(type) {
	case *contentfulManagement.ContentType:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		if response, ok := response.(*contentfulManagement.ErrorStatusCode); ok {
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

	resp.Diagnostics.Append(util.PrivateDataSetValue(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *contentTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_content_type.ContentTypeModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int
	currentVersionDiags := util.PrivateDataGetValue(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	putContentTypeParams := contentfulManagement.PutContentTypeParams{
		SpaceID:            data.SpaceId.ValueString(),
		EnvironmentID:      data.EnvironmentId.ValueString(),
		ContentTypeID:      data.ContentTypeId.ValueString(),
		XContentfulVersion: currentVersion,
	}

	putContentTypeRequest, putContentTypeRequestDiags := data.ToPutContentTypeReq(ctx)
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
	case *contentfulManagement.PutContentTypeOK:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, (*contentfulManagement.ContentType)(response))...)

	default:
		resp.Diagnostics.AddError("Failed to update content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	activateContentTypeParams := contentfulManagement.ActivateContentTypeParams{
		SpaceID:            data.SpaceId.ValueString(),
		EnvironmentID:      data.EnvironmentId.ValueString(),
		ContentTypeID:      data.ContentTypeId.ValueString(),
		XContentfulVersion: currentVersion,
	}

	activateContentTypeResponse, err := r.providerData.client.ActivateContentType(ctx, activateContentTypeParams)

	tflog.Info(ctx, "content_type.update.activate", map[string]interface{}{
		"params":   activateContentTypeParams,
		"response": activateContentTypeResponse,
		"err":      err,
	})

	switch response := activateContentTypeResponse.(type) {
	case *contentfulManagement.ContentType:
		currentVersion = response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to activate content type", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(util.PrivateDataSetValue(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	r.providerData.editorInterfaceVersionOffset.Increment(data.ContentTypeId.ValueString())
}

//nolint:cyclop
func (r *contentTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_content_type.ContentTypeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	deactivateContentTypeParams := contentfulManagement.DeactivateContentTypeParams{
		SpaceID:       data.SpaceId.ValueString(),
		EnvironmentID: data.EnvironmentId.ValueString(),
		ContentTypeID: data.ContentTypeId.ValueString(),
	}

	deactivateContentTypeResponse, err := r.providerData.client.DeactivateContentType(ctx, deactivateContentTypeParams)

	tflog.Info(ctx, "content_type.delete.deactivate", map[string]interface{}{
		"params":   deactivateContentTypeParams,
		"response": deactivateContentTypeResponse,
		"err":      err,
	})

	switch response := deactivateContentTypeResponse.(type) {
	case *contentfulManagement.NoContent:
	case *contentfulManagement.ContentType:

	default:
		handled := false

		if response, ok := response.(*contentfulManagement.ErrorStatusCode); ok {
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

	deleteContentTypeParams := contentfulManagement.DeleteContentTypeParams{
		SpaceID:       data.SpaceId.ValueString(),
		EnvironmentID: data.EnvironmentId.ValueString(),
		ContentTypeID: data.ContentTypeId.ValueString(),
	}
	deleteContentTypeResponse, err := r.providerData.client.DeleteContentType(ctx, deleteContentTypeParams)

	tflog.Info(ctx, "content_type.delete", map[string]interface{}{
		"params":   deleteContentTypeParams,
		"response": deleteContentTypeResponse,
		"err":      err,
	})

	switch response := deleteContentTypeResponse.(type) {
	case *contentfulManagement.NoContent:

	default:
		handled := false

		if response, ok := response.(*contentfulManagement.ErrorStatusCode); ok {
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
