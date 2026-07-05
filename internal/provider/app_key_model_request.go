package provider

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *AppKeyModel) ToAppKeyRequestData(ctx context.Context) (cm.AppKeyRequestData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := cm.AppKeyRequestData{}

	jwkModel, ok := model.JWK.GetValue()
	if !ok {
		req.Generate = cm.NewOptBool(true)
	} else {
		jwk, jwkDiags := jwkModel.ToAppKeyJWK(ctx, path.Root("jwk"))
		diags.Append(jwkDiags...)

		req.Jwk = cm.NewOptAppKeyJWK(jwk)
	}

	return req, diags
}

func (model AppKeyJWKModel) ToAppKeyJWK(_ context.Context, attrPath path.Path) (cm.AppKeyJWK, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	jwk := cm.AppKeyJWK{}

	if model.Alg.IsNull() || model.Alg.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("alg"), "Missing app key JWK alg", "The alg value must be known.")
	} else {
		jwk.Alg = cm.AppKeyJWKAlg(model.Alg.ValueString())
	}

	if model.Kty.IsNull() || model.Kty.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("kty"), "Missing app key JWK kty", "The kty value must be known.")
	} else {
		jwk.Kty = cm.AppKeyJWKKty(model.Kty.ValueString())
	}

	if model.Use.IsNull() || model.Use.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("use"), "Missing app key JWK use", "The use value must be known.")
	} else {
		jwk.Use = cm.AppKeyJWKUse(model.Use.ValueString())
	}

	if model.Kid.IsNull() || model.Kid.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("kid"), "Missing app key JWK kid", "Set kid to the key identifier for the public key.")
	} else {
		jwk.Kid = model.Kid.ValueString()
	}

	if model.X5t.IsNull() || model.X5t.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("x5t"), "Missing app key JWK x5t", "Set x5t to the certificate thumbprint for the public key.")
	} else {
		jwk.X5t = model.X5t.ValueString()
	}

	if model.X5c.IsNull() || model.X5c.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("x5c"), "Missing app key JWK x5c", "Set x5c to at least one certificate for the public key.")
	} else {
		x5c := make([]string, 0, len(model.X5c.Elements()))
		for idx, value := range model.X5c.Elements() {
			if value.IsNull() || value.IsUnknown() {
				diags.AddAttributeError(
					attrPath.AtName("x5c").AtListIndex(idx),
					"Invalid app key JWK x5c certificate",
					fmt.Sprintf("The x5c certificate at index %d must be known and non-null.", idx),
				)

				continue
			}

			x5c = append(x5c, value.ValueString())
		}

		jwk.X5c = x5c
	}

	return jwk, diags
}
