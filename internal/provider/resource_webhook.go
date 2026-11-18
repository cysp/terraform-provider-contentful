package provider

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                   = (*webhookResource)(nil)
	_ resource.ResourceWithModifyPlan     = (*webhookResource)(nil)
	_ resource.ResourceWithValidateConfig = (*webhookResource)(nil)
	_ resource.ResourceWithConfigure      = (*webhookResource)(nil)
	_ resource.ResourceWithIdentity       = (*webhookResource)(nil)
	_ resource.ResourceWithImportState    = (*webhookResource)(nil)
)

//nolint:ireturn
func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

type webhookResource struct {
	providerData ContentfulProviderData
}

func webhookIdentityAttributeNames() []string { return []string{"space_id", "webhook_id"} }

func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = WebhookResourceSchema(ctx)
}

func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *webhookResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(webhookIdentityAttributeNames())
}

func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, webhookIdentityAttributeNames(), req, resp)
}

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var (
		plan   WebhookModel
		config WebhookModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToWebhookDefinitionData(ctx, config, path.Empty())
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

	store := writeOnlySecretVerifiers{}
	_, prepareDiags := prepareWebhookWrite(ctx, &request, WebhookModel{}, plan, config, &store)
	resp.Diagnostics.Append(prepareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.CreateWebhookDefinitionParams{
		SpaceID: plan.SpaceID.ValueString(),
	}

	response, err := r.providerData.client.CreateWebhookDefinition(ctx, &request, params)

	tflog.Info(ctx, "webhook.create")

	definition, ok := response.(*cm.WebhookDefinitionStatusCode)
	if err != nil || !ok {
		resp.Diagnostics.AddError("Failed to create webhook", webhookMutationErrorDetail(response, err, len(store.records) != 0))

		return
	}

	resp.Diagnostics.Append(publishWebhookResponse(ctx, resp.Private, &store, resp.Identity, &resp.State, definition.Response, WebhookModel{}, plan, config, len(store.records) != 0)...)
}

func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookModel

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

	params := cm.GetWebhookDefinitionParams{
		SpaceID:             state.SpaceID.ValueString(),
		WebhookDefinitionID: state.WebhookID.ValueString(),
	}

	response, err := r.providerData.client.GetWebhookDefinition(ctx, params)

	tflog.Info(ctx, "webhook.read")

	version := 0

	var data WebhookModel

	switch response := response.(type) {
	case *cm.WebhookDefinition:
		sanitized, secretDiags := redactWebhookResponse(*response, state, state, WebhookModel{})
		resp.Diagnostics.Append(secretDiags...)

		if resp.Diagnostics.HasError() {
			return
		}

		responseModel, responseModelDiags := NewWebhookResourceModelFromResponse(ctx, sanitized, state.Headers.Elements())
		resp.Diagnostics.Append(responseModelDiags...)

		if response.HttpBasicPassword.IsEmpty() && !state.HTTPBasicPassword.IsNull() && !state.HTTPBasicPassword.IsUnknown() {
			responseModel.HTTPBasicPassword = state.HTTPBasicPassword
		}

		data = responseModel
		version = response.Sys.Version

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read webhook", webhookReadErrorDetail(response, err, state))
			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read webhook", webhookReadErrorDetail(response, err, state))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(validateResourceReadIdentity(ctx, req.State, webhookIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	store, pruneDiags := webhookVerifiersAfterRead(ctx, req.Private, state, data)
	resp.Diagnostics.Append(pruneDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(publishWebhookState(ctx, resp.Private, &store, resp.Identity, &resp.State, data, version, true)...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var (
		plan   WebhookModel
		config WebhookModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToWebhookDefinitionData(ctx, config, path.Empty())
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

	var state WebhookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	store, hashDiags := readWebhookVerifiers(ctx, req.Private, state)
	resp.Diagnostics.Append(hashDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	protected := len(store.records) != 0
	needsPUT, prepareDiags := prepareWebhookWrite(ctx, &request, state, plan, config, &store)
	resp.Diagnostics.Append(prepareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	version, versionDiags := requiredPrivateVersion(ctx, req.Private)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	protected = protected || len(store.records) != 0

	if !needsPUT {
		// A storage-only migration changes the Plan and verifier without changing
		// Contentful. Computed fields resolve from the acknowledged prior state.
		state.HTTPBasicPassword = plan.HTTPBasicPassword
		state.Headers = plan.Headers
		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(publishWebhookState(ctx, resp.Private, &store, resp.Identity, &resp.State, state, version, true)...)

		return
	}

	params := cm.UpdateWebhookDefinitionParams{
		SpaceID:             plan.SpaceID.ValueString(),
		WebhookDefinitionID: plan.WebhookID.ValueString(),
		XContentfulVersion:  version,
	}

	response, err := r.providerData.client.UpdateWebhookDefinition(ctx, &request, params)

	tflog.Info(ctx, "webhook.update")

	definition, ok := response.(*cm.WebhookDefinitionStatusCode)
	if err != nil || !ok {
		resp.Diagnostics.AddError("Failed to update webhook", webhookMutationErrorDetail(response, err, protected))

		return
	}

	resp.Diagnostics.Append(publishWebhookResponse(ctx, resp.Private, &store, resp.Identity, &resp.State, definition.Response, state, plan, config, protected)...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookModel

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

	params := cm.DeleteWebhookDefinitionParams{
		SpaceID:             state.SpaceID.ValueString(),
		WebhookDefinitionID: state.WebhookID.ValueString(),
	}

	response, err := r.providerData.client.DeleteWebhookDefinition(ctx, params)

	tflog.Info(ctx, "webhook.delete")

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Webhook already deleted", webhookReadErrorDetail(response, err, state))

			return
		}

		resp.Diagnostics.AddError("Failed to delete webhook", webhookReadErrorDetail(response, err, state))
	}
}

func webhookMutationErrorDetail(response any, err error, protected bool) string {
	if protected {
		return webhookResponseStatus(response) + "Contentful did not acknowledge the webhook write. The operation may have reached Contentful. Reads cannot confirm secret bytes. Coordinate recovery with the webhook receiver before applying again."
	}

	return util.ErrorDetailFromContentfulManagementResponse(response, err)
}

func webhookResponseStatus(response any) string {
	if result, ok := response.(cm.StatusCodeResponse); ok {
		return fmt.Sprintf("HTTP %d. ", result.GetStatusCode())
	}

	return ""
}

func webhookReadErrorDetail(response any, err error, state WebhookModel) string {
	if state.Name.IsNull() || state.hasRedactedCredentials() {
		return webhookResponseStatus(response) + "Contentful did not confirm the webhook operation."
	}

	return util.ErrorDetailFromContentfulManagementResponse(response, err)
}
