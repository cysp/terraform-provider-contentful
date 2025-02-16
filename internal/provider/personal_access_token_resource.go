package provider

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*personalAccessTokenResource)(nil)
	_ resource.ResourceWithConfigure   = (*personalAccessTokenResource)(nil)
	_ resource.ResourceWithImportState = (*personalAccessTokenResource)(nil)
)

//nolint:ireturn
func NewPersonalAccessTokenResource() resource.Resource {
	return &personalAccessTokenResource{}
}

type personalAccessTokenResource struct {
	providerData ContentfulProviderData
}

func (r *personalAccessTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_personal_access_token"
}

func (r *personalAccessTokenResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = PersonalAccessTokenResourceSchema(ctx)
}

func (r *personalAccessTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *personalAccessTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *personalAccessTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PersonalAccessTokenModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := data.ToCreatePersonalAccessTokenReq(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreatePersonalAccessToken(ctx, &request)

	tflog.Info(ctx, "personal_access_token.create", map[string]interface{}{
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.PersonalAccessToken:
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		resp.Diagnostics.AddError("Failed to create personal access token", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *personalAccessTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PersonalAccessTokenModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetPersonalAccessTokenParams{
		AccessTokenID: data.Id.ValueString(),
	}

	response, err := r.providerData.client.GetPersonalAccessToken(ctx, params)

	tflog.Info(ctx, "personal_access_token.read", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.PersonalAccessToken:
		resp.Diagnostics.Append(data.ReadFromResponse(ctx, response)...)

	default:
		if response, ok := response.(*cm.ErrorStatusCode); ok {
			if response.StatusCode == http.StatusNotFound {
				resp.Diagnostics.AddWarning("Failed to read personal access token", util.ErrorDetailFromContentfulManagementResponse(response, err))
				resp.State.RemoveResource(ctx)

				return
			}
		}

		resp.Diagnostics.AddError("Failed to read personal access token", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *personalAccessTokenResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Personal access tokens cannot be updated")
}

func (r *personalAccessTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PersonalAccessTokenModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.RevokePersonalAccessTokenParams{
		AccessTokenID: data.Id.ValueString(),
	}

	response, err := r.providerData.client.RevokePersonalAccessToken(ctx, params)

	tflog.Info(ctx, "personal_access_token.delete", map[string]interface{}{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.PersonalAccessToken:
		if !response.RevokedAt.IsSet() || response.RevokedAt.IsNull() {
			resp.Diagnostics.AddError("Failed to revoke personal access token", "Personal access token was not revoked")
		}

	default:
		handled := false

		if response, ok := response.(*cm.ErrorStatusCode); ok {
			switch response.StatusCode {
			case http.StatusNotFound:
				resp.Diagnostics.AddWarning("personal access token not found", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true

			case http.StatusConflict:
				resp.Diagnostics.AddWarning("personal access token already revoked", util.ErrorDetailFromContentfulManagementResponse(response, err))

				handled = true
			}
		}

		if !handled {
			resp.Diagnostics.AddError("Failed to revoke personal access token", util.ErrorDetailFromContentfulManagementResponse(response, err))
		}
	}
}
