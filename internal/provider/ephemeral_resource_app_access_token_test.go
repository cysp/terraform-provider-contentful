package provider_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/stretchr/testify/require"
)

func TestAccAppAccessTokenEphemeralResourceSavedPlan(t *testing.T) {
	const (
		variable      = "TF_VAR_app_private_key"
		defaultLocale = "en-AU"
	)

	planKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)
	applyKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)
	applyDER, err := x509.MarshalPKCS8PrivateKey(applyKey)
	require.NoError(t, err)

	planPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(planKey)})
	applyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: applyDER})

	t.Setenv(variable, string(planPEM))
	t.Setenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN", cmt.ValidAccessToken)
	cma := appAccessTokenTestServer(t, defaultLocale, &planKey.PublicKey, &applyKey.PublicKey)

	var exchangeMu sync.Mutex

	exchanges := make(map[string]string)
	started := time.Now()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /spaces/space/environments/environment/app_installations/app/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)

		jwt, bearer := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if err != nil || len(body) != 0 || r.URL.RawQuery != "" || !bearer {
			http.Error(w, "Unexpected token exchange request", http.StatusBadRequest)

			return
		}

		record := httptest.NewRecorder()
		cma.ServeHTTP(record, r)

		if record.Code == http.StatusCreated {
			var issued struct {
				Token string `json:"token"`
			}

			err := json.Unmarshal(record.Body.Bytes(), &issued)
			if err != nil || issued.Token == "" {
				http.Error(w, "Expected an App Access Token", http.StatusInternalServerError)

				return
			}

			exchangeMu.Lock()
			exchanges[issued.Token] = jwt
			exchangeMu.Unlock()
		}

		maps.Copy(w.Header(), record.Header())
		w.WriteHeader(record.Code)
		_, _ = w.Write(record.Body.Bytes())
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")

		exchangeMu.Lock()
		_, issued := exchanges[token]
		exchangeMu.Unlock()

		if !issued {
			_ = cmt.WriteContentfulManagementErrorResponse(w, http.StatusUnauthorized, "AccessTokenInvalid", new("Expected an app access token issued during this phase."), nil)

			return
		}

		cma.ServeHTTP(w, r)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	secrets := []string{"-----BEGIN RSA PRIVATE KEY-----", "-----BEGIN PRIVATE KEY-----"}
	checkExchanges := func(key *rsa.PublicKey) {
		t.Helper()
		exchangeMu.Lock()
		defer exchangeMu.Unlock()

		require.NotEmpty(t, exchanges)

		for token, jwt := range exchanges {
			verifyAppJWT(t, jwt, key, started)
			secrets = append(secrets, jwt, token)
		}

		clear(exchanges)
	}
	checkSecrets := func(data string) {
		t.Helper()

		for _, secret := range secrets {
			if strings.Contains(data, secret) {
				t.Fatal("credential found in a Terraform artifact")
			}
		}
	}
	dir := t.TempDir()
	fixture := t.TempDir()
	require.NoError(t, os.CopyFS(fixture, os.DirFS("../../examples/ephemeral-resources/contentful_app_access_token")))
	require.NoError(t, os.CopyFS(fixture, os.DirFS(filepath.Join("testdata", t.Name()))))

	var stateFilename string

	resource.Test(t, validateAcceptanceTestCase(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_10_0)},
		WorkingDir:             dir,
		Steps: []resource.TestStep{{
			ProtoV6ProviderFactories: makeTestAccProtoV6ProviderFactories(provider.WithContentfulURL(server.URL)),
			ConfigDirectory:          config.StaticDirectory(fixture),
			ConfigVariables: config.Variables{
				"contentful_space_id":       config.StringVariable("space"),
				"contentful_environment_id": config.StringVariable("environment"),
				"app_definition_id":         config.StringVariable("app"),
				"default_locale_code":       config.StringVariable(defaultLocale),
			},
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectKnownOutputValue("planned_locale", knownvalue.StringExact(defaultLocale)),
				plancheck.ExpectResourceAction("contentful_entry.author", plancheck.ResourceActionCreate),
				afterPlan(func() {
					require.Empty(t, appAccessTokenTestEntries(t, cma))

					files, err := filepath.Glob(filepath.Join(dir, "*", "tfplan"))
					require.NoError(t, err)
					require.Len(t, files, 1)
					checkExchanges(&planKey.PublicKey)
					checkSavedPlanMembers(t, files[0], checkSecrets)
					stateFilename = filepath.Join(filepath.Dir(files[0]), "terraform.tfstate")

					t.Setenv(variable, string(applyPEM))
				}),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_entry.author", tfjsonpath.New("published_version"), knownvalue.NotNull()),
			},
			PostApplyFunc: func() {
				checkExchanges(&applyKey.PublicKey)

				data, err := os.ReadFile(stateFilename)
				require.NoError(t, err)
				checkSecrets(string(data))

				entries := appAccessTokenTestEntries(t, cma)
				require.Len(t, entries, 1)
				entry := entries[0]
				require.Equal(t, "author", entry.Sys.ContentType.Sys.ID)
				require.Equal(t, cm.NewOptInt(entry.Sys.Version-1), entry.Sys.PublishedVersion, "the stored Entry must have no unpublished changes")

				var names map[string]string
				require.NoError(t, json.Unmarshal(entry.Fields.Value["name"], &names))
				require.NotEmpty(t, strings.TrimSpace(names[defaultLocale]))
			},
		}},
	}))
}

func appAccessTokenTestEntries(t *testing.T, server *cmt.Server) []cm.Entry {
	t.Helper()

	result, err := server.Handler().GetEntries(t.Context(), cm.GetEntriesParams{SpaceID: "space", EnvironmentID: "environment"})
	require.NoError(t, err)

	collection, ok := result.(*cm.EntryCollection)
	require.True(t, ok)

	return collection.Items
}

func appAccessTokenTestServer(t *testing.T, defaultLocale string, keys ...*rsa.PublicKey) *cmt.Server {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "Test app"})

	for _, key := range keys {
		der, err := x509.MarshalPKIXPublicKey(key)
		require.NoError(t, err)

		jwk := testAccAppKeyJWKFromDER(der)
		body := testJSON(&cm.AppKeyRequestData{Jwk: cm.AppKeyJWK{
			Alg: "RS256", Kty: "RSA", Use: "sig", X5c: []string{jwk.x5c}, Kid: jwk.kid, X5t: jwk.x5t,
		}})
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/organizations/organization/app_definitions/app/keys", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
		req.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

		response := httptest.NewRecorder()
		server.ServeHTTP(response, req)
		require.Equal(t, http.StatusCreated, response.Code, "register app key: %s", response.Body.String())
	}

	server.RegisterSpaceEnvironment("space", "environment")
	server.SetAppInstallation("space", "environment", "app", cm.AppInstallationData{})
	server.SetLocale(cmt.NewLocaleFromData("space", "environment", "australian-english", cm.LocaleData{Name: "English (Australia)", Code: defaultLocale}, true))
	server.SetContentType("space", "environment", "author", cm.ContentTypeRequestData{
		Name:         "Author",
		DisplayField: "name",
		Fields: []cm.ContentTypeRequestDataFieldsItem{
			{ID: "name", Name: "Name", Type: "Symbol", Required: cm.NewOptBool(true)},
		},
	})
	activated, err := server.Handler().ActivateContentType(t.Context(), cm.ActivateContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "author", XContentfulVersion: 1,
	})
	require.NoError(t, err)
	require.IsType(t, &cm.ContentTypeStatusCode{}, activated)

	return server
}
