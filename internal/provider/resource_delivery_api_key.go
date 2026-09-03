//nolint:dupl
package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func deliveryAPIKeyIdentityAttributeNames() []string { return []string{"space_id", "api_key_id"} }

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
	resp.IdentitySchema = resourceIdentitySchema(deliveryAPIKeyIdentityAttributeNames())
}

func (r *deliveryAPIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, deliveryAPIKeyIdentityAttributeNames(), req, resp)
}

func (r *deliveryAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan   DeliveryAPIKeyModel
		config DeliveryAPIKeyModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToDeliveryAPIKeyRequestData(ctx, config)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	version := 1

	params := cm.CreateDeliveryAPIKeyParams{
		SpaceID: plan.SpaceID.ValueString(),
	}

	response, err := r.providerData.client.CreateDeliveryAPIKey(ctx, &request, params)

	tflog.Info(ctx, "delivery_api_key.create", map[string]any{
		"params": params,
		// "request": request, omitted to avoid logging sensitive values
		// "response": response, omitted to avoid logging sensitive values
		"err": err,
	})

	var data DeliveryAPIKeyModel

	switch response := response.(type) {
	case *cm.ApiKeyStatusCode:
		responseModel, responseModelDiags := NewDeliveryAPIKeyResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to create delivery API key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, deliveryAPIKeyIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

//nolint:dupl
func (r *deliveryAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeliveryAPIKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceReadContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	params := cm.GetDeliveryAPIKeyParams{
		SpaceID:  state.SpaceID.ValueString(),
		APIKeyID: state.APIKeyID.ValueString(),
	}

	response, err := r.providerData.client.GetDeliveryAPIKey(ctx, params)

	tflog.Info(ctx, "delivery_api_key.read", map[string]any{
		"params": params,
		// "response": response, omitted to avoid logging sensitive values
		"err": err,
	})

	version := 0

	var data DeliveryAPIKeyModel

	switch response := response.(type) {
	case *cm.ApiKey:
		responseModel, responseModelDiags := NewDeliveryAPIKeyResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Sys.Version

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read delivery API key", util.ErrorDetailFromContentfulManagementResponse(response, err))
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read delivery API key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, deliveryAPIKeyIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

func (r *deliveryAPIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan   DeliveryAPIKeyModel
		config DeliveryAPIKeyModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToDeliveryAPIKeyRequestData(ctx, config)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	version, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.UpdateDeliveryAPIKeyParams{
		SpaceID:            plan.SpaceID.ValueString(),
		APIKeyID:           plan.APIKeyID.ValueString(),
		XContentfulVersion: version,
	}

	response, err := r.providerData.client.UpdateDeliveryAPIKey(ctx, &request, params)

	tflog.Info(ctx, "delivery_api_key.update", map[string]any{
		"params": params,
		// "request": request, omitted to avoid logging sensitive values
		// "response": response, omitted to avoid logging sensitive values
		"err": err,
	})

	var data DeliveryAPIKeyModel

	switch response := response.(type) {
	case *cm.ApiKeyStatusCode:
		responseModel, responseModelDiags := NewDeliveryAPIKeyResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Response.Sys.Version

	default:
		resp.Diagnostics.AddError("Failed to update delivery API key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, deliveryAPIKeyIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", version)...)
}

//nolint:dupl
func (r *deliveryAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeliveryAPIKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceDeleteContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

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
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Delivery API key already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

			return
		}

		resp.Diagnostics.AddError("Failed to delete delivery API key", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}
}
