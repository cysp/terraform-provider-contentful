package provider_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The shared server owns storage, validation, masking, and mask replay. This
// adapter records HTTP requests and injects adversarial mutation responses.
type appInstallationSecretFixture struct {
	mu                  sync.Mutex
	server              *cmt.Server
	maskOverride        string
	putResponseOverride map[string]any
	failPut             bool
	requests            []appInstallationSecretRequest
	storedAfterPut      []map[string]any
	methods             []string
	err                 error
}

type appInstallationSecretRequest struct {
	body        map[string]any
	path        string
	query       string
	marketplace string
	version     string
}

func newAppInstallationSecretFixture(t *testing.T) *appInstallationSecretFixture {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetAppDefinition("organization", "app", cm.AppDefinitionData{
		Name: "App",
		Parameters: cm.NewOptAppDefinitionParameters(cm.AppDefinitionParameters{Installation: []cm.AppDefinitionParameter{
			{ID: "secret", Type: "Secret", Name: "Secret"},
			{ID: "visible", Type: "Symbol", Name: "Visible"},
			{ID: "remove", Type: "Symbol", Name: "Remove"},
			{ID: "retained", Type: "Symbol", Name: "Retained"},
		}}),
	})

	return &appInstallationSecretFixture{server: server}
}

func (f *appInstallationSecretFixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.methods = append(f.methods, r.Method)
	if r.Method == http.MethodPut {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			f.err = err

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		var decoded map[string]any

		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.UseNumber()

		err = decoder.Decode(&decoded)
		if err != nil {
			f.err = err

			w.WriteHeader(http.StatusBadRequest)

			return
		}

		f.requests = append(f.requests, appInstallationSecretRequest{body: decoded, path: r.URL.Path, query: r.URL.RawQuery, marketplace: r.Header.Get("X-Contentful-Marketplace"), version: r.Header.Get("X-Contentful-Version")})
		r.Body = io.NopCloser(bytes.NewReader(body))
	}

	response := httptest.NewRecorder()
	f.server.ServeHTTP(response, r)

	if r.Method == http.MethodPut && response.Code == http.StatusOK {
		stored, _ := f.server.StoredAppInstallation("space", "master", "app")

		var parameters map[string]any
		if len(stored.Parameters) != 0 {
			err := json.Unmarshal(stored.Parameters, &parameters)
			if err != nil {
				f.err = err
			}
		}

		f.storedAfterPut = append(f.storedAfterPut, parameters)
	}

	maps.Copy(w.Header(), response.Header())
	w.Header().Set("Content-Type", "application/json")

	if response.Code == http.StatusOK && r.Method == http.MethodPut && f.failPut {
		// Delegate first: the write committed even though its response now fails.
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "ServerError"}, "message": "Synthetic failure after write"}))

		return
	}

	body, err := f.responseBody(response, r.Method)
	if err != nil {
		f.err = err

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(response.Code)
	// #nosec G705 -- forwards the JSON server response or JSON-encoded override with application/json set above.
	_, _ = w.Write(body)
}

func (f *appInstallationSecretFixture) responseBody(response *httptest.ResponseRecorder, method string) ([]byte, error) {
	body := response.Body.Bytes()
	if response.Code != http.StatusOK || f.maskOverride == "" && (method != http.MethodPut || f.putResponseOverride == nil) {
		return body, nil
	}

	var decoded map[string]any

	err := json.Unmarshal(body, &decoded)
	if err != nil {
		return nil, fmt.Errorf("decode app installation response: %w", err)
	}

	if parameters, ok := decoded["parameters"].(map[string]any); ok && f.maskOverride != "" {
		if _, ok := parameters["secret"].(string); ok {
			parameters["secret"] = f.maskOverride
		}
	}

	if method == http.MethodPut && f.putResponseOverride != nil {
		decoded["parameters"] = f.putResponseOverride
	}

	return []byte(testJSON(decoded)), nil
}

func (f *appInstallationSecretFixture) handlerError() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.err
}

func (f *appInstallationSecretFixture) check(t *testing.T, expected map[string]any, requests []appInstallationSecretRequest) resource.TestCheckFunc {
	t.Helper()

	return func(_ *terraform.State) error {
		t.Helper()
		f.mu.Lock()
		defer f.mu.Unlock()

		stored, ok := f.server.StoredAppInstallation("space", "master", "app")
		require.True(t, ok, "installation exists")

		var parameters map[string]any
		if len(stored.Parameters) != 0 {
			require.NoError(t, json.Unmarshal(stored.Parameters, &parameters))
		}

		assert.Equal(t, expected, parameters, "backing parameters")
		assert.Equal(t, requests, f.requests, "exact PUT requests")

		return nil
	}
}

func (f *appInstallationSecretFixture) checkDestroyed(t *testing.T) resource.TestCheckFunc {
	t.Helper()

	return func(_ *terraform.State) error {
		t.Helper()

		_, ok := f.server.StoredAppInstallation("space", "master", "app")
		assert.False(t, ok, "installation remains after destroy")

		return nil
	}
}

func (f *appInstallationSecretFixture) setStored(parameters map[string]any) {
	fields := cm.AppInstallationData{}
	if parameters != nil {
		fields.Parameters = []byte(testJSON(parameters))
	}

	f.server.SetAppInstallation("space", "master", "app", fields)
}

func appInstallationSecretConfig(parameters map[string]any, extra string) string {
	encoded := "null"
	if parameters != nil {
		encoded = fmt.Sprintf("%q", testJSON(parameters))
	}

	return fmt.Sprintf(`
resource "contentful_app_installation" "test" {
 space_id = "space"
 environment_id = "master"
 app_definition_id = "app"
 parameters = %[1]s
 %[2]s
}
`, encoded, extra)
}

func appInstallationSecretPut(parameters map[string]any, marketplace string) appInstallationSecretRequest {
	body := map[string]any{}
	if parameters != nil {
		body["parameters"] = parameters
	}

	return appInstallationSecretRequest{body: body, path: "/spaces/space/environments/master/app_installations/app", marketplace: marketplace}
}
