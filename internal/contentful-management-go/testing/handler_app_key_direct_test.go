package cmtesting_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppKeyCollectionPagination(t *testing.T) {
	t.Parallel()

	handler := newAppKeyTestHandler(t)

	keyIDs := make([]string, 0, 3)
	for range 3 {
		keyIDs = append(keyIDs, createAppKey(t, handler, appKeyRequest(t)).Sys.ID)
	}

	params := cm.GetAppKeysParams{
		OrganizationID:  "organization",
		AppDefinitionID: "app-definition",
		Skip:            cm.NewOptInt64(1),
		Limit:           cm.NewOptInt64(1),
	}
	response, err := handler.GetAppKeys(context.Background(), params)
	require.NoError(t, err)

	collection, ok := response.(*cm.AppKeyCollection)
	require.True(t, ok)
	assert.Equal(t, 3, collection.Total)
	assert.Equal(t, 1, collection.Skip)
	assert.Equal(t, 1, collection.Limit)
	require.Len(t, collection.Items, 1)
	assert.Equal(t, keyIDs[1], collection.Items[0].Sys.ID)

	repeatedResponse, err := handler.GetAppKeys(context.Background(), params)
	require.NoError(t, err)

	repeatedCollection, ok := repeatedResponse.(*cm.AppKeyCollection)
	require.True(t, ok)
	require.Len(t, repeatedCollection.Items, 1)
	assert.Equal(t, keyIDs[1], repeatedCollection.Items[0].Sys.ID)
}

func TestAppKeyCollectionRejectsInvalidPagination(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		params  cm.GetAppKeysParams
		message string
	}{
		"negative skip": {
			params:  cm.GetAppKeysParams{Skip: cm.NewOptInt64(-1)},
			message: "Invalid skip parameter: should be a nonnegative integer",
		},
		"zero limit": {
			params:  cm.GetAppKeysParams{Limit: cm.NewOptInt64(0)},
			message: "Invalid limit parameter: should be a positive integer lower or equal 1000",
		},
		"limit above maximum": {
			params:  cm.GetAppKeysParams{Limit: cm.NewOptInt64(1001)},
			message: "Invalid limit parameter: should be a positive integer lower or equal 1000",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			handler := newAppKeyTestHandler(t)
			test.params.OrganizationID = "organization"
			test.params.AppDefinitionID = "app-definition"

			response, err := handler.GetAppKeys(context.Background(), test.params)
			require.NoError(t, err)
			requireContentfulError(t, response, http.StatusBadRequest, "BadRequest", test.message)
		})
	}
}

func TestAppKeyCreateAcceptsOpaqueFingerprintableMaterial(t *testing.T) {
	t.Parallel()

	handler := newAppKeyTestHandler(t)
	publicKeyDER := bytes.Repeat([]byte{0}, 600)
	request := appKeyRequestFromDER(publicKeyDER)

	created := createAppKey(t, handler, request)
	assert.Equal(t, cm.AppKeyJWKFingerprint(publicKeyDER), created.Sys.ID)
}

func TestAppKeyCreateAcceptsNonCanonicalBase64PaddingBits(t *testing.T) {
	t.Parallel()

	handler := newAppKeyTestHandler(t)
	publicKeyDER := bytes.Repeat([]byte{0}, 550)
	request := appKeyRequestFromDER(publicKeyDER)

	jwk := request.Jwk
	require.Len(t, jwk.X5c, 1)
	require.True(t, strings.HasSuffix(jwk.X5c[0], "AA=="))

	jwk.X5c[0] = jwk.X5c[0][:len(jwk.X5c[0])-3] + "P=="
	decoded, err := base64.StdEncoding.DecodeString(jwk.X5c[0])
	require.NoError(t, err)
	require.Equal(t, publicKeyDER, decoded)
	require.NotEqual(t, jwk.X5c[0], base64.StdEncoding.EncodeToString(decoded))

	request.Jwk = jwk

	created := createAppKey(t, handler, request)
	assert.Equal(t, cm.AppKeyJWKFingerprint(publicKeyDER), created.Sys.ID)
}

