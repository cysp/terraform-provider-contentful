package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ ephemeral.EphemeralResourceWithConfigure = (*appAccessTokenEphemeralResource)(nil)

//nolint:ireturn
func NewAppAccessTokenEphemeralResource() ephemeral.EphemeralResource {
	return &appAccessTokenEphemeralResource{}
}

type appAccessTokenEphemeralResource struct {
	providerData ContentfulProviderData
}

type appAccessTokenModel struct {
	AppInstallationIdentityModel

	JWT       types.String `tfsdk:"jwt"`
	Token     types.String `tfsdk:"token"`
	ExpiresAt types.String `tfsdk:"expires_at"`
}

func (r *appAccessTokenEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_access_token"
}

func (r *appAccessTokenEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	idValidators := []validator.String{discoveryIDValidator{}}
	resp.Schema = schema.Schema{
		Description: "Exchanges a signed JWT for an app access token scoped to one app installation. The token is valid for ten minutes.",
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{
				Description: "ID of the space where the app is installed.",
				Required:    true,
				Validators:  idValidators,
			},
			"environment_id": schema.StringAttribute{
				Description: "ID of the concrete environment where the app is installed.",
				Required:    true,
				Validators:  idValidators,
			},
			"app_definition_id": schema.StringAttribute{
				Description: "ID of the installed app definition.",
				Required:    true,
				Validators:  idValidators,
			},
			"jwt": schema.StringAttribute{
				Description: "Signed JWT from [contentful_app_jwt](app_jwt) or another signer that follows [Contentful's JWT requirements](https://www.contentful.com/developers/docs/references/content-management-api/app-access-token/). Use the app definition ID as the issuer and supply the JWT without the Bearer prefix.",
				Required:    true,
				Sensitive:   true,
			},
			"token": schema.StringAttribute{
				Description: "App access token issued for this installation.",
				Computed:    true,
				Sensitive:   true,
			},
			"expires_at": schema.StringAttribute{
				Description: "Time the app access token expires, in RFC 3339 format.",
				Computed:    true,
			},
		},
	}
}

func (r *appAccessTokenEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	resp.Diagnostics.Append(setProviderData(req.ProviderData, &r.providerData)...)
}

func (r *appAccessTokenEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data appAccessTokenModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(requireDiscoveryID(path.Root("space_id"), data.SpaceID)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("environment_id"), data.EnvironmentID)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("app_definition_id"), data.AppDefinitionID)...)

	if data.JWT.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(path.Root("jwt"), "Missing JWT", "Supply a signed JWT before requesting the app access token.")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.CreateAppAccessTokenParams{
		SpaceID:         data.SpaceID.ValueString(),
		EnvironmentID:   data.EnvironmentID.ValueString(),
		AppDefinitionID: data.AppDefinitionID.ValueString(),
	}

	var status int

	response, err := r.providerData.client.CreateAppAccessToken(
		cm.WithAppJWT(ctx, data.JWT.ValueString()),
		params,
		cm.WithEditResponse(func(resp *http.Response) error {
			status = resp.StatusCode

			return nil
		}),
	)

	var token *cm.AppAccessToken

	switch response := response.(type) {
	case *cm.CreateAppAccessTokenOK:
		token = (*cm.AppAccessToken)(response)
	case *cm.CreateAppAccessTokenCreated:
		token = (*cm.AppAccessToken)(response)
	default:
		resp.Diagnostics.AddError("Failed to issue app access token", appAccessTokenFailureDetail(status, err))

		return
	}

	data.Token = types.StringValue(token.Token)
	data.ExpiresAt = types.StringValue(token.Sys.ExpiresAt.UTC().Format(time.RFC3339Nano))
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

func appAccessTokenFailureDetail(status int, err error) string {
	// Transport and decoding errors can include credentials.
	if errors.Is(err, context.Canceled) {
		return "The app access token request was canceled."
	}

	netErr, ok := errors.AsType[net.Error](err)
	if errors.Is(err, context.DeadlineExceeded) || (ok && netErr.Timeout()) {
		return "The app access token request timed out."
	}

	switch status {
	case 0:
		return "The app access token request failed before receiving an HTTP response."
	case http.StatusOK, http.StatusCreated:
		return fmt.Sprintf("Contentful returned HTTP %d, but the provider could not read a valid token and expiration time.", status)
	case http.StatusUnauthorized:
		return "HTTP 401. Check that the JWT is current and signed with a key registered for this app."
	default:
		return fmt.Sprintf("Contentful returned HTTP %d for the token request.", status)
	}
}
