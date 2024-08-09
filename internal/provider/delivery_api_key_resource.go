package provider

import (
	"context"
	"net/http"

	"github.com/cysp/terraform-provider-contentful/internal/provider/resource_delivery_api_key"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

var (
	_ resource.Resource                = (*deliveryApiKeyResource)(nil)
	_ resource.ResourceWithConfigure   = (*deliveryApiKeyResource)(nil)
	_ resource.ResourceWithImportState = (*deliveryApiKeyResource)(nil)
)

//nolint:ireturn
func NewDeliveryApiKeyResource() resource.Resource {
	return &deliveryApiKeyResource{}
}

type deliveryApiKeyResource struct {
	providerData ContentfulProviderData
}

func (r *deliveryApiKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delivery_api_key"
}

func (r *deliveryApiKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_delivery_api_key.DeliveryApiKeyResourceSchema(ctx)
}

func (r *deliveryApiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	util.ProviderDataFromResourceConfigureRequest(req, &r.providerData, resp)
}

func (r *deliveryApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	util.ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("environment_id"),
		path.Root("app_definition_id"),
	}, req, resp)
}

func (r *deliveryApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_delivery_api_key.DeliveryApiKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 0

	params := contentfulManagement.PostApiKeyParams{
		SpaceID: data.SpaceId.ValueString(),
	}

	request, requestDiags := data.ToPostAPIKeyReq()
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PostApiKey(ctx, &request, params)

	tflog.Info(ctx, "delivery_api_key.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *contentfulManagement.ApiKey:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(response)...)

	default:
		resp.Diagnostics.AddError("Failed to create delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(util.PrivateDataSetValue(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *deliveryApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_delivery_api_key.DeliveryApiKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := contentfulManagement.GetApiKeyParams{
		SpaceID:  data.SpaceId.ValueString(),
		APIKeyID: data.ApiKeyId.ValueString(),
	}

	response, err := r.providerData.client.GetApiKey(ctx, params)

	tflog.Info(ctx, "delivery_api_key.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0

	switch response := response.(type) {
	case *contentfulManagement.ApiKey:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(response)...)

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

	resp.Diagnostics.Append(util.PrivateDataSetValue(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *deliveryApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_delivery_api_key.DeliveryApiKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	tflog.Info(ctx, "delivery_api_key.update.id", map[string]interface{}{"api_key_is": data.ApiKeyId.ValueStringPointer()})

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int
	currentVersionDiags := util.PrivateDataGetValue(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	params := contentfulManagement.PutApiKeyParams{
		SpaceID:            data.SpaceId.ValueString(),
		APIKeyID:           data.ApiKeyId.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := data.ToPutAPIKeyReq()
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutApiKey(ctx, &request, params)

	tflog.Info(ctx, "delivery_api_key.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *contentfulManagement.ApiKey:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(response)...)

	default:
		resp.Diagnostics.AddError("Failed to update delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(util.PrivateDataSetValue(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *deliveryApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_delivery_api_key.DeliveryApiKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.DeleteApiKey(ctx, contentfulManagement.DeleteApiKeyParams{
		SpaceID:  data.SpaceId.ValueString(),
		APIKeyID: data.ApiKeyId.ValueString(),
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
