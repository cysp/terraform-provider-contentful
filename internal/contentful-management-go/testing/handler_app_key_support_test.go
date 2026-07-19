package cmtesting_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAppKeyRSABits = 4096

func requireAppKeyErrorDetailsContain(t *testing.T, response any, fragment string) {
	t.Helper()

	errorStatus, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	errorResponse, ok := errorStatus.Response.GetError()
	require.True(t, ok)
	require.Contains(t, string(errorResponse.Details), fragment)
}

func assertAppKeyHasPrivateMaterial(t *testing.T, appKey cm.AppKey) {
	t.Helper()

	generated, ok := appKey.Generated.Get()
	require.True(t, ok)

	privateKeyBlock, remaining := pem.Decode([]byte(generated.PrivateKey))
	require.NotNil(t, privateKeyBlock)
	assert.Empty(t, remaining)
	assert.Equal(t, "PRIVATE KEY", privateKeyBlock.Type)

	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
	require.NoError(t, err)

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	require.True(t, ok)

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&rsaPrivateKey.PublicKey)
	require.NoError(t, err)
	require.Len(t, appKey.Jwk.X5c, 1)
	assert.Equal(t, base64.StdEncoding.EncodeToString(publicKeyDER), appKey.Jwk.X5c[0])

	expectedFingerprint := cm.AppKeyJWKFingerprint(publicKeyDER)
	assert.Equal(t, expectedFingerprint, appKey.Jwk.Kid)
	assert.Equal(t, expectedFingerprint, appKey.Jwk.X5t)
}

func assertAppKeyHasNoPrivateMaterial(t *testing.T, appKey cm.AppKey) {
	t.Helper()

	assert.False(t, appKey.Generated.IsSet())
}

func newAppKeyTestHandler(t *testing.T) *cmt.Handler {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app-definition", cm.AppDefinitionData{Name: "App"})

	return server.Handler()
}

func appKeyCreateParams() cm.CreateAppKeyParams {
	return cm.CreateAppKeyParams{
		OrganizationID:  "organization",
		AppDefinitionID: "app-definition",
	}
}

func appKeyRequest(t *testing.T) *cm.AppKeyRequestData {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, testAppKeyRSABits)
	require.NoError(t, err)

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	return appKeyRequestFromDER(publicKeyDER)
}

func appKeyRequestFromDER(publicKeyDER []byte) *cm.AppKeyRequestData {
	keyID := cm.AppKeyJWKFingerprint(publicKeyDER)

	request := cm.NewAppKeyRequestData(cm.AppKeyJWK{
		Alg: cm.AppKeyJWKAlgRS256,
		Kty: cm.AppKeyJWKKtyRSA,
		Use: cm.AppKeyJWKUseSig,
		Kid: keyID,
		X5c: []string{base64.StdEncoding.EncodeToString(publicKeyDER)},
		X5t: keyID,
	})

	return &request
}

func createAppKey(t *testing.T, handler *cmt.Handler, request *cm.AppKeyRequestData) cm.AppKey {
	t.Helper()

	response, err := handler.CreateAppKey(context.Background(), request, appKeyCreateParams())
	require.NoError(t, err)

	appKey, ok := response.(*cm.AppKey)
	require.True(t, ok)

	return *appKey
}
