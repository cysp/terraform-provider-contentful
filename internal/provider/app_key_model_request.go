package provider

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (m *AppKeyModel) ToAppKeyRequestData(ctx context.Context) (cm.AppKeyRequestData, diag.Diagnostics) {
	jwkModel, _ := m.JWK.GetValue()
	jwk, diags := jwkModel.ToAppKeyJWK(ctx, path.Root("jwk"))

	return cm.NewAppKeyRequestData(jwk), diags
}

func (m AppKeyJWKModel) ToAppKeyJWK(_ context.Context, attrPath path.Path) (cm.AppKeyJWK, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	jwk := cm.AppKeyJWK{}

	if m.Alg.IsNull() || m.Alg.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("alg"), "Missing app key JWK alg", "The alg value must be known.")
	} else {
		jwk.Alg = cm.AppKeyJWKAlg(m.Alg.ValueString())
	}

	if m.Kty.IsNull() || m.Kty.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("kty"), "Missing app key JWK kty", "The kty value must be known.")
	} else {
		jwk.Kty = cm.AppKeyJWKKty(m.Kty.ValueString())
	}

	if m.Use.IsNull() || m.Use.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("use"), "Missing app key JWK use", "The use value must be known.")
	} else {
		jwk.Use = cm.AppKeyJWKUse(m.Use.ValueString())
	}

	if m.Kid.IsNull() || m.Kid.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("kid"), "Missing app key JWK kid", "Set kid to the key identifier for the public key.")
	} else {
		jwk.Kid = m.Kid.ValueString()
	}

	if m.X5t.IsNull() || m.X5t.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("x5t"), "Missing app key JWK x5t", "Set x5t to the public-key fingerprint.")
	} else {
		jwk.X5t = m.X5t.ValueString()
	}

	if m.X5c.IsNull() || m.X5c.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("x5c"), "Missing app key JWK x5c", "Set x5c to the public key material.")
	} else {
		x5c := make([]string, 0, len(m.X5c.Elements()))
		for idx, value := range m.X5c.Elements() {
			if value.IsNull() || value.IsUnknown() {
				diags.AddAttributeError(
					attrPath.AtName("x5c").AtListIndex(idx),
					"Invalid app key JWK x5c public key",
					fmt.Sprintf("The x5c public key at index %d must be known and non-null.", idx),
				)

				continue
			}

			x5c = append(x5c, value.ValueString())
		}

		jwk.X5c = x5c
	}

	if !diags.HasError() {
		diags.Append(validateAppKeyJWKMaterial(jwk, attrPath)...)
	}

	return jwk, diags
}

func validateAppKeyJWKMaterial(jwk cm.AppKeyJWK, attrPath path.Path) diag.Diagnostics {
	if len(jwk.X5c) == 0 {
		return nil
	}

	return validateAppKeyJWKMaterialValues(jwk.X5c[0], &jwk.Kid, &jwk.X5t, attrPath)
}

func validateKnownAppKeyJWKMaterial(jwk AppKeyJWKModel, attrPath path.Path) diag.Diagnostics {
	if jwk.X5c.IsNull() || jwk.X5c.IsUnknown() || len(jwk.X5c.Elements()) != 1 {
		return nil
	}

	x5c := jwk.X5c.Elements()[0]
	if x5c.IsNull() || x5c.IsUnknown() {
		return nil
	}

	var kid, x5t *string

	if !jwk.Kid.IsNull() && !jwk.Kid.IsUnknown() {
		kidValue := jwk.Kid.ValueString()
		kid = &kidValue
	}

	if !jwk.X5t.IsNull() && !jwk.X5t.IsUnknown() {
		x5tValue := jwk.X5t.ValueString()
		x5t = &x5tValue
	}

	return validateAppKeyJWKMaterialValues(x5c.ValueString(), kid, x5t, attrPath)
}

func validateAppKeyJWKMaterialValues(x5c string, kid, x5t *string, attrPath path.Path) diag.Diagnostics {
	diags := diag.Diagnostics{}

	material, err := cm.DecodeAppKeyJWKMaterial(x5c)
	if err != nil {
		diags.AddAttributeError(attrPath.AtName("x5c").AtListIndex(0), "Invalid app key JWK x5c", "The first x5c value must use standard base64 encoding without whitespace.")

		return diags
	}

	if x5t != nil && *x5t != material.Fingerprint {
		diags.AddAttributeError(attrPath.AtName("x5t"), "Invalid app key JWK x5t", fmt.Sprintf("The x5t value must match the base64url-encoded SHA-256 digest of x5c[0]. Expected %q.", material.Fingerprint))
	}

	if kid != nil && *kid != material.Fingerprint {
		diags.AddAttributeError(attrPath.AtName("kid"), "Invalid app key JWK kid", fmt.Sprintf("The kid value must match the base64url-encoded SHA-256 digest of x5c[0]. Expected %q.", material.Fingerprint))
	}

	return diags
}