func TestAppKeyCreateRejectsBase64LineBreaks(t *testing.T) {
	t.Parallel()

	handler := newAppKeyTestHandler(t)
	request := appKeyRequestFromDER(bytes.Repeat([]byte{0}, 550))

	jwk := request.Jwk
	jwk.X5c[0] = jwk.X5c[0][:4] + "\n" + jwk.X5c[0][4:]
	request.Jwk = jwk

	response, err := handler.CreateAppKey(context.Background(), request, appKeyCreateParams())
	require.NoError(t, err)
	requireContentfulError(t, response, http.StatusUnprocessableEntity, "ValidationFailed", "Validation error")
}

func TestAppKeyCreateEnforcesX5CEncodedLength(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		derSize        int
		expectedStatus int
	}{
		"below minimum": {
			derSize:        549,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		"minimum": {
			derSize:        550,
			expectedStatus: http.StatusCreated,
		},
		"maximum": {
			derSize:        1062,
			expectedStatus: http.StatusCreated,
		},
		"above maximum": {
			derSize:        1063,
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			handler := newAppKeyTestHandler(t)
			request := appKeyRequestFromDER(bytes.Repeat([]byte{1}, test.derSize))
			response, err := handler.CreateAppKey(context.Background(), request, appKeyCreateParams())
			require.NoError(t, err)

			if test.expectedStatus == http.StatusCreated {
				_, ok := response.(*cm.AppKey)
				require.True(t, ok)

				return
			}

			requireContentfulError(t, response, test.expectedStatus, "ValidationFailed", "Validation error")
		})
	}
}

func TestAppKeyDuplicateAndLimitSemantics(t *testing.T) {
	t.Parallel()

	handler := newAppKeyTestHandler(t)
	duplicateRequest := appKeyRequest(t)
	createAppKey(t, handler, duplicateRequest)

	duplicateResponse, err := handler.CreateAppKey(context.Background(), duplicateRequest, appKeyCreateParams())
	require.NoError(t, err)
	requireContentfulError(t, duplicateResponse, http.StatusBadRequest, "BadRequest", "The key is already in use")

	createAppKey(t, handler, appKeyRequest(t))
	createAppKey(t, handler, appKeyRequest(t))

	limitResponse, err := handler.CreateAppKey(context.Background(), appKeyRequest(t), appKeyCreateParams())
	require.NoError(t, err)
	requireContentfulError(t, limitResponse, http.StatusForbidden, "AccessDenied", "Forbidden")
}

func TestAppKeyFingerprintIsGloballyUniqueAndReusableAfterDeletion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.SetAppDefinition("organization", "first-app", cm.AppDefinitionData{Name: "First App"})
	server.SetAppDefinition("organization", "second-app", cm.AppDefinitionData{Name: "Second App"})
	server.SetAppDefinition("other-organization", "third-app", cm.AppDefinitionData{Name: "Third App"})

	handler := server.Handler()
	request := appKeyRequest(t)
	firstParams := cm.CreateAppKeyParams{
		OrganizationID:  "organization",
		AppDefinitionID: "first-app",
	}

	firstResponse, err := handler.CreateAppKey(context.Background(), request, firstParams)
	require.NoError(t, err)

	firstKey, ok := firstResponse.(*cm.AppKey)
	require.True(t, ok)

	for name, params := range map[string]cm.CreateAppKeyParams{
		"another app": {
			OrganizationID:  "organization",
			AppDefinitionID: "second-app",
		},
		"another organization": {
			OrganizationID:  "other-organization",
			AppDefinitionID: "third-app",
		},
	} {
		response, err := handler.CreateAppKey(context.Background(), request, params)
		require.NoError(t, err, name)
		requireContentfulError(t, response, http.StatusBadRequest, "BadRequest", "The key is already in use")
	}

	deleteResponse, err := handler.DeleteAppKey(context.Background(), cm.DeleteAppKeyParams{
		OrganizationID:  firstParams.OrganizationID,
		AppDefinitionID: firstParams.AppDefinitionID,
		KeyKid:          firstKey.Sys.ID,
	})
	require.NoError(t, err)

	_, ok = deleteResponse.(*cm.NoContent)
	require.True(t, ok)

	reusedResponse, err := handler.CreateAppKey(context.Background(), request, cm.CreateAppKeyParams{
		OrganizationID:  "organization",
		AppDefinitionID: "second-app",
	})
	require.NoError(t, err)

	_, ok = reusedResponse.(*cm.AppKey)
	require.True(t, ok)
}

