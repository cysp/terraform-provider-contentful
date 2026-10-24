package cmtesting_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensionHTTPVersionPreconditions(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	const extensionPath = "/spaces/space/environments/environment/extensions/extension"

	put := func(name string, version int) *httptest.ResponseRecorder {
		body := testJSON(map[string]any{"extension": map[string]any{
			"name": name, "srcdoc": "<!doctype html>", "fieldTypes": []any{}, "sidebar": false,
		}})
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, extensionPath, strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer CFPAT-12345")
		request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
		request.Header.Set("X-Contentful-Version", strconv.Itoa(version))

		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		return response
	}
	assertExtension := func(response *httptest.ResponseRecorder, status int, name string, version int) {
		t.Helper()
		assert.Equal(t, status, response.Code)

		var decoded struct {
			Sys struct {
				ID      string `json:"id"`
				Version int    `json:"version"`
			} `json:"sys"`
			Extension struct {
				Name string `json:"name"`
			} `json:"extension"`
		}

		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &decoded))
		assert.Equal(t, "extension", decoded.Sys.ID)
		assert.Equal(t, version, decoded.Sys.Version)
		assert.Equal(t, name, decoded.Extension.Name)
	}

	assertExtension(put("Initial", 0), http.StatusCreated, "Initial", 1)
	assertExtension(put("Updated", 1), http.StatusOK, "Updated", 2)

	for _, version := range []int{0, 1, 3} {
		response := put("Rejected", version)
		assert.Equal(t, http.StatusConflict, response.Code)

		var failure struct {
			Sys struct {
				Type string `json:"type"`
				ID   string `json:"id"`
			} `json:"sys"`
			Message string `json:"message"`
		}

		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &failure))
		assert.Equal(t, "Error", failure.Sys.Type)
		assert.Equal(t, "VersionMismatch", failure.Sys.ID)
		assert.NotEmpty(t, failure.Message)

		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, extensionPath, nil)
		request.Header.Set("Authorization", "Bearer CFPAT-12345")

		stored := httptest.NewRecorder()
		server.ServeHTTP(stored, request)
		assertExtension(stored, http.StatusOK, "Updated", 2)
	}

	assertExtension(put("Latest", 2), http.StatusOK, "Latest", 3)
}
