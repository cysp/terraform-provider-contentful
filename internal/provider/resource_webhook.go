package provider

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	ctx, cancel, timeoutDiagnostics := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	store := writeOnlySecretVerifiers{}
	mask := webhookMutationRedaction(store, config)
	requestModel, _, prepareDiags := prepareWebhookWrite(ctx, WebhookModel{}, plan, config, &store)
	resp.Diagnostics.Append(prepareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validatePlannedWebhookHeaderValues(path.Root("headers"), requestModel.Headers.Elements(), nil, true)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := requestModel.ToWebhookDefinitionData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.CreateWebhookDefinitionParams{
		SpaceID: plan.SpaceID.ValueString(),
	}

	response, err := r.providerData.client.CreateWebhookDefinition(ctx, &request, params)

	tflog.Info(ctx, "webhook.create", map[string]any{"params": params})

	definition, ok := response.(*cm.WebhookDefinitionStatusCode)
	if err != nil || !ok {
		resp.Diagnostics.AddError("Failed to create webhook", webhookMutationErrorDetail(response, err, mask.any()))

		return
	}

	resp.Diagnostics.Append(publishWebhookResponse(ctx, resp.Private, &store, resp.Identity, &resp.State, definition.Response, WebhookModel{}, plan, mask)...)
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

	tflog.Info(ctx, "webhook.read", map[string]any{"params": params})

	version := 0

	var data WebhookModel

	switch response := response.(type) {
	case *cm.WebhookDefinition:
		sanitized, secretDiags := webhookReadRedaction(state).redact(*response, state)
		resp.Diagnostics.Append(secretDiags...)

		if resp.Diagnostics.HasError() {
			return
		}

		responseModel, responseModelDiags := NewWebhookResourceModelFromResponse(ctx, sanitized, state)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		version = response.Sys.Version

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read webhook", webhookReadErrorDetail(response, err, state))
			removeResourceFromStateWithIdentity(ctx, &resp.State, resp.Identity, &resp.Diagnostics)

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
		state  WebhookModel
	)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, timeoutDiagnostics := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	unchanged, compareDiags := webhookRemoteValuesUnchanged(ctx, req.Plan, req.State)
	resp.Diagnostics.Append(compareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if unchanged && config.HTTPBasicPasswordWO.IsNull() && !config.Headers.IsUnknown() && len(config.writeOnlyHeaderValues()) == 0 {
		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, webhookIdentityAttributeNames(), &state)...)

		return
	}

	store, hashDiags := readWebhookVerifiers(ctx, req.Private, state)
	resp.Diagnostics.Append(hashDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	mask := webhookMutationRedaction(store, config)
	requestModel, needsPUT, prepareDiags := prepareWebhookWrite(ctx, state, plan, config, &store)
	resp.Diagnostics.Append(prepareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !needsPUT {
		// A storage-only migration changes the Plan and verifier without changing
		// Contentful. Computed fields resolve from the acknowledged prior state.
		state.HTTPBasicPassword = plan.HTTPBasicPassword
		state.Headers = plan.Headers
		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(publishStateAfterPrivate(ctx, resp.Identity, &resp.State, webhookIdentityAttributeNames(), &state, func() diag.Diagnostics { return store.publish(ctx, resp.Private) })...)

		return
	}

	request, requestDiags := requestModel.ToWebhookDefinitionData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	version, versionDiags := webhookUpdateVersion(ctx, req.Private, plan.Version, unchanged)
	resp.Diagnostics.Append(versionDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.UpdateWebhookDefinitionParams{
		SpaceID:             plan.SpaceID.ValueString(),
		WebhookDefinitionID: plan.WebhookID.ValueString(),
		XContentfulVersion:  cm.NewOptInt(version),
	}

	response, err := r.providerData.client.UpdateWebhookDefinition(ctx, &request, params)

	tflog.Info(ctx, "webhook.update", map[string]any{"params": params})

	definition, ok := response.(*cm.WebhookDefinitionStatusCode)
	if err != nil || !ok {
		resp.Diagnostics.AddError("Failed to update webhook", webhookMutationErrorDetail(response, err, mask.any() || webhookReadRedaction(state).any()))

		return
	}

	resp.Diagnostics.Append(publishWebhookResponse(ctx, resp.Private, &store, resp.Identity, &resp.State, definition.Response, state, plan, mask)...)
}

// A saved plan can pin sys.version before an ephemeral credential first
// appears at apply. Only send a versioned PUT that the effective plan allows.
func webhookUpdateVersion(ctx context.Context, private PrivateProviderData, planned types.Int64, unchanged bool) (int, diag.Diagnostics) {
	if unchanged && !planned.IsUnknown() {
		return 0, diag.Diagnostics{diag.NewErrorDiagnostic("Webhook rotation requires a new plan", "The saved plan retained the observed webhook version. Plan again with the intended write-only credential before applying.")}
	}

	return requiredPrivateVersion(ctx, private)
}

func webhookRemoteValuesUnchanged(ctx context.Context, plan tfsdk.Plan, state tfsdk.State) (bool, diag.Diagnostics) {
	var planned, prior types.Object

	diags := plan.Get(ctx, &planned)
	diags.Append(state.Get(ctx, &prior)...)

	if diags.HasError() {
		return false, diags
	}

	plannedValues, priorValues := planned.Attributes(), prior.Attributes()

	for name, attribute := range plan.Schema.GetAttributes() {
		if name == "timeouts" {
			continue
		}

		value := plannedValues[name]
		// A computed-only observation can retain its prior value if no write is needed.
		if attribute.IsComputed() && !attribute.IsOptional() && value.IsUnknown() {
			continue
		}

		if !value.Equal(priorValues[name]) {
			return false, diags
		}
	}

	return true, diags
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

	tflog.Info(ctx, "webhook.delete", map[string]any{"params": params})

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
	if state.Name.IsNull() || webhookReadRedaction(state).any() {
		return webhookResponseStatus(response) + "Contentful did not confirm the webhook operation."
	}

	return util.ErrorDetailFromContentfulManagementResponse(response, err)
}
