package contentfulmanagement_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeAppKeyJWKMaterial(t *testing.T) {
	t.Parallel()

	publicKeyDER := bytes.Repeat([]byte{0x42}, 600)
	x5c := base64.StdEncoding.EncodeToString(publicKeyDER)
	fingerprint := sha256.Sum256(publicKeyDER)

	material, err := cm.DecodeAppKeyJWKMaterial(x5c)

	require.NoError(t, err)
	assert.Equal(t, publicKeyDER, material.DER)
	assert.Equal(t, base64.RawURLEncoding.EncodeToString(fingerprint[:]), material.Fingerprint)
}

func TestDecodeAppKeyJWKMaterialRejectsInvalidStandardBase64(t *testing.T) {
	t.Parallel()

	valid := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 600))

	tests := map[string]string{
		"whitespace":       valid[:100] + "\n" + valid[100:],
		"invalid alphabet": "!" + valid[1:],
	}

	for name, x5c := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := cm.DecodeAppKeyJWKMaterial(x5c)

			require.Error(t, err)
		})
	}
}

func TestNewAppKeyRequestDataEncodesJWK(t *testing.T) {
	t.Parallel()

	request := cm.NewAppKeyRequestData(cm.AppKeyJWK{
		Alg: cm.AppKeyJWKAlgRS256,
		Kty: cm.AppKeyJWKKtyRSA,
		Use: cm.AppKeyJWKUseSig,
		X5c: []string{"cHVibGljLWtleQ=="},
		Kid: "kid",
		X5t: "x5t",
	})

	assert.JSONEq(t, `{
		"alg": "RS256",
		"kty": "RSA",
		"use": "sig",
		"x5c": ["cHVibGljLWtleQ=="],
		"kid": "kid",
		"x5t": "x5t"
	}`, string(request.Jwk))
	assert.Empty(t, request.Generate)
}
