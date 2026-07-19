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
	_ resource.Resource                   = (*appKeyResource)(nil)
	_ resource.ResourceWithConfigure      = (*appKeyResource)(nil)
	_ resource.ResourceWithIdentity       = (*appKeyResource)(nil)
	_ resource.ResourceWithImportState    = (*appKeyResource)(nil)
	_ resource.ResourceWithValidateConfig = (*appKeyResource)(nil)
)

//nolint:ireturn
func NewAppKeyResource() resource.Resource {
	return &appKeyResource{}
}

type appKeyResource struct {
	providerData ContentfulProviderData
}

func (r *appKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_key"
}

func (r *appKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = AppKeyResourceSchema(ctx)
}

func (r *appKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *appKeyResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"organization_id":   identityschema.StringAttribute{RequiredForImport: true},
			"app_definition_id": identityschema.StringAttribute{RequiredForImport: true},
			"key_kid":           identityschema.StringAttribute{RequiredForImport: true},
		},
	}
}

func (r *appKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportStatePassthroughMultipartID(ctx, []path.Path{
		path.Root("organization_id"),
		path.Root("app_definition_id"),
		path.Root("key_kid"),
	}, req, resp)
}

func (r *appKeyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config AppKeyModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	jwk, ok := config.JWK.GetValue()
	if !ok {
		return
	}

	resp.Diagnostics.Append(validateKnownAppKeyJWKMaterial(jwk, path.Root("jwk"))...)
}

func (r *appKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AppKeyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := plan.Timeouts.Create(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	params := cm.CreateAppKeyParams{
		OrganizationID:  plan.OrganizationID.ValueString(),
		AppDefinitionID: plan.AppDefinitionID.ValueString(),
	}

	request, requestDiags := plan.ToAppKeyRequestData(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreateAppKey(ctx, &request, params)

	tflog.Info(ctx, "app_key.create", map[string]any{
		"params": params,
		"err":    err,
	})

	responseAppKey, ok := response.(*cm.AppKey)
	if !ok {
		resp.Diagnostics.AddError("Failed to create app key", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data := NewAppKeyResourceModelFromResponse(*responseAppKey)
	data.Timeouts = plan.Timeouts

	var identityModel AppKeyIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *appKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AppKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := state.Timeouts.Read(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout = max(timeout, minimumStoredResourceOperationTimeout)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	params := cm.GetAppKeyParams{
		OrganizationID:  state.OrganizationID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
		KeyKid:          state.KeyKID.ValueString(),
	}

	response, err := r.providerData.client.GetAppKey(ctx, params)

	tflog.Info(ctx, "app_key.read", map[string]any{
		"params": params,
		"err":    err,
	})

	var data AppKeyModel

	switch response := response.(type) {
	case *cm.AppKey:
		data = NewAppKeyResourceModelFromResponse(*response)
	default:
		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read app key", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read app key", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	data.Timeouts = state.Timeouts

	var identityModel AppKeyIdentityModel
	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *appKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan AppKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	state.Timeouts = plan.Timeouts

	var identityModel AppKeyIdentityModel

	resp.Diagnostics.Append(CopyAttributeValues(ctx, &identityModel, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Identity.Set(ctx, &identityModel)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *appKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AppKeyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := state.Timeouts.Delete(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout = max(timeout, minimumStoredResourceOperationTimeout)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	params := cm.DeleteAppKeyParams{
		OrganizationID:  state.OrganizationID.ValueString(),
		AppDefinitionID: state.AppDefinitionID.ValueString(),
		KeyKid:          state.KeyKID.ValueString(),
	}

	response, err := r.providerData.client.DeleteAppKey(ctx, params)

	tflog.Info(ctx, "app_key.delete", map[string]any{
		"params": params,
		"err":    err,
	})

	switch response := response.(type) {
	case *cm.NoContent:
	default:
		handled := false

		if response, ok := response.(cm.StatusCodeResponse); ok {
			if response.GetStatusCode() == http.StatusNotFound {
				resp.Diagnostics.AddWarning("App key already deleted", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to delete app key", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
