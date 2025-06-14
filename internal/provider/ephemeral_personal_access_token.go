package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ ephemeral.EphemeralResource              = (*ephemeralPersonalAccessTokenResource)(nil)
	_ ephemeral.EphemeralResourceWithConfigure = (*ephemeralPersonalAccessTokenResource)(nil)
)

//nolint:ireturn
func NewEphemeralPersonalAccessTokenResource() ephemeral.EphemeralResource {
	return &ephemeralPersonalAccessTokenResource{}
}

type ephemeralPersonalAccessTokenResource struct {
	providerData ContentfulProviderData
}

func (r *ephemeralPersonalAccessTokenResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_personal_access_token"
}

func (r *ephemeralPersonalAccessTokenResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = PersonalAccessTokenEphemeralResourceSchema(ctx)
}

func (r *ephemeralPersonalAccessTokenResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromEphemeralResourceConfigureRequest(req, &r.providerData)...)
}

//nolint:dupl
func (r *ephemeralPersonalAccessTokenResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data PersonalAccessTokenModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request, requestDiags := data.ToPersonalAccessTokenRequestFields(ctx)
	resp.Diagnostics.Append(requestDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.providerData.client.CreatePersonalAccessToken(ctx, &request)

	tflog.Info(ctx, "personal_access_token.open", map[string]interface{}{
		"request":  request,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.PersonalAccessTokenStatusCode:
		responseModel, responseModelDiags := NewPersonalAccessTokenResourceModelFromResponse(ctx, response.Response, data.Token, data.ExpiresIn)
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel

	default:
		resp.Diagnostics.AddError("Failed to create personal access token", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
