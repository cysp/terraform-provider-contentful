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
	_ resource.Resource                = (*extensionResource)(nil)
	_ resource.ResourceWithConfigure   = (*extensionResource)(nil)
	_ resource.ResourceWithImportState = (*extensionResource)(nil)
)

//nolint:ireturn
func NewExtensionResource() resource.Resource {
	return &extensionResource{}
}

type extensionResource struct {
	providerData ContentfulProviderData
}

func (r *extensionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_extension"
}

func (r *extensionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ExtensionResourceSchema(ctx)
}

func (r *extensionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *extensionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("extension_id"),
	}, req, resp)
}

//nolint:dupl
func (r *extensionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ExtensionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutExtensionParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		ExtensionID:   data.ExtensionID.ValueString(),
	}

	extensionFields, extensionFieldsDiags := data.ToExtensionFields()
	resp.Diagnostics.Append(extensionFieldsDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := cm.PutExtensionReq{
		Extension: extensionFields,
	}

	response, err := r.providerData.client.PutExtension(ctx, &request, params)

	tflog.Info(ctx, "extension.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.Extension:
		resp.Diagnostics.Append(data.ReadFromResponse(response)...)

	default:
		resp.Diagnostics.AddError("Failed to create extension", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *extensionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ExtensionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetExtensionParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		ExtensionID:   data.ExtensionID.ValueString(),
	}

	response, err := r.providerData.client.GetExtension(ctx, params)

	tflog.Info(ctx, "extension.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.Extension:
		resp.Diagnostics.Append(data.ReadFromResponse(response)...)

	default:
		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read extension", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read extension", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *extensionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ExtensionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.PutExtensionParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		ExtensionID:   data.ExtensionID.ValueString(),
	}

	extensionFields, extensionFieldsDiags := data.ToExtensionFields()
	resp.Diagnostics.Append(extensionFieldsDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := cm.PutExtensionReq{
		Extension: extensionFields,
	}

	response, err := r.providerData.client.PutExtension(ctx, &request, params)

	tflog.Info(ctx, "extension.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.Extension:
		resp.Diagnostics.Append(data.ReadFromResponse(response)...)

	default:
		resp.Diagnostics.AddError("Failed to update extension", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *extensionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ExtensionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.DeleteExtensionParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		ExtensionID:   data.ExtensionID.ValueString(),
	}

	response, err := r.providerData.client.DeleteExtension(ctx, params)

	tflog.Info(ctx, "extension.delete", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Extension already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete extension", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
