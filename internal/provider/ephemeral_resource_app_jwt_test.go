package provider_test

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestAppJWTSignsWith2048BitKey(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	started := time.Now()
	_, resp := openAppJWT(t, string(privateKey))
	require.Empty(t, resp.Diagnostics)

	var token string
	require.Empty(t, resp.Result.GetAttribute(t.Context(), path.Root("jwt"), &token))
	verifyAppJWT(t, token, &key.PublicKey, started)
}

func TestAppJWTRejectsInvalidKey(t *testing.T) {
	t.Parallel()

	_, edKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(edKey)
	require.NoError(t, err)

	pkcs8 := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

	shortKey, err := rsa.GenerateKey(rand.Reader, 1024) // #nosec G403 -- Rejection fixture.
	require.NoError(t, err)

	shortPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(shortKey)})

	for _, testcase := range []struct {
		name string
		key  string
	}{
		{name: "malformed PEM", key: "invalid-private-key-canary"},
		{name: "non-RSA PKCS8", key: string(pkcs8)},
		{name: "RSA below 2048 bits", key: string(shortPEM)},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			config, resp := openAppJWT(t, testcase.key)

			require.Equal(t, []string{"private_key"}, attributeDiagnosticPaths(t, resp.Diagnostics))

			diagnostic := resp.Diagnostics.Errors()[0]
			for line := range strings.Lines(testcase.key) {
				if strings.Contains(diagnostic.Summary()+diagnostic.Detail(), strings.TrimSpace(line)) {
					t.Fatal("private key input found in diagnostic")
				}
			}

			require.True(t, resp.Result.Raw.Equal(config.Raw))
		})
	}
}

func openAppJWT(t *testing.T, privateKey string) (tfsdk.Config, ephemeral.OpenResponse) {
	t.Helper()

	ctx := t.Context()
	resource := provider.NewAppJWTEphemeralResource()

	var schema ephemeral.SchemaResponse
	resource.Schema(ctx, ephemeral.SchemaRequest{}, &schema)
	config := tfsdk.Config{
		Schema: schema.Schema,
		Raw: tftypes.NewValue(schema.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
			"app_definition_id": tftypes.NewValue(tftypes.String, "app"),
			"private_key":       tftypes.NewValue(tftypes.String, privateKey),
			"jwt":               tftypes.NewValue(tftypes.String, nil),
		}),
	}
	resp := ephemeral.OpenResponse{Result: tfsdk.EphemeralResultData(config)}
	resource.Open(ctx, ephemeral.OpenRequest{Config: config}, &resp)

	return config, resp
}

func verifyAppJWT(t *testing.T, token string, key *rsa.PublicKey, started time.Time) {
	t.Helper()

	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	require.NoError(t, err)
	require.JSONEq(t, `{"alg":"RS256","typ":"JWT"}`, string(header))

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)

	var claims struct {
		Issuer    string `json:"iss"`
		IssuedAt  int64  `json:"iat"`
		ExpiresAt int64  `json:"exp"`
	}
	require.NoError(t, json.Unmarshal(payload, &claims))
	require.Equal(t, "app", claims.Issuer)
	require.GreaterOrEqual(t, claims.IssuedAt, started.Unix())
	require.LessOrEqual(t, claims.IssuedAt, time.Now().Unix())
	require.EqualValues(t, 600, claims.ExpiresAt-claims.IssuedAt)

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	require.NoError(t, err)

	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	require.NoError(t, rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature))
}
