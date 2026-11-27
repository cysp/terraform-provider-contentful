package provider

import (
	"context"
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/text/unicode/norm"
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
	defer omitWebhookCredentialSource(&resp.Diagnostics)

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

	store := webhookVerifiers{}
	mask := webhookMutationRedaction(store, config)
	requestModel, _, prepareDiags := prepareWebhookWrite(ctx, WebhookModel{}, plan, config, &store)
	resp.Diagnostics.Append(prepareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := requestModel.toWebhookDefinitionCreateData(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	change, encodeDiags := store.encode()
	resp.Diagnostics.Append(encodeDiags...)

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
		resp.Diagnostics.AddError("Failed to create webhook", webhookOperationErrorDetail(http.MethodPost, response))

		return
	}

	if definition.Response.Sys.ID == "" {
		resp.Diagnostics.AddError("Missing created webhook ID", "Contentful returned an empty ID. The provider could not identify the created webhook or acknowledge its credentials. Inspect webhooks in the space before retrying.")

		return
	}

	if definition.Response.Sys.Space.Sys.ID != params.SpaceID {
		resp.Diagnostics.AddError("Unexpected space in created webhook", "Contentful returned a webhook from a different space. Inspect the requested space and import the created webhook if present before retrying.")

		return
	}

	_, sentPassword := request.HttpBasicPassword.Get()
	resp.Diagnostics.Append(publishWebhookResponse(ctx, resp.Private, webhookVerifiers{}, change, resp.Identity, &resp.State, definition.Response, WebhookModel{}, plan, mask, sentPassword)...)
}

func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer omitWebhookCredentialSource(&resp.Diagnostics)

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

	response, _ := r.providerData.client.GetWebhookDefinition(ctx, params)

	tflog.Info(ctx, "webhook.read", map[string]any{"params": params})

	version := 0

	var passwordNull bool

	var data WebhookModel

	switch response := response.(type) {
	case *cm.WebhookDefinition:
		passwordNull = response.HttpBasicPassword.IsNull()
		sanitized, secretDiags := webhookReadRedaction(state).redact(*response, state)
		resp.Diagnostics.Append(secretDiags...)

		if resp.Diagnostics.HasError() {
			return
		}

		responseModel, responseModelDiags := NewWebhookResourceModelFromResponse(ctx, sanitized, state)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
		if response.HttpBasicUsername.IsEmpty() && response.HttpBasicPassword.IsEmpty() {
			data.HTTPBasicUsername = state.HTTPBasicUsername
			data.HTTPBasicPassword = state.HTTPBasicPassword
		}

		version = response.Sys.Version

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read webhook", "HTTP 404. Contentful reported that the webhook does not exist.")
			removeResourceFromStateWithIdentity(ctx, &resp.State, resp.Identity, &resp.Diagnostics)

			return
		}

		resp.Diagnostics.AddError("Failed to read webhook", webhookOperationErrorDetail(http.MethodGet, response))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(validateResourceReadIdentity(ctx, req.State, webhookIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	change, pruneDiags := webhookVerifiersAfterRead(ctx, req.Private, state, data, passwordNull)
	resp.Diagnostics.Append(pruneDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(publishWebhookState(ctx, resp.Private, change, resp.Identity, &resp.State, data, version, true)...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer omitWebhookCredentialSource(&resp.Diagnostics)

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

	priorVerifiers, hashDiags := readWebhookVerifiers(ctx, req.Private, state)
	resp.Diagnostics.Append(hashDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	preparedVerifiers := priorVerifiers
	mask := webhookMutationRedaction(priorVerifiers, config)
	requestModel, needsPUT, prepareDiags := prepareWebhookWrite(ctx, state, plan, config, &preparedVerifiers)
	resp.Diagnostics.Append(prepareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	change, encodeDiags := preparedVerifiers.encode()
	resp.Diagnostics.Append(encodeDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !needsPUT {
		// A storage-only migration changes the Plan and verifier without changing
		// Contentful. Computed fields resolve from the acknowledged prior state.
		state.HTTPBasicPassword = plan.HTTPBasicPassword
		state.Headers = plan.Headers
		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(publishStateAfterPrivate(ctx, resp.Identity, &resp.State, webhookIdentityAttributeNames(), &state, func() diag.Diagnostics { return change.publish(ctx, resp.Private) })...)

		return
	}

	request, requestDiags := requestModel.toWebhookDefinitionData(ctx, path.Empty(), state.HTTPBasicUsername)
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
		resp.Diagnostics.AddError("Failed to update webhook", webhookOperationErrorDetail(http.MethodPut, response))

		return
	}

	if !plan.identifiesWebhook(definition.Response.Sys) {
		state.Version = types.Int64Null()
		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(setResourceIdentityStateAndMutationVersion(ctx, resp.Identity, &resp.State, resp.Private, webhookIdentityAttributeNames(), &state, definition.Response.Sys.Version, false)...)
		resp.Diagnostics.AddError("Contentful returned a different webhook identity", "The PUT response identifies another webhook. Terraform retained the target’s prior settings and credentials and cleared its version. Refresh the target before applying again.")

		return
	}

	mask.guardDelegatedHeaders(request.Headers)

	responsePlan := plan

	responsePlan.HTTPBasicUsername = expectedWebhookResponseUsername(plan.HTTPBasicUsername, request.HttpBasicUsername, definition.Response.HttpBasicUsername)
	if preparedVerifiers.discardMismatchedBasicVerifier(responsePlan.HTTPBasicUsername) {
		change, encodeDiags = preparedVerifiers.encode()
		resp.Diagnostics.Append(encodeDiags...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	_, sentPassword := request.HttpBasicPassword.Get()
	resp.Diagnostics.Append(publishWebhookResponse(ctx, resp.Private, priorVerifiers, change, resp.Identity, &resp.State, definition.Response, state, responsePlan, mask, sentPassword)...)
}

// Core stores strings in NFC. A delegated PUT sent no username, so its
// response may use the original spelling. Verifiers still bind to that spelling.
func expectedWebhookResponseUsername(planned types.String, sent, returned cm.OptNilString) types.String {
	if username, ok := returned.Get(); sent.IsEmpty() && ok && norm.NFC.String(username) == planned.ValueString() {
		return types.StringValue(username)
	}

	return planned
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

	response, _ := r.providerData.client.DeleteWebhookDefinition(ctx, params)

	tflog.Info(ctx, "webhook.delete", map[string]any{"params": params})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Webhook already deleted", "HTTP 404. Contentful reported that the webhook does not exist.")

			return
		}

		resp.Diagnostics.AddError("Failed to delete webhook", webhookOperationErrorDetail(http.MethodDelete, response))
	}
}

// Free-form response bodies and decoder errors can include remote credential bytes,
// including credentials that have never appeared in Terraform state.
func webhookOperationErrorDetail(method string, response any) string {
	if _, ok := response.(*cm.ApplicationJSONError); ok {
		// The bare error union represents 401 on all four Webhook operations.
		return "HTTP 401. Contentful rejected authentication. Check the provider's Contentful management access token."
	}

	if failure, ok := response.(*cm.ErrorStatusCode); ok {
		if cmaError, decoded := failure.GetError(); decoded {
			switch {
			case method == http.MethodPut && failure.StatusCode == http.StatusConflict && cmaError.Sys.ID == cm.ErrorSysIDConflict:
				return "HTTP 409. Contentful reported a conflict on the versioned webhook update. Create a fresh Terraform plan and review any remote changes before applying again."
			case (method == http.MethodPost || method == http.MethodPut) && failure.StatusCode == http.StatusUnprocessableEntity && cmaError.Sys.ID == cm.ErrorSysIDValidationFailed:
				return "HTTP 422. Contentful rejected the webhook definition. Review the configuration before applying again."
			}
		}
	}

	detail := "Contentful did not confirm the webhook operation."

	switch method {
	case http.MethodPost, http.MethodPut:
		detail = "Contentful did not acknowledge the webhook write. The operation may have reached Contentful. Reads cannot confirm secret bytes. Coordinate recovery with the webhook receiver before applying again."
	case http.MethodDelete:
		detail = "Check whether the webhook still exists before retrying deletion."
	}

	if status, ok := response.(cm.StatusCodeResponse); ok {
		return fmt.Sprintf("HTTP %d. %s", status.GetStatusCode(), detail)
	}

	return detail
}
