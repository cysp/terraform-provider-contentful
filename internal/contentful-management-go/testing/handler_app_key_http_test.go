package cmtesting_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentfulManagementServerAppKeyValidationParity(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		payload         string
		expectedDetails string
	}{
		"missing mode": {
			payload: `{}`,
			expectedDetails: `{"errors":[
				{"name":"required","details":"The property \"jwk\" is required here","path":["jwk"]},
				{"name":"required","details":"The property \"generate\" is required here","path":["generate"]},
				{"name":"unless","details":"The property \"jwk\" or \"generate\" are required here","path":[]}
			]}`,
		},
		"generate false": {
			payload: `{"generate":false}`,
			expectedDetails: `{"errors":[
				{"name":"in","details":"Value must be one of expected values","path":["generate"],"value":false,"expected":[true]}
			]}`,
		},
		"both modes": {
			payload: `{"generate":true,"jwk":{}}`,
			expectedDetails: `{"errors":[
				{"name":"unless","details":"\"jwk\" can't be set when \"generate\" is also set","path":[]}
			]}`,
		},
		"both modes invalid": {
			payload: `{"jwk":null,"generate":null}`,
			expectedDetails: `{"errors":[
				{"name":"type","details":"The type of \"jwk\" is incorrect, expected type: Object","path":["jwk"],"type":"Object","value":null},
				{"name":"type","details":"The type of \"generate\" is incorrect, expected type: Boolean","path":["generate"],"type":"Boolean","value":null},
				{"name":"in","details":"Value must be one of expected values","path":["generate"],"value":null,"expected":[true]},
				{"name":"unless","details":"The property \"jwk\" or \"generate\" are required here","path":[]}
			]}`,
		},
		"valid jwk type with invalid generate": {
			payload: `{"jwk":{},"generate":null}`,
			expectedDetails: `{"errors":[
				{"name":"required","details":"The property \"alg\" is required here","path":["jwk","alg"]},
				{"name":"required","details":"The property \"kty\" is required here","path":["jwk","kty"]},
				{"name":"required","details":"The property \"use\" is required here","path":["jwk","use"]},
				{"name":"required","details":"The property \"x5c\" is required here","path":["jwk","x5c"]},
				{"name":"required","details":"The property \"kid\" is required here","path":["jwk","kid"]},
				{"name":"required","details":"The property \"x5t\" is required here","path":["jwk","x5t"]},
				{"name":"type","details":"The type of \"generate\" is incorrect, expected type: Boolean","path":["generate"],"type":"Boolean","value":null},
				{"name":"in","details":"Value must be one of expected values","path":["generate"],"value":null,"expected":[true]},
				{"name":"unless","details":"The property \"jwk\" or \"generate\" are required here","path":[]}
			]}`,
		},
		"invalid jwk with valid generate": {
			payload: `{"jwk":null,"generate":true}`,
			expectedDetails: `{"errors":[
				{"name":"type","details":"The type of \"jwk\" is incorrect, expected type: Object","path":["jwk"],"type":"Object","value":null},
				{"name":"unless","details":"The property \"jwk\" or \"generate\" are required here","path":[]}
			]}`,
		},
		"null jwk": {
			payload: `{"jwk":null}`,
			expectedDetails: `{"errors":[
				{"name":"type","details":"The type of \"jwk\" is incorrect, expected type: Object","path":["jwk"],"type":"Object","value":null}
			]}`,
		},
		"scalar jwk": {
			payload: `{"jwk":"abc"}`,
			expectedDetails: `{"errors":[
				{"name":"type","details":"The type of \"jwk\" is incorrect, expected type: Object","path":["jwk"],"type":"Object","value":"abc"}
			]}`,
		},
		"null generate": {
			payload: `{"generate":null}`,
			expectedDetails: `{"errors":[
				{"name":"type","details":"The type of \"generate\" is incorrect, expected type: Boolean","path":["generate"],"type":"Boolean","value":null},
				{"name":"in","details":"Value must be one of expected values","path":["generate"],"value":null,"expected":[true]}
			]}`,
		},
		"string generate": {
			payload: `{"generate":"true"}`,
			expectedDetails: `{"errors":[
				{"name":"type","details":"The type of \"generate\" is incorrect, expected type: Boolean","path":["generate"],"type":"Boolean","value":"true"},
				{"name":"in","details":"Value must be one of expected values","path":["generate"],"value":"true","expected":[true]}
			]}`,
		},
		"empty jwk": {
			payload: `{"jwk":{}}`,
			expectedDetails: `{"errors":[
				{"name":"required","details":"The property \"alg\" is required here","path":["jwk","alg"]},
				{"name":"required","details":"The property \"kty\" is required here","path":["jwk","kty"]},
				{"name":"required","details":"The property \"use\" is required here","path":["jwk","use"]},
				{"name":"required","details":"The property \"x5c\" is required here","path":["jwk","x5c"]},
				{"name":"required","details":"The property \"kid\" is required here","path":["jwk","kid"]},
				{"name":"required","details":"The property \"x5t\" is required here","path":["jwk","x5t"]}
			]}`,
		},
		"invalid fields accumulate": {
			payload: `{"jwk":{"alg":"RS512","kty":"EC","use":"enc","x5c":[],"kid":"x","x5t":"x"}}`,
			expectedDetails: `{"errors":[
				{"name":"in","details":"Value must be one of expected values","path":["jwk","alg"],"value":"RS512","expected":["RS256"]},
				{"name":"in","details":"Value must be one of expected values","path":["jwk","kty"],"value":"EC","expected":["RSA"]},
				{"name":"in","details":"Value must be one of expected values","path":["jwk","use"],"value":"enc","expected":["sig"]},
				{"name":"size","details":"Size must be at least 1 and at most 1","path":["jwk","x5c"],"value":[],"min":1,"max":1},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","kid"],"value":"x","min":42,"max":45},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","x5t"],"value":"x","min":42,"max":45}
			]}`,
		},
		"null x5c": {
			payload: `{"jwk":{"alg":"RS256","kty":"RSA","use":"sig","x5c":null,"kid":"x","x5t":"x"}}`,
			expectedDetails: `{"errors":[
				{"name":"type","details":"The type of \"x5c\" is incorrect, expected type: Array","path":["jwk","x5c"],"type":"Array","value":null},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","kid"],"value":"x","min":42,"max":45},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","x5t"],"value":"x","min":42,"max":45}
			]}`,
		},
		"null enum": {
			payload: `{"jwk":{"alg":null,"kty":"RSA","use":"sig","x5c":[],"kid":"x","x5t":"x"}}`,
			expectedDetails: `{"errors":[
				{"name":"type","details":"The type of \"alg\" is incorrect, expected type: String","path":["jwk","alg"],"type":"String","value":null},
				{"name":"in","details":"Value must be one of expected values","path":["jwk","alg"],"value":null,"expected":["RS256"]},
				{"name":"size","details":"Size must be at least 1 and at most 1","path":["jwk","x5c"],"value":[],"min":1,"max":1},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","kid"],"value":"x","min":42,"max":45},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","x5t"],"value":"x","min":42,"max":45}
			]}`,
		},
		"numeric enum": {
			payload: `{"jwk":{"alg":1,"kty":"RSA","use":"sig","x5c":[],"kid":"x","x5t":"x"}}`,
			expectedDetails: `{"errors":[
				{"name":"type","details":"The type of \"alg\" is incorrect, expected type: String","path":["jwk","alg"],"type":"String","value":1},
				{"name":"in","details":"Value must be one of expected values","path":["jwk","alg"],"value":1,"expected":["RS256"]},
				{"name":"size","details":"Size must be at least 1 and at most 1","path":["jwk","x5c"],"value":[],"min":1,"max":1},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","kid"],"value":"x","min":42,"max":45},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","x5t"],"value":"x","min":42,"max":45}
			]}`,
		},
		"scalar x5c": {
			payload: `{"jwk":{"alg":"RS256","kty":"RSA","use":"sig","x5c":"abc","kid":"x","x5t":"x"}}`,
			expectedDetails: `{"errors":[
				{"name":"type","details":"The type of \"x5c\" is incorrect, expected type: Array","path":["jwk","x5c"],"type":"Array","value":"abc"},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","kid"],"value":"x","min":42,"max":45},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","x5t"],"value":"x","min":42,"max":45}
			]}`,
		},
		"invalid x5c element type": {
			payload: `{"jwk":{"alg":"RS256","kty":"RSA","use":"sig","x5c":[null],"kid":"x","x5t":"x"}}`,
			expectedDetails: `{"errors":[
				{"name":"type","details":"The type of \"0\" is incorrect, expected type: String","path":["jwk","x5c",0],"type":"String","value":null},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","kid"],"value":"x","min":42,"max":45},
				{"name":"size","details":"Size must be at least 42 and at most 45","path":["jwk","x5t"],"value":"x","min":42,"max":45}
			]}`,
		},
		"invalid fingerprint types": {
			payload: `{"jwk":{"alg":"RS256","kty":"RSA","use":"sig","x5c":[],"kid":1,"x5t":null}}`,
			expectedDetails: `{"errors":[
				{"name":"size","details":"Size must be at least 1 and at most 1","path":["jwk","x5c"],"value":[],"min":1,"max":1},
				{"name":"type","details":"The type of \"kid\" is incorrect, expected type: String","path":["jwk","kid"],"type":"String","value":1},
				{"name":"type","details":"The type of \"x5t\" is incorrect, expected type: String","path":["jwk","x5t"],"type":"String","value":null}
			]}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			status, responseBody := postAppKeyRequest(t, test.payload)
			assert.Equal(t, http.StatusUnprocessableEntity, status)

			var response cm.Error
			require.NoError(t, json.Unmarshal(responseBody, &response))
			assert.Equal(t, "ValidationFailed", response.Sys.ID)
			assert.Equal(t, "Validation error", response.Message.Or(""))
			assert.JSONEq(t, test.expectedDetails, string(response.Details))
		})
	}
}

func TestContentfulManagementServerAppKeyFingerprintValidationParity(t *testing.T) {
	t.Parallel()

	request := appKeyRequest(t)

	var jwk cm.AppKeyJWK

	require.NoError(t, json.Unmarshal(request.Jwk, &jwk))
	jwk.X5t = strings.Repeat("x", 43)

	encodedJWK, err := json.Marshal(jwk)
	require.NoError(t, err)

	request.Jwk = encodedJWK

	payload, err := json.Marshal(request)
	require.NoError(t, err)

	status, responseBody := postAppKeyRequest(t, string(payload))
	assert.Equal(t, http.StatusUnprocessableEntity, status)

	var response cm.Error
	require.NoError(t, json.Unmarshal(responseBody, &response))
	assert.Equal(t, "ValidationFailed", response.Sys.ID)
	assert.Equal(t, "Validation error", response.Message.Or(""))
	assert.JSONEq(t, `{"errors":[{
		"name":"invalid",
		"details":"jwk.x5t must be the base64url sha256 fingerprint of the base64 encoded DER public key.",
		"path":["jwk","x5t"]
	}]}`, string(response.Details))
}

func TestContentfulManagementServerAppKeyEncodingValidationParity(t *testing.T) {
	t.Parallel()

	fingerprint := strings.Repeat("x", 43)
	payload := `{"jwk":{
		"alg":"RS256",
		"kty":"RSA",
		"use":"sig",
		"x5c":["!"],
		"kid":"` + fingerprint + `",
		"x5t":"` + fingerprint + `"
	}}`

	status, responseBody := postAppKeyRequest(t, payload)
	assert.Equal(t, http.StatusUnprocessableEntity, status)

	var response cm.Error
	require.NoError(t, json.Unmarshal(responseBody, &response))
	assert.Equal(t, "ValidationFailed", response.Sys.ID)
	assert.Equal(t, "Validation error", response.Message.Or(""))
	assert.JSONEq(t, `{"errors":[
		{
			"name":"size",
			"details":"Size must be at least 736 and at most 1416",
			"path":["jwk","x5c",0],
			"value":"!",
			"min":736,
			"max":1416
		},
		{
			"name":"regexp",
			"details":"Does not match /^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$/",
			"path":["jwk","x5c",0],
			"value":"!"
		}
	]}`, string(response.Details))
}

func postAppKeyRequest(t *testing.T, payload string) (int, []byte) {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app-definition", cm.AppDefinitionData{Name: "App"})

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		testServer.URL+"/organizations/organization/app_definitions/app-definition/keys",
		bytes.NewBufferString(payload),
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

	response, err := testServer.Client().Do(request)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, response.Body.Close())
	})

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return response.StatusCode, body
}
