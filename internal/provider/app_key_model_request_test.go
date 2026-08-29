package provider_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	p "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppKeyRequestAggregatesRequiredStringDiagnostics(t *testing.T) {
	t.Parallel()

	model := p.AppKeyModel{
		JWK: p.NewTypedObject(p.AppKeyJWKModel{
			Alg: types.StringNull(),
			Kty: types.StringUnknown(),
			Use: types.StringNull(),
			Kid: types.StringUnknown(),
			X5t: types.StringNull(),
			X5c: p.NewTypedList([]types.String{types.StringUnknown()}),
		}),
	}

	request, diags := model.ToAppKeyRequestData(context.Background())

	assert.Equal(t, cm.AppKeyRequestData{}, request)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{
		"jwk.alg",
		"jwk.kty",
		"jwk.use",
		"jwk.kid",
		"jwk.x5t",
		"jwk.x5c[0]",
	}, attributeDiagnosticPaths(t, diags))
}

func TestValidateAppKeyJWKMaterialAcceptsFingerprintableMaterial(t *testing.T) {
	t.Parallel()

	material := make([]byte, 600)
	fingerprint := sha256.Sum256(material)
	x5t := base64.RawURLEncoding.EncodeToString(fingerprint[:])
	model := p.AppKeyModel{
		JWK: p.NewTypedObject(p.AppKeyJWKModel{
			Alg: types.StringValue("RS256"),
			Kty: types.StringValue("RSA"),
			Use: types.StringValue("sig"),
			X5c: p.NewTypedList([]types.String{types.StringValue(base64.StdEncoding.EncodeToString(material))}),
			Kid: types.StringValue(x5t),
			X5t: types.StringValue(x5t),
		}),
	}
	_, diags := model.ToAppKeyRequestData(context.Background())

	assert.False(t, diags.HasError(), diags)
}

func TestValidateAppKeyJWKMaterialDoesNotEnforceUndocumentedContentfulSizeBounds(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	x5c := base64.StdEncoding.EncodeToString(publicKeyDER)
	require.Less(t, len(x5c), 736)

	fingerprint := sha256.Sum256(publicKeyDER)
	x5t := base64.RawURLEncoding.EncodeToString(fingerprint[:])
	model := p.AppKeyModel{
		JWK: p.NewTypedObject(p.AppKeyJWKModel{
			Alg: types.StringValue("RS256"),
			Kty: types.StringValue("RSA"),
			Use: types.StringValue("sig"),
			X5c: p.NewTypedList([]types.String{types.StringValue(x5c)}),
			Kid: types.StringValue(x5t),
			X5t: types.StringValue(x5t),
		}),
	}

	_, diags := model.ToAppKeyRequestData(context.Background())
	assert.False(t, diags.HasError(), diags)
}