func TestAppKeyDeleteIsNotIdempotentAtTheAPI(t *testing.T) {
	t.Parallel()

	handler := newAppKeyTestHandler(t)
	created := createAppKey(t, handler, appKeyRequest(t))

	deleteResponse, err := handler.DeleteAppKey(context.Background(), cm.DeleteAppKeyParams{
		OrganizationID:  "organization",
		AppDefinitionID: "app-definition",
		KeyKid:          created.Sys.ID,
	})
	require.NoError(t, err)

	_, ok := deleteResponse.(*cm.NoContent)
	require.True(t, ok)

	getResponse, err := handler.GetAppKey(context.Background(), cm.GetAppKeyParams{
		OrganizationID:  "organization",
		AppDefinitionID: "app-definition",
		KeyKid:          created.Sys.ID,
	})
	require.NoError(t, err)
	requireContentfulError(t, getResponse, http.StatusNotFound, cm.ErrorSysIDNotFound, "The resource could not be found.")

	listResponse, err := handler.GetAppKeys(context.Background(), cm.GetAppKeysParams{
		OrganizationID:  "organization",
		AppDefinitionID: "app-definition",
	})
	require.NoError(t, err)

	collection, ok := listResponse.(*cm.AppKeyCollection)
	require.True(t, ok)
	assert.Empty(t, collection.Items)

	repeatedDeleteResponse, err := handler.DeleteAppKey(context.Background(), cm.DeleteAppKeyParams{
		OrganizationID:  "organization",
		AppDefinitionID: "app-definition",
		KeyKid:          created.Sys.ID,
	})
	require.NoError(t, err)
	requireContentfulError(t, repeatedDeleteResponse, http.StatusNotFound, cm.ErrorSysIDNotFound, "The resource could not be found.")
}

func TestAppKeyOperationsRejectCrossOrganizationParent(t *testing.T) {
	t.Parallel()

	handler := newAppKeyTestHandler(t)

	listResponse, err := handler.GetAppKeys(context.Background(), cm.GetAppKeysParams{
		OrganizationID:  "other-organization",
		AppDefinitionID: "app-definition",
	})
	require.NoError(t, err)
	requireContentfulError(t, listResponse, http.StatusNotFound, cm.ErrorSysIDNotFound, "The resource could not be found.")
	requireAppDefinitionDoesNotExistDetails(t, listResponse)

	createResponse, err := handler.CreateAppKey(context.Background(), appKeyRequest(t), cm.CreateAppKeyParams{
		OrganizationID:  "other-organization",
		AppDefinitionID: "app-definition",
	})
	require.NoError(t, err)
	requireContentfulError(t, createResponse, http.StatusNotFound, cm.ErrorSysIDNotFound, "The resource could not be found.")
	requireAppDefinitionDoesNotExistDetails(t, createResponse)

	getResponse, err := handler.GetAppKey(context.Background(), cm.GetAppKeyParams{
		OrganizationID:  "other-organization",
		AppDefinitionID: "app-definition",
		KeyKid:          "missing-key",
	})
	require.NoError(t, err)
	requireContentfulError(t, getResponse, http.StatusNotFound, cm.ErrorSysIDNotFound, "The resource could not be found.")
	requireAppDefinitionDoesNotExistDetails(t, getResponse)

	deleteResponse, err := handler.DeleteAppKey(context.Background(), cm.DeleteAppKeyParams{
		OrganizationID:  "other-organization",
		AppDefinitionID: "app-definition",
		KeyKid:          "missing-key",
	})
	require.NoError(t, err)
	requireContentfulError(t, deleteResponse, http.StatusNotFound, cm.ErrorSysIDNotFound, "The resource could not be found.")
	requireAppDefinitionDoesNotExistDetails(t, deleteResponse)
}
