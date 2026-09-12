package cmtesting_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

func requireAppDefinitionDoesNotExistDetails(t *testing.T, response any) {
	t.Helper()

	errorStatus, ok := response.(*cm.ErrorStatusCode)
	require.True(t, ok)
	errorResponse, ok := errorStatus.Response.GetError()
	require.True(t, ok)
	require.Contains(t, string(errorResponse.Details), "AppDefinition does not exist.")
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

func appKeyRequest(t *testing.T, index int) *cm.AppKeyRequestData {
	t.Helper()

	// These fixtures contain public 4096-bit RSA keys and fixed SHA-256
	// fingerprints. No private keys are needed to exercise JWK handling.
	data, err := os.ReadFile("testdata/app_key_public_keys.json")
	require.NoError(t, err)

	var keys []cm.AppKeyJWK
	require.NoError(t, json.Unmarshal(data, &keys))
	require.GreaterOrEqual(t, index, 0)
	require.Less(t, index, len(keys))
	request := cm.NewAppKeyRequestData(keys[index])

	return &request
}

func appKeyRequestFromDER(publicKeyDER []byte) *cm.AppKeyRequestData {
	fingerprint := sha256.Sum256(publicKeyDER)
	keyID := base64.RawURLEncoding.EncodeToString(fingerprint[:])

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

	response, err := handler.CreateAppKey(t.Context(), request, appKeyCreateParams())
	require.NoError(t, err)

	appKey, ok := response.(*cm.AppKey)
	require.True(t, ok)

	return *appKey
}
