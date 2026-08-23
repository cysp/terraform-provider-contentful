package integration_tests_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAppSigningSecretTreatsRedactedValueAsOpaqueMetadata(t *testing.T) {
	t.Parallel()

	secret := cm.AppSigningSecret{
		Sys:           cm.NewAppSigningSecretSys("organization-id", "app-definition-id"),
		RedactedValue: "changed-redaction-format",
	}

	testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/organizations/organization-id/app_definitions/app-definition-id/signing_secret", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(secret))
	}))
	defer testserver.Close()

	client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

	response, err := client.GetAppSigningSecret(t.Context(), cm.GetAppSigningSecretParams{
		OrganizationID:  "organization-id",
		AppDefinitionID: "app-definition-id",
	})
	require.NoError(t, err)

	secretResponse, ok := response.(*cm.AppSigningSecret)
	require.True(t, ok)
	assert.Equal(t, secret.RedactedValue, secretResponse.RedactedValue)
}

func TestPutAppSigningSecretTreatsRedactedValueAsOpaqueMetadata(t *testing.T) {
	t.Parallel()

	secret := cm.AppSigningSecret{
		Sys:           cm.NewAppSigningSecretSys("organization-id", "app-definition-id"),
		RedactedValue: "changed-redaction-format",
	}

	testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/organizations/organization-id/app_definitions/app-definition-id/signing_secret", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		assert.NoError(t, json.NewEncoder(w).Encode(secret))
	}))
	defer testserver.Close()

	client := testContentfulManagementClient(t, testserver.URL, cmt.ValidAccessToken)

	response, err := client.PutAppSigningSecret(t.Context(), &cm.AppSigningSecretRequestData{
		Value: strings.Repeat("s", 64),
	}, cm.PutAppSigningSecretParams{
		OrganizationID:  "organization-id",
		AppDefinitionID: "app-definition-id",
	})
	require.NoError(t, err)

	secretResponse, ok := response.(*cm.AppSigningSecretStatusCode)
	require.True(t, ok)
	assert.Equal(t, secret.RedactedValue, secretResponse.Response.RedactedValue)
}
