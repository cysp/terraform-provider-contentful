package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*appSigningSecretResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*appSigningSecretResource)(nil)
	_ resource.ResourceWithConfigure   = (*appSigningSecretResource)(nil)
	_ resource.ResourceWithIdentity    = (*appSigningSecretResource)(nil)
	_ resource.ResourceWithImportState = (*appSigningSecretResource)(nil)
)

//nolint:ireturn
func NewAppSigningSecretResource() resource.Resource {
	return &appSigningSecretResource{}
}

type appSigningSecretResource struct {
	providerData ContentfulProviderData
}

func appSigningSecretIdentityAttributeNames() []string {
	return []string{"organization_id", "app_definition_id"}
}

func (r *appSigningSecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_signing_secret"
}

func (r *appSigningSecretResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AppSigningSecretResourceSchema(ctx)
}

func (r *appSigningSecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *appSigningSecretResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = resourceIdentitySchema(appSigningSecretIdentityAttributeNames())
}

func (r *appSigningSecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, appSigningSecretIdentityAttributeNames(), req, resp)
}

//nolint:dupl // Keep resource lifecycle and HTTP policies explicit.
func (r *appSigningSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppSigningSecretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("value_wo"), &plan.ValueWO)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := plan.ToAppSigningSecretRequest(ctx, path.Empty())
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

	tflog.Info(ctx, "app_signing_secret.create")
	// Create always writes, including replacement and recreation after absence.
	data, putDiags := r.put(ctx, plan, request, types.StringNull())
	resp.Diagnostics.Append(putDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(publishSigningSecretState(ctx, resp.Private, &hashes, resp.Identity, &resp.State, appSigningSecretIdentityAttributeNames(), &data)...)
}

func (r *appSigningSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppSigningSecretModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = maskSigningSecretValues(ctx, state.Value)

	ctx, cancel, timeoutDiagnostics := resourceReadContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	params := cm.GetAppSigningSecretParams{
		OrganizationID:  state.OrganizationID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.GetAppSigningSecret(ctx, params)

	tflog.Info(ctx, "app_signing_secret.read", map[string]any{
		"params": params,
		// "response": response, omitted to avoid logging sensitive values
		"err": appSigningSecretLogError(err, state.Value),
	})

	var data AppSigningSecretModel

	switch response := response.(type) {
	case *cm.AppSigningSecret:
		readState, readDiags := NewAppSigningSecretResourceModelFromResponse(*response, params.OrganizationID, params.AppDefinitionID, state.Value)
		resp.Diagnostics.Append(readDiags...)

		data = readState

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("Failed to read app signing secret", signingSecretErrorDetail(response, err, state.Value))
			// Read must return identity even when Terraform did not send a prior identity.
			resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appSigningSecretIdentityAttributeNames(), &state)...)

			if resp.Diagnostics.HasError() {
				return
			}

			resp.State.RemoveResource(ctx)

			return
		}

		resp.Diagnostics.AddError("Failed to read app signing secret", signingSecretErrorDetail(response, err, state.Value))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appSigningSecretIdentityAttributeNames(), &data)...)
}

//nolint:dupl // Keep resource lifecycle and HTTP policies explicit.
func (r *appSigningSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan AppSigningSecretModel
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
		resp.Diagnostics.Append(publishSigningSecretState(ctx, resp.Private, &hashes, resp.Identity, &resp.State, appSigningSecretIdentityAttributeNames(), &state)...)

		return
	}

	request, requestDiags := plan.ToAppSigningSecretRequest(ctx, path.Empty())
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
		resp.Diagnostics.Append(publishSigningSecretState(ctx, resp.Private, &hashes, resp.Identity, &resp.State, appSigningSecretIdentityAttributeNames(), &state)...)

		return
	}

	tflog.Info(ctx, "app_signing_secret.update")
	data, putDiags := r.put(ctx, plan, request, state.Value)
	resp.Diagnostics.Append(putDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(publishSigningSecretState(ctx, resp.Private, &hashes, resp.Identity, &resp.State, appSigningSecretIdentityAttributeNames(), &data)...)
}

func appSigningSecretLogError(err error, values ...types.String) any {
	if err == nil {
		return nil
	}

	return redactSigningSecretValues(err.Error(), values...)
}

func (r *appSigningSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppSigningSecretModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = maskSigningSecretValues(ctx, state.Value)

	ctx, cancel, timeoutDiagnostics := resourceDeleteContext(ctx, state.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	params := cm.DeleteAppSigningSecretParams{
		OrganizationID:  state.OrganizationID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
	}

	response, err := r.providerData.client.DeleteAppSigningSecret(withContentfulRequestNoRedirect(ctx), params)

	tflog.Info(ctx, "app_signing_secret.delete", map[string]any{
		"params": params,
		// "response": response, omitted to avoid logging sensitive values
		"err": appSigningSecretLogError(err, state.Value),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		if contentfulResponseIsNotFound(response) {
			resp.Diagnostics.AddWarning("App signing secret already deleted", signingSecretErrorDetail(response, err, state.Value))

			return
		}

		resp.Diagnostics.AddError("Failed to delete app signing secret", signingSecretErrorDetail(response, err, state.Value))
	}
}

func (r *appSigningSecretResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	modifySigningSecretPlan(ctx, req, resp)
}

const appSigningSecretMutationRecovery = " The operation may have reached Contentful. Reads cannot verify the complete secret. Returning to a previously acknowledged value may skip a write and does not guarantee restoring that value in Contentful. Coordinate with the app backend before retrying or deliberately rotating the secret."

func (r *appSigningSecretResource) put(ctx context.Context, plan AppSigningSecretModel, request cm.AppSigningSecretRequestData, priorValue types.String) (AppSigningSecretModel, diag.Diagnostics) {
	ctx = maskSigningSecretValues(ctx, priorValue, plan.Value, plan.ValueWO)

	var diags diag.Diagnostics

	organizationID, organizationDiags := requestRequiredString(plan.OrganizationID, path.Root("organization_id"))
	appDefinitionID, appDiags := requestRequiredString(plan.AppDefinitionID, path.Root("app_definition_id"))

	diags.Append(organizationDiags...)
	diags.Append(appDiags...)

	if diags.HasError() {
		return AppSigningSecretModel{}, diags
	}

	params := cm.PutAppSigningSecretParams{OrganizationID: organizationID, AppDefinitionID: appDefinitionID}
	response, err := r.providerData.client.PutAppSigningSecret(withContentfulRequestNoRedirect(ctx), &request, params)
	tflog.Info(ctx, "app_signing_secret.write", map[string]any{
		"params": params,
		"err":    appSigningSecretLogError(err, priorValue, plan.Value, plan.ValueWO),
	})

	secret, ok := response.(*cm.AppSigningSecretStatusCode)
	if err != nil || !ok {
		diags.AddError("Failed to write app signing secret", signingSecretErrorDetail(response, err, priorValue, plan.Value, plan.ValueWO)+appSigningSecretMutationRecovery)

		return AppSigningSecretModel{}, diags
	}

	data, responseDiags := NewAppSigningSecretResourceModelFromResponse(secret.Response, organizationID, appDefinitionID, plan.Value)
	diags.Append(responseDiags...)

	if diags.HasError() {
		diags.AddError("Unconfirmed app signing secret write", appSigningSecretMutationRecovery)

		return AppSigningSecretModel{}, diags
	}

	data.Timeouts = plan.Timeouts

	return data, diags
}
