package provider

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//nolint:ireturn
func NewAppJWTEphemeralResource() ephemeral.EphemeralResource {
	return &appJWTEphemeralResource{}
}

type appJWTEphemeralResource struct{}

type appJWTModel struct {
	AppDefinitionID types.String `tfsdk:"app_definition_id"`
	PrivateKey      types.String `tfsdk:"private_key"`
	JWT             types.String `tfsdk:"jwt"`
}

func (*appJWTEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_jwt"
}

func (*appJWTEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Signs an RS256 JWT locally with the app's private key. The JWT expires after ten minutes.",
		Attributes: map[string]schema.Attribute{
			"app_definition_id": schema.StringAttribute{
				Description: "App definition ID used as the JWT issuer.",
				Required:    true,
				Validators:  []validator.String{discoveryIDValidator{}},
			},
			"private_key": schema.StringAttribute{
				Description: "Unencrypted RSA private key of at least 2048 bits in PKCS#1 or PKCS#8 PEM format.",
				Required:    true,
				Sensitive:   true,
			},
			"jwt": schema.StringAttribute{
				Description: "RS256-signed JWT for the app definition.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (*appJWTEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	const (
		jwtLifetime    = 10 * time.Minute
		minimumRSABits = 2048
	)

	var data appJWTModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(requireDiscoveryID(path.Root("app_definition_id"), data.AppDefinitionID)...)

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(data.PrivateKey.ValueString()))
	if err != nil || key.N.BitLen() < minimumRSABits {
		resp.Diagnostics.AddAttributeError(path.Root("private_key"), "Invalid private key", "Supply an unencrypted RSA private key of at least 2048 bits in PKCS#1 or PKCS#8 PEM format.")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		Issuer:    data.AppDefinitionID.ValueString(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(jwtLifetime)),
	})

	signed, err := token.SignedString(key)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("private_key"), "Unable to sign JWT", "The RSA private key could not sign an RS256 JWT.")

		return
	}

	data.JWT = types.StringValue(signed)
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
