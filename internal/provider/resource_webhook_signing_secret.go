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

func (r *webhookSigningSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookSigningSecretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diags := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	tflog.Info(ctx, "webhook_signing_secret.create")
	data, diags := r.put(ctx, plan, types.StringNull())
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, webhookSigningSecretIdentityAttributeNames(), &data)...)
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

func (r *webhookSigningSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan WebhookSigningSecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Imported null and previously applied values both survive timeout-only
	// changes. Config is not an authority to override the effective Plan.
	if plan.Value.Equal(state.Value) {
		state.Timeouts = plan.Timeouts
		resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, webhookSigningSecretIdentityAttributeNames(), &state)...)

		return
	}

	ctx, cancel, diags := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	tflog.Info(ctx, "webhook_signing_secret.update")
	data, diags := r.put(ctx, plan, state.Value)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, webhookSigningSecretIdentityAttributeNames(), &data)...)
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

func (r *webhookSigningSecretResource) put(ctx context.Context, plan WebhookSigningSecretModel, priorValue types.String) (WebhookSigningSecretModel, diag.Diagnostics) {
	spaceID, diags := webhookSigningSecretSpaceID(plan.SpaceID)
	request, requestDiags := plan.ToWebhookSigningSecretRequest(ctx, path.Empty())
	diags.Append(requestDiags...)

	if diags.HasError() {
		return WebhookSigningSecretModel{}, diags
	}

	response, err := r.providerData.client.PutWebhookSigningSecret(withContentfulRequestNoRetry(ctx), &request, cm.PutWebhookSigningSecretParams{SpaceID: spaceID})

	var secret cm.WebhookSigningSecret

	switch response := response.(type) {
	case *cm.PutWebhookSigningSecretOK:
		secret = cm.WebhookSigningSecret(*response)
	case *cm.PutWebhookSigningSecretCreated:
		secret = cm.WebhookSigningSecret(*response)
	default:
		diags.AddError("Failed to write webhook signing secret", signingSecretErrorDetail(response, err, priorValue, plan.Value)+webhookSigningSecretMutationRecovery)

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
