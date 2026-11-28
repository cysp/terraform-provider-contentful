package cmtesting_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppAccessTokenExchange(t *testing.T) {
	t.Parallel()

	key, appKey := newAppAccessTokenSigningKey(t)

	synctest.Test(t, func(t *testing.T) {
		server := appInstallationServer(t, cm.AppDefinitionParameter{ID: "secret", Name: "Secret", Type: "Secret"})
		server.SetAppKey("org", "app", *appKeyRequest(t, 0))
		server.SetAppKey("org", "app", *appKey)
		server.RegisterSpaceEnvironment("space", "other")

		exchange := func(bearer string, status int) *httptest.ResponseRecorder {
			t.Helper()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, appInstallationPath+"/access_tokens", nil)

			return appIdentityHTTP(t, server, req, bearer, status)
		}
		signed := func(issuer, kid string) string {
			t.Helper()

			token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
				Issuer: issuer, IssuedAt: jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
			})
			if kid != "" {
				token.Header["kid"] = kid
			}

			raw, err := token.SignedString(key)
			require.NoError(t, err)

			return raw
		}

		appJWT := signed("app", "")
		exchange(appJWT, http.StatusNotFound)
		server.SetAppInstallation("space", "environment", "app", cm.AppInstallationData{Parameters: []byte(`{"secret":"alpha"}`)})
		exchange(cmt.ValidAccessToken, http.StatusUnauthorized)
		exchange(signed("other", ""), http.StatusUnauthorized)
		exchange(signed("app", appKeyRequest(t, 0).Jwk.Kid), http.StatusUnauthorized)
		exchange(signed("app", appKey.Jwk.Kid), http.StatusCreated)
		response := exchange(appJWT, http.StatusCreated)

		var issued cm.AppAccessToken
		require.NoError(t, json.NewDecoder(response.Body).Decode(&issued))
		require.NotEmpty(t, issued.Token)
		assert.Equal(t, time.Now().UTC().Add(10*time.Minute), issued.Sys.ExpiresAt)

		checkAppAccessTokenPermissions(t, server, issued.Token)

		checkLocales := func(status int) {
			t.Helper()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/spaces/space/environments/environment/locales", nil)
			appIdentityHTTP(t, server, req, issued.Token, status)
		}

		time.Sleep(time.Minute)
		exchange(appJWT, http.StatusUnauthorized)
		checkLocales(http.StatusOK)
		server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space", "environment", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("other")}))
		checkLocales(http.StatusForbidden)
		time.Sleep(time.Until(issued.Sys.ExpiresAt))
		checkLocales(http.StatusUnauthorized)
	})
}

func TestAppAccessTokenJWTRequirements(t *testing.T) {
	t.Parallel()

	key, appKey := newAppAccessTokenSigningKey(t)
	server := appInstallationServer(t)
	server.SetAppKey("org", "app", *appKey)
	server.SetAppInstallation("space", "environment", "app", cm.AppInstallationData{})

	expiresAt := jwt.NewNumericDate(time.Now().Add(5 * time.Minute))
	issuedAt := jwt.NewNumericDate(expiresAt.Add(-10 * time.Minute))

	for _, testCase := range []struct {
		name     string
		typ      string
		issuedAt *jwt.NumericDate
		status   int
	}{
		{"ten minute lifetime", "JWT", issuedAt, http.StatusCreated},
		{"missing issued at", "JWT", nil, http.StatusUnauthorized},
		{"missing type", "", issuedAt, http.StatusUnauthorized},
		{"wrong type", "other", issuedAt, http.StatusUnauthorized},
		{"over ten minute lifetime", "JWT", jwt.NewNumericDate(issuedAt.Add(-time.Second)), http.StatusUnauthorized},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
				Issuer: "app", IssuedAt: testCase.issuedAt, ExpiresAt: expiresAt,
			})
			if testCase.typ == "" {
				delete(token.Header, "typ")
			} else {
				token.Header["typ"] = testCase.typ
			}

			raw, err := token.SignedString(key)
			require.NoError(t, err)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, appInstallationPath+"/access_tokens", nil)
			appIdentityHTTP(t, server, req, raw, testCase.status)
		})
	}
}

