package provider

import (
	"context"
	"net/http"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*appSigningSecretResource)(nil)
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

func (r *appSigningSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppSigningSecretModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = maskAppSigningSecretValues(ctx, plan.Value)

	ctx, cancel, timeoutDiagnostics := resourceCreateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	params := cm.PutAppSigningSecretParams{
		OrganizationID:  plan.OrganizationID.ValueString(),
		AppDefinitionID: plan.AppDefinitionID.ValueString(),
	}

	request, requestDiags := plan.ToAppSigningSecretRequest(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppSigningSecret(ctx, &request, params)

	tflog.Info(ctx, "app_signing_secret.create", map[string]any{
		"params": params,
		// "request": request, omitted to avoid logging sensitive values
		// "response": response, omitted to avoid logging sensitive values
		"err": appSigningSecretLogError(err, plan.Value),
	})

	var data AppSigningSecretModel

	switch response := response.(type) {
	case *cm.AppSigningSecretStatusCode:
		mutationState, mutationStateDiags := NewAppSigningSecretResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(mutationStateDiags...)

		if resp.Diagnostics.HasError() {
			return
		}

		if mutationState.Value.IsNull() && !plan.Value.IsUnknown() {
			mutationState.Value = plan.Value
		}

		data = mutationState

	default:
		resp.Diagnostics.AddError("Failed to create app signing secret", appSigningSecretErrorDetail(response, err, plan.Value))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appSigningSecretIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *appSigningSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppSigningSecretModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = maskAppSigningSecretValues(ctx, state.Value)

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
		readState, readStateDiags := NewAppSigningSecretResourceModelFromResponse(ctx, *response)
		resp.Diagnostics.Append(readStateDiags...)

		if resp.Diagnostics.HasError() {
			return
		}

		if readState.Value.IsNull() && !state.Value.IsUnknown() {
			readState.Value = state.Value
		}

		data = readState

	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read app signing secret", appSigningSecretErrorDetail(response, err, state.Value))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read app signing secret", appSigningSecretErrorDetail(response, err, state.Value))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = state.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appSigningSecretIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *appSigningSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan AppSigningSecretModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = maskAppSigningSecretValues(ctx, state.Value, plan.Value)

	ctx, cancel, timeoutDiagnostics := resourceUpdateContext(ctx, plan.Timeouts)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	defer cancel()

	params := cm.PutAppSigningSecretParams{
		OrganizationID:  plan.OrganizationID.ValueString(),
		AppDefinitionID: plan.AppDefinitionID.ValueString(),
	}

	request, requestDiags := plan.ToAppSigningSecretRequest(ctx, path.Empty())
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.PutAppSigningSecret(ctx, &request, params)

	tflog.Info(ctx, "app_signing_secret.update", map[string]any{
		"params": params,
		// "request": request, omitted to avoid logging sensitive values
		// "response": response, omitted to avoid logging sensitive values
		"err": appSigningSecretLogError(err, state.Value, plan.Value),
	})

	var data AppSigningSecretModel

	switch response := response.(type) {
	case *cm.AppSigningSecretStatusCode:
		mutationState, mutationStateDiags := NewAppSigningSecretResourceModelFromResponse(ctx, response.Response)
		resp.Diagnostics.Append(mutationStateDiags...)

		if resp.Diagnostics.HasError() {
			return
		}

		if mutationState.Value.IsNull() {
			if !plan.Value.IsUnknown() {
				mutationState.Value = plan.Value
			} else if !state.Value.IsUnknown() {
				mutationState.Value = state.Value
			}
		}

		data = mutationState

	default:
		resp.Diagnostics.AddError("Failed to update app signing secret", appSigningSecretErrorDetail(response, err, state.Value, plan.Value))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Timeouts = plan.Timeouts

	resp.Diagnostics.Append(setResourceIdentityAndState(ctx, resp.Identity, &resp.State, appSigningSecretIdentityAttributeNames(), &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func maskAppSigningSecretValues(ctx context.Context, values ...types.String) context.Context {
	knownValues := make([]string, 0, len(values))

	for _, value := range values {
		if !value.IsNull() && !value.IsUnknown() && value.ValueString() != "" {
			knownValues = append(knownValues, value.ValueString())
		}
	}

	return tflog.MaskLogStrings(ctx, knownValues...)
}

func appSigningSecretErrorDetail(response any, err error, values ...types.String) string {
	return redactAppSigningSecretValues(util.ErrorDetailFromContentfulManagementResponse(response, err), values...)
}

func appSigningSecretLogError(err error, values ...types.String) any {
	if err == nil {
		return nil
	}

	return redactAppSigningSecretValues(err.Error(), values...)
}

func redactAppSigningSecretValues(text string, values ...types.String) string {
	redacted := text

	for _, value := range values {
		if !value.IsNull() && !value.IsUnknown() && value.ValueString() != "" {
			redacted = strings.ReplaceAll(redacted, value.ValueString(), "***")
		}
	}

	return redacted
}

func (r *appSigningSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppSigningSecretModel

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

	response, err := r.providerData.client.DeleteAppSigningSecret(ctx, cm.DeleteAppSigningSecretParams{
		OrganizationID:  state.OrganizationID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
	})

	switch response := response.(type) {
	case *cm.NoContent:

	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("App signing secret already deleted", appSigningSecretErrorDetail(response, err, state.Value))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete app signing secret", appSigningSecretErrorDetail(response, err, state.Value))
		}
	}
}
