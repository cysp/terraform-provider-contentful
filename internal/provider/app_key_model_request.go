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

	if diags.HasError() {
		return cm.AppKeyRequestData{}, diags
	}

	return cm.NewAppKeyRequestData(jwk), diags
}

func (m AppKeyJWKModel) ToAppKeyJWK(_ context.Context, attrPath path.Path) (cm.AppKeyJWK, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	alg, algDiags := requestRequiredString(m.Alg, attrPath.AtName("alg"))
	diags.Append(algDiags...)

	kty, ktyDiags := requestRequiredString(m.Kty, attrPath.AtName("kty"))
	diags.Append(ktyDiags...)

	use, useDiags := requestRequiredString(m.Use, attrPath.AtName("use"))
	diags.Append(useDiags...)

	kid, kidDiags := requestRequiredString(m.Kid, attrPath.AtName("kid"))
	diags.Append(kidDiags...)

	x5t, x5tDiags := requestRequiredString(m.X5t, attrPath.AtName("x5t"))
	diags.Append(x5tDiags...)

	var x5c []string

	if m.X5c.IsNull() || m.X5c.IsUnknown() {
		diags.AddAttributeError(attrPath.AtName("x5c"), "Missing app key JWK x5c", "Set x5c to the public key material.")
	} else {
		var x5cDiags diag.Diagnostics

		x5c, x5cDiags = knownStringListElements(attrPath.AtName("x5c"), m.X5c.Elements())
		diags.Append(x5cDiags...)
	}

	if diags.HasError() {
		return cm.AppKeyJWK{}, diags
	}

	jwk := cm.AppKeyJWK{
		Alg: cm.AppKeyJWKAlg(alg),
		Kty: cm.AppKeyJWKKty(kty),
		Use: cm.AppKeyJWKUse(use),
		X5c: x5c,
		Kid: kid,
		X5t: x5t,
	}

	diags.Append(validateAppKeyJWKMaterial(jwk, attrPath)...)

	if diags.HasError() {
		return cm.AppKeyJWK{}, diags
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

	fingerprint, err := cm.AppKeyJWKFingerprintFromX5C(x5c)
	if err != nil {
		diags.AddAttributeError(attrPath.AtName("x5c").AtListIndex(0), "Invalid app key JWK x5c", "The first x5c value must use valid standard base64 encoding without CR or LF.")

		return diags
	}

	if x5t != nil && *x5t != fingerprint {
		diags.AddAttributeError(attrPath.AtName("x5t"), "Invalid app key JWK x5t", fmt.Sprintf("The x5t value must match the base64url-encoded SHA-256 digest of x5c[0]. Expected %q.", fingerprint))
	}

	if kid != nil && *kid != fingerprint {
		diags.AddAttributeError(attrPath.AtName("kid"), "Invalid app key JWK kid", fmt.Sprintf("The kid value must match the base64url-encoded SHA-256 digest of x5c[0]. Expected %q.", fingerprint))
	}

	return diags
}