func newAppAccessTokenSigningKey(t *testing.T) (*rsa.PrivateKey, *cm.AppKeyRequestData) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)

	return key, appKeyRequestFromDER(der)
}

func checkAppAccessTokenPermissions(t *testing.T, server *cmt.Server, token string) {
	t.Helper()

	const (
		environment = "/spaces/space/environments/environment"
		alias       = "/spaces/space/environments/editorial"
		locales     = environment + "/locales"
	)

	request := func(method, path, bearer string, status int) *httptest.ResponseRecorder {
		t.Helper()

		req := httptest.NewRequestWithContext(t.Context(), method, path, nil)

		return appIdentityHTTP(t, server, req, bearer, status)
	}

	update := httptest.NewRequestWithContext(t.Context(), http.MethodPut, appInstallationPath, strings.NewReader(`{"parameters":{"secret":"bravo"}}`))
	update.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	appIdentityHTTP(t, server, update, token, http.StatusForbidden)
	request(http.MethodDelete, appInstallationPath, token, http.StatusForbidden)

	server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space", "editorial", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("environment")}))

	for _, read := range []struct{ token, parameters string }{
		{token, `{"secret":"alpha"}`},
		{cmt.ValidAccessToken, `{"secret":"*****"}`},
	} {
		readResponse := request(http.MethodGet, alias+"/app_installations/app", read.token, http.StatusOK)

		var installation cm.AppInstallation
		require.NoError(t, json.NewDecoder(readResponse.Body).Decode(&installation))
		assert.Equal(t, "editorial", installation.Sys.Environment.Sys.ID)
		assert.JSONEq(t, read.parameters, string(installation.Parameters))
	}

	server.SetEditorInterface("space", "environment", "author", cm.EditorInterfaceData{})
	server.SetTag("space", "environment", "editorial", cm.TagRequest{Name: "Editorial"})
	server.SetLocale(cmt.NewLocaleFromData("space", "environment", "secondary", cm.LocaleData{Name: "French", Code: "fr"}, false))
	request(http.MethodGet, environment+"/content_types", token, http.StatusOK)
	request(http.MethodGet, alias+"/content_types/author/editor_interface", token, http.StatusOK)
	request(http.MethodGet, environment+"/entries", token, http.StatusOK)
	request(http.MethodDelete, environment+"/tags/editorial", token, http.StatusNoContent)

	localeRead := request(http.MethodGet, alias+"/locales/secondary", token, http.StatusOK)

	var locale cm.Locale
	require.NoError(t, json.NewDecoder(localeRead.Body).Decode(&locale))
	assert.Equal(t, "editorial", locale.Sys.Environment.Sys.ID)
	request(http.MethodDelete, alias+"/locales/secondary", token, http.StatusNoContent)
	request(http.MethodGet, locales+"/secondary", cmt.ValidAccessToken, http.StatusNotFound)

	create := httptest.NewRequestWithContext(t.Context(), http.MethodPost, environment+"/entries", strings.NewReader(`{"fields":{"name":{"en-US":"Ursula Le Guin"}}}`))
	create.Header.Set("X-Contentful-Content-Type", "author")
	create.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

	created := appIdentityHTTP(t, server, create, token, http.StatusCreated)

	var entry cm.Entry
	require.NoError(t, json.NewDecoder(created.Body).Decode(&entry))
	require.NotEmpty(t, entry.Sys.ID)

	aliasEntryPath := alias + "/entries/" + entry.Sys.ID
	readAliasEntry := func(bearer string) cm.Entry {
		t.Helper()

		response := request(http.MethodGet, aliasEntryPath, bearer, http.StatusOK)

		var result cm.Entry
		require.NoError(t, json.NewDecoder(response.Body).Decode(&result))

		return result
	}

	for _, bearer := range []string{token, cmt.ValidAccessToken} {
		result := readAliasEntry(bearer)
		assert.Equal(t, entry.Sys.ID, result.Sys.ID)
		assert.Equal(t, "editorial", result.Sys.Environment.Sys.ID)
		assert.Equal(t, entry.Fields, result.Fields)
	}

	entryPath := environment + "/entries/" + entry.Sys.ID
	publish := httptest.NewRequestWithContext(t.Context(), http.MethodPut, aliasEntryPath+"/published", nil)
	publish.Header.Set("X-Contentful-Version", "1")
	publishedResponse := appIdentityHTTP(t, server, publish, token, http.StatusOK)

	var published cm.Entry
	require.NoError(t, json.NewDecoder(publishedResponse.Body).Decode(&published))
	assert.Equal(t, "editorial", published.Sys.Environment.Sys.ID)
	assert.Equal(t, entry.Sys.ID, published.Sys.ID)
	request(http.MethodDelete, entryPath+"/published", token, http.StatusOK)
	request(http.MethodDelete, entryPath, token, http.StatusNoContent)

	server.SetEntry("space", "other", "author", "author", cm.EntryRequest{})
	server.SetEntry("other", "environment", "author", "author", cm.EntryRequest{})
	server.SetAppDefinition("org", "other", cm.AppDefinitionData{Name: "Other"})

	install := httptest.NewRequestWithContext(t.Context(), http.MethodPut, alias+"/app_installations/other", strings.NewReader(`{}`))
	install.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	appIdentityHTTP(t, server, install, cmt.ValidAccessToken, http.StatusOK)

	for _, path := range []string{
		"/spaces/space/environments/other/entries/author",
		"/spaces/other/environments/environment/entries/author",
		environment + "/app_installations/other",
	} {
		request(http.MethodDelete, path, token, http.StatusForbidden)
		request(http.MethodGet, path, cmt.ValidAccessToken, http.StatusOK)
	}

	request(http.MethodGet, locales, token, http.StatusOK)
	request(http.MethodGet, "/spaces/space/environments/other/locales", token, http.StatusForbidden)
	request(http.MethodGet, "/spaces/other/environments/environment/locales", token, http.StatusForbidden)
	request(http.MethodGet, "/spaces/space/environments/environment/app_installations/other", token, http.StatusForbidden)
	request(http.MethodDelete, alias+"/app_installations/other", cmt.ValidAccessToken, http.StatusNoContent)
	request(http.MethodGet, environment+"/app_installations/other", cmt.ValidAccessToken, http.StatusNotFound)

	for _, denied := range []struct {
		path   string
		status int
		id     string
	}{
		{"/organizations/org/app_definitions/app", http.StatusNotFound, "NotFound"},
		{"/spaces/space", http.StatusUnauthorized, "AccessTokenInvalid"},
		{environment, http.StatusUnauthorized, "AccessTokenInvalid"},
		{environment + "/live_preview/variables", http.StatusForbidden, "AccessDenied"},
	} {
		response := request(http.MethodGet, denied.path, token, denied.status)

		var failure cm.Error
		require.NoError(t, json.NewDecoder(response.Body).Decode(&failure))
		assert.Equal(t, denied.id, failure.Sys.ID, "GET %s", denied.path)
	}

	server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space", "editorial", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("other")}))
	server.SetEntry("space", "other", "book", entry.Sys.ID, cm.EntryRequest{})
	request(http.MethodGet, aliasEntryPath, token, http.StatusForbidden)
	assert.Equal(t, "book", readAliasEntry(cmt.ValidAccessToken).Sys.ContentType.Sys.ID)
	request(http.MethodGet, alias+"/locales", token, http.StatusForbidden)
	server.SetLocale(cmt.NewLocaleFromData("space", "other", "secondary", cm.LocaleData{Name: "French", Code: "fr"}, false))
	request(http.MethodDelete, alias+"/locales/secondary", token, http.StatusForbidden)
	request(http.MethodGet, "/spaces/space/environments/other/locales/secondary", cmt.ValidAccessToken, http.StatusOK)
}

func appIdentityHTTP(t *testing.T, server *cmt.Server, req *http.Request, bearer string, status int) *httptest.ResponseRecorder {
	t.Helper()

	req.Header.Set("Authorization", "Bearer "+bearer)

	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	require.Equal(t, status, resp.Code, "%s %s", req.Method, req.URL.Path)

	return resp
}
