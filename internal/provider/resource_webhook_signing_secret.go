package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*webhookSigningSecretResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*webhookSigningSecretResource)(nil)
	_ resource.ResourceWithConfigure   = (*webhookSigningSecretResource)(nil)
	_ resource.ResourceWithIdentity    = (*webhookSigningSecretResource)(nil)
	_ resource.ResourceWithImportState = (*webhookSigningSecretResource)(nil)
)

//nolint:ireturn
func NewWebhookSigningSecretResource() resource.Resource {
	return &webhookSigningSecretResource{}
}

type webhookSigningSecretResource struct {
	providerData ContentfulProviderData
}

func webhookSigningSecretIdentityAttributeNames() []string {
	return []string{"space_id"}
}

func (r *webhookSigningSecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_signing_secret"
}

func (r *webhookSigningSecretResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = WebhookSigningSecretResourceSchema(ctx)
}

func (r *webhookSigningSecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *webhookSigningSecretResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(webhookSigningSecretIdentityAttributeNames())
}

func (r *webhookSigningSecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, webhookSigningSecretIdentityAttributeNames(), req, resp)
}

//nolint:dupl // Keep resource lifecycle and HTTP policies explicit.
func (r *webhookSigningSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookSigningSecretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("value_wo"), &plan.ValueWO)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToWebhookSigningSecretRequest(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = maskSigningSecretValues(ctx, plan.Value, plan.ValueWO)

	ctx, cancel, timeoutDiags := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	hashes := writeOnlySecretHashes{}

	_, valuePath, _ := resolveSigningSecretValue(plan.Value, plan.ValueWO, path.Empty())
	_, prepareDiags := prepareSigningSecretValue(ctx, &hashes, types.StringNull(), types.StringValue(request.Value), valuePath)
	resp.Diagnostics.Append(prepareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "webhook_signing_secret.create")
	// Create always writes, including replacement and recreation after absence.
	data, putDiags := r.put(ctx, plan, request, types.StringNull())
	resp.Diagnostics.Append(putDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(publishSigningSecretState(ctx, resp.Private, &hashes, resp.Identity, &resp.State, webhookSigningSecretIdentityAttributeNames(), &data)...)
}

func (r *webhookSigningSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookSigningSecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	spaceID, diags := webhookSigningSecretSpaceID(state.SpaceID)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diags := resourceReadContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	tflog.Info(ctx, "webhook_signing_secret.read")

	response, err := r.providerData.client.GetWebhookSigningSecret(ctx, cm.GetWebhookSigningSecretParams{SpaceID: spaceID})
	if err == nil && webhookSigningSecretNotFound(response) {
		// Read must return identity even when Terraform did not send a prior identity.
		resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, webhookSigningSecretIdentityAttributeNames(), &state)...)

		if resp.Diagnostics.HasError() {
			return
		}

		resp.State.RemoveResource(ctx)

		return
	}

	secret, ok := response.(*cm.WebhookSigningSecret)
	if err != nil || !ok {
		resp.Diagnostics.AddError("Failed to read webhook signing secret", signingSecretErrorDetail(response, err, state.Value))

		return
	}

	data, diags := NewWebhookSigningSecretResourceModelFromResponse(*secret, spaceID, state.Value)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, webhookSigningSecretIdentityAttributeNames(), &data)...)
}

