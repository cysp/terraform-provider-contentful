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
	_ resource.Resource                = (*deliveryAPIKeyResource)(nil)
	_ resource.ResourceWithConfigure   = (*deliveryAPIKeyResource)(nil)
	_ resource.ResourceWithIdentity    = (*deliveryAPIKeyResource)(nil)
	_ resource.ResourceWithImportState = (*deliveryAPIKeyResource)(nil)
)

//nolint:ireturn
func NewDeliveryAPIKeyResource() resource.Resource {
	return &deliveryAPIKeyResource{}
}

type deliveryAPIKeyResource struct {
	providerData ContentfulProviderData
}

func (r *deliveryAPIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delivery_api_key"
}

func (r *deliveryAPIKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DeliveryAPIKeyResourceSchema(ctx)
}

func (r *deliveryAPIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *deliveryAPIKeyResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"space_id":   identityschema.StringAttribute{RequiredForImport: true},
			"api_key_id": identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *deliveryAPIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("api_key_id"),
	}, req, resp)
}

//nolint:dupl
func (r *deliveryAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeliveryAPIKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1

	params := cm.CreateDeliveryAPIKeyParams{
		SpaceID: plan.SpaceID.ValueString(),
	}

	request, requestDiags := plan.ToAPIKeyRequestFields(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreateDeliveryAPIKey(ctx, &request, params)

	tflog.Info(ctx, "delivery_api_key.create", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data DeliveryAPIKeyModel

	switch response := response.(type) {
	case *cm.ApiKeyStatusCode:
		responseModel, responseModelDiags := NewDeliveryAPIKeyResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel DeliveryAPIKeyIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

//nolint:dupl
func (r *deliveryAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeliveryAPIKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetDeliveryAPIKeyParams{
		SpaceID:  state.SpaceID.ValueString(),
		APIKeyID: state.APIKeyID.ValueString(),
	}

	response, err := r.providerData.client.GetDeliveryAPIKey(ctx, params)

	tflog.Info(ctx, "delivery_api_key.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0

	var data DeliveryAPIKeyModel

	switch response := response.(type) {
	case *cm.ApiKey:
		responseModel, responseModelDiags := NewDeliveryAPIKeyResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Sys.Version

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel DeliveryAPIKeyIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

func (r *deliveryAPIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeliveryAPIKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int

	currentVersionDiags := GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	params := cm.UpdateDeliveryAPIKeyParams{
		SpaceID:            plan.SpaceID.ValueString(),
		APIKeyID:           plan.APIKeyID.ValueString(),
		XContentfulVersion: currentVersion,
	}

	request, requestDiags := plan.ToAPIKeyRequestFields(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.UpdateDeliveryAPIKey(ctx, &request, params)

	tflog.Info(ctx, "delivery_api_key.update", map[string]any{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	var data DeliveryAPIKeyModel

	switch response := response.(type) {
	case *cm.ApiKeyStatusCode:
		responseModel, responseModelDiags := NewDeliveryAPIKeyResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		currentVersion = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	var identityModel DeliveryAPIKeyIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
}

//nolint:dupl
func (r *deliveryAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeliveryAPIKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.DeleteDeliveryAPIKeyParams{
		SpaceID:  state.SpaceID.ValueString(),
		APIKeyID: state.APIKeyID.ValueString(),
	}

	response, err := r.providerData.client.DeleteDeliveryAPIKey(ctx, params)

	tflog.Info(ctx, "delivery_api_key.delete", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Delivery api key already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete delivery api key", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
