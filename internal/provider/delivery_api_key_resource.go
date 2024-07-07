package provider

import (
	"context"
	"net/http"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*deliveryApiKeyResource)(nil)
	_ resource.ResourceWithConfigure   = (*deliveryApiKeyResource)(nil)
	_ resource.ResourceWithImportState = (*deliveryApiKeyResource)(nil)
)

//nolint:ireturn,revive,stylecheck
func NewDeliveryApiKeyResource() resource.Resource {
	return &deliveryApiKeyResource{}
}

//nolint:revive,stylecheck
type deliveryApiKeyResource struct {
	providerData ContentfulProviderData
}

func (r *deliveryApiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delivery_api_key"
}

func (r *deliveryApiKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DeliveryApiKeyResourceSchema(ctx)
}

func (r *deliveryApiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *deliveryApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("api_key_id"),
	}, req, resp)
}

//nolint:dupl
func (r *deliveryApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeliveryApiKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1

	params := contentfulManagement.CreateDeliveryApiKeyParams{
		SpaceID: data.SpaceId.ValueString(),
	}

	request, requestDiags := data.ToCreateDeliveryAPIKeyReq(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreateDeliveryApiKey(ctx, &request, params)

	tflog.Info(ctx, "delivery_api_key.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *contentfulManagement.ApiKey:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		resp.Diagnostics.AddError("Failed to create delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *deliveryApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeliveryApiKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := contentfulManagement.GetDeliveryApiKeyParams{
		SpaceID:  data.SpaceId.ValueString(),
		APIKeyID: data.ApiKeyId.ValueString(),
	}

	response, err := r.providerData.client.GetDeliveryApiKey(ctx, params)

	tflog.Info(ctx, "delivery_api_key.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0

	switch response := response.(type) {
	case *contentfulManagement.ApiKey:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		if response, ok := response.(*contentfulManagement.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *deliveryApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeliveryApiKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int
	currentVersionDiags := GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	params := contentfulManagement.UpdateDeliveryApiKeyParams{
		SpaceID:            data.SpaceId.ValueString(),
		APIKeyID:           data.ApiKeyId.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := data.ToUpdateDeliveryAPIKeyReq(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.UpdateDeliveryApiKey(ctx, &request, params)

	tflog.Info(ctx, "delivery_api_key.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *contentfulManagement.ApiKey:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		resp.Diagnostics.AddError("Failed to update delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *deliveryApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeliveryApiKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := contentfulManagement.DeleteDeliveryApiKeyParams{
		SpaceID:  data.SpaceId.ValueString(),
		APIKeyID: data.ApiKeyId.ValueString(),
	}

	response, err := r.providerData.client.DeleteDeliveryApiKey(ctx, params)

	tflog.Info(ctx, "delivery_api_key.delete", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *contentfulManagement.NoContent:

	default:
		handled := false

		if response, ok := response.(*contentfulManagement.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Delivery api key already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