//nolint:dupl // Keep resource lifecycle and HTTP policies explicit.
func (r *webhookSigningSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan WebhookSigningSecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("value_wo"), &plan.ValueWO)...)

	if resp.Diagnostics.HasError() {
		return
	}

	hashes, hashDiags := readSigningSecretHashes(ctx, req.Private)
	resp.Diagnostics.Append(hashDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Ordinary fields retain their effective values in Plan, including ignores.
	if plan.ValueWO.IsNull() && !plan.Value.IsUnknown() && plan.Value.Equal(state.Value) {
		if !plan.Value.IsNull() {
			resp.Diagnostics.Append(hashes.remove(path.Root("value_wo"))...)
		}

		if resp.Diagnostics.HasError() {
			return
		}

		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(publishSigningSecretState(ctx, resp.Private, &hashes, resp.Identity, &resp.State, webhookSigningSecretIdentityAttributeNames(), &state)...)

		return
	}

	request, requestDiags := plan.ToWebhookSigningSecretRequest(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = maskSigningSecretValues(ctx, state.Value, plan.Value, plan.ValueWO)

	ctx, cancel, timeoutDiags := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	_, valuePath, _ := resolveSigningSecretValue(plan.Value, plan.ValueWO, path.Empty())
	matches, prepareDiags := prepareSigningSecretValue(ctx, &hashes, state.Value, types.StringValue(request.Value), valuePath)
	resp.Diagnostics.Append(prepareDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if matches {
		// Same bytes can still change their public/private representation.
		state.Value = plan.Value
		state.ValueWO = types.StringNull()
		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(publishSigningSecretState(ctx, resp.Private, &hashes, resp.Identity, &resp.State, webhookSigningSecretIdentityAttributeNames(), &state)...)

		return
	}

	tflog.Info(ctx, "webhook_signing_secret.update")
	data, putDiags := r.put(ctx, plan, request, state.Value)
	resp.Diagnostics.Append(putDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(publishSigningSecretState(ctx, resp.Private, &hashes, resp.Identity, &resp.State, webhookSigningSecretIdentityAttributeNames(), &data)...)
}

const webhookSigningSecretMutationRecovery = " The operation may have reached Contentful. Reads cannot verify the complete secret. Coordinate with all webhook receivers before retrying; a later apply can overwrite or delete intervening changes."

func (r *webhookSigningSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookSigningSecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	spaceID, diags := webhookSigningSecretSpaceID(state.SpaceID)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diags := resourceDeleteContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	tflog.Info(ctx, "webhook_signing_secret.delete")

	response, err := r.providerData.client.DeleteWebhookSigningSecret(withContentfulRequestNoRetry(ctx), cm.DeleteWebhookSigningSecretParams{SpaceID: spaceID})
	if err == nil {
		if _, ok := response.(*cm.NoContent); ok || webhookSigningSecretNotFound(response) {
			return
		}
	}

	resp.Diagnostics.AddError("Failed to delete webhook signing secret", signingSecretErrorDetail(response, err, state.Value)+webhookSigningSecretMutationRecovery)
}

func (r *webhookSigningSecretResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifySigningSecretPlan(ctx, req, resp)
}

func (r *webhookSigningSecretResource) put(ctx context.Context, plan WebhookSigningSecretModel, request cm.WebhookSigningSecretRequestData, priorValue types.String) (WebhookSigningSecretModel, diag.Diagnostics) {
	ctx = maskSigningSecretValues(ctx, priorValue, plan.Value, plan.ValueWO)
	spaceID, diags := webhookSigningSecretSpaceID(plan.SpaceID)

	if diags.HasError() {
		return WebhookSigningSecretModel{}, diags
	}

	response, err := r.providerData.client.PutWebhookSigningSecret(withContentfulRequestNoRetry(ctx), &request, cm.PutWebhookSigningSecretParams{SpaceID: spaceID})
	if err != nil {
		diags.AddError("Failed to write webhook signing secret", signingSecretErrorDetail(response, err, priorValue, plan.Value, plan.ValueWO)+webhookSigningSecretMutationRecovery)

		return WebhookSigningSecretModel{}, diags
	}

	var secret cm.WebhookSigningSecret

	switch response := response.(type) {
	case *cm.PutWebhookSigningSecretOK:
		secret = cm.WebhookSigningSecret(*response)
	case *cm.PutWebhookSigningSecretCreated:
		secret = cm.WebhookSigningSecret(*response)
	default:
		diags.AddError("Failed to write webhook signing secret", signingSecretErrorDetail(response, err, priorValue, plan.Value, plan.ValueWO)+webhookSigningSecretMutationRecovery)

		return WebhookSigningSecretModel{}, diags
	}

	data, responseDiags := NewWebhookSigningSecretResourceModelFromResponse(secret, spaceID, plan.Value)
	diags.Append(responseDiags...)

	if diags.HasError() {
		diags.AddError("Unconfirmed webhook signing secret write", webhookSigningSecretMutationRecovery)

		return WebhookSigningSecretModel{}, diags
	}

	data.Timeouts = plan.Timeouts

	return data, diags
}

func webhookSigningSecretNotFound(response any) bool {
	responseError, ok := response.(*cm.ErrorStatusCode)
	if !ok || responseError.StatusCode != http.StatusNotFound {
		return false
	}

	contentfulError, ok := responseError.Response.GetError()

	return ok && contentfulError.Sys.ID == cm.ErrorSysIDNotFound
}
