package contentfulmanagement_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppKeyJWKFingerprintFromX5C(t *testing.T) {
	t.Parallel()

	publicKeyDER := bytes.Repeat([]byte{0}, 550)
	canonicalX5C := base64.StdEncoding.EncodeToString(publicKeyDER)
	require.True(t, strings.HasSuffix(canonicalX5C, "AA=="))

	noncanonicalX5C := canonicalX5C[:len(canonicalX5C)-3] + "P=="
	require.NotEqual(t, canonicalX5C, noncanonicalX5C)

	fingerprint := sha256.Sum256(publicKeyDER)

	canonical, err := cm.AppKeyJWKFingerprintFromX5C(canonicalX5C)
	require.NoError(t, err)
	noncanonical, err := cm.AppKeyJWKFingerprintFromX5C(noncanonicalX5C)
	require.NoError(t, err)

	assert.Equal(t, base64.RawURLEncoding.EncodeToString(fingerprint[:]), canonical)
	assert.Equal(t, canonical, noncanonical)
}

func TestAppKeyJWKFingerprintFromX5CRejectsInvalidStandardBase64(t *testing.T) {
	t.Parallel()

	valid := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 600))

	tests := map[string]string{
		"line feed":        valid[:100] + "\n" + valid[100:],
		"carriage return":  valid[:100] + "\r" + valid[100:],
		"space":            valid[:100] + " " + valid[100:],
		"tab":              valid[:100] + "\t" + valid[100:],
		"invalid alphabet": "!" + valid[1:],
		"invalid padding":  valid + "=",
	}

	for name, x5c := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := cm.AppKeyJWKFingerprintFromX5C(x5c)

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

	assert.Equal(t, cm.AppKeyJWK{
		Alg: cm.AppKeyJWKAlgRS256,
		Kty: cm.AppKeyJWKKtyRSA,
		Use: cm.AppKeyJWKUseSig,
		X5c: []string{"cHVibGljLWtleQ=="},
		Kid: "kid",
		X5t: "x5t",
	}, request.Jwk)
}
