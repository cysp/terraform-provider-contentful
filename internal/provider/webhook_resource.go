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
	_ resource.Resource                = (*webhookResource)(nil)
	_ resource.ResourceWithConfigure   = (*webhookResource)(nil)
	_ resource.ResourceWithImportState = (*webhookResource)(nil)
)

//nolint:ireturn
func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

type webhookResource struct {
	providerData ContentfulProviderData
}

func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := WebhookResourceSchema(ctx)
	s.Attributes["filters"] = WebhookFiltersSchema(ctx, true)
	resp.Schema = s
}

func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("space_id"),
		path.Root("webhook_id"),
	}, req, resp)
}

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WebhookModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	currentVersion := 1

	params := contentfulManagement.CreateWebhookDefinitionParams{
		SpaceID: data.SpaceId.ValueString(),
	}

	request, requestDiags := data.ToCreateWebhookDefinitionReq(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreateWebhookDefinition(ctx, &request, params)

	tflog.Info(ctx, "webhook.create", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *contentfulManagement.WebhookDefinition:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		resp.Diagnostics.AddError("Failed to create webhook", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//nolint:dupl
func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WebhookModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := contentfulManagement.GetWebhookDefinitionParams{
		SpaceID:             data.SpaceId.ValueString(),
		WebhookDefinitionID: data.WebhookId.ValueString(),
	}

	response, err := r.providerData.client.GetWebhookDefinition(ctx, params)

	tflog.Info(ctx, "webhook.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	currentVersion := 0

	switch response := response.(type) {
	case *contentfulManagement.WebhookDefinition:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		if response, ok := response.(*contentfulManagement.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read webhook", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read webhook", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WebhookModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var currentVersion int
	currentVersionDiags := GetPrivateProviderData(ctx, req.Private, "version", &currentVersion)
	resp.Diagnostics.Append(currentVersionDiags...)

	params := contentfulManagement.UpdateWebhookDefinitionParams{
		SpaceID:             data.SpaceId.ValueString(),
		WebhookDefinitionID: data.WebhookId.ValueString(),
		XContentfulVersion:  currentVersion,
	}

	request, requestDiags := data.ToUpdateWebhookDefinitionReq(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.UpdateWebhookDefinition(ctx, &request, params)

	tflog.Info(ctx, "webhook.update", map[string]interface{}{
		"params":   params,
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *contentfulManagement.WebhookDefinition:
		currentVersion = response.Sys.Version
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		resp.Diagnostics.AddError("Failed to update webhook", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(SetPrivateProviderData(ctx, resp.Private, "version", currentVersion)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WebhookModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.DeleteWebhookDefinition(ctx, contentfulManagement.DeleteWebhookDefinitionParams{
		SpaceID:             data.SpaceId.ValueString(),
		WebhookDefinitionID: data.WebhookId.ValueString(),
	})

	switch response := response.(type) {
	case *contentfulManagement.NoContent:

	default:
		handled := false

		if response, ok := response.(*contentfulManagement.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Webhook already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete webhook", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
