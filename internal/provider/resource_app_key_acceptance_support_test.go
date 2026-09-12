package provider_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

const (
	testAccAppKeyOrganizationID         = "2zuSjSO4A0e6GKBrhJRe2m"
	testAccAppKeyAppDefinitionID        = "2fxGxOcam8Fo5m1wC11fhn"
	testAccAppKeyOtherOrganizationID    = "other-organization"
	testAccAppKeyOtherAppDefinitionID   = "other-app-definition"
	testAccAppKeyThirdAppDefinitionID   = "third-app-definition"
	testAccAppKeyResourceAddress        = "contentful_app_key.test"
	testAccAppKeyCreateBeforeDestroyHCL = "lifecycle { create_before_destroy = true }"
)

type testAccAppKeyJWKData struct {
	kid string
	x5c string
	x5t string
}

func testAccAppKeyJWK(t *testing.T, index int) testAccAppKeyJWKData {
	t.Helper()

	// Share the mock API's public keys and literal fingerprints across suites.
	data, err := os.ReadFile("../contentful-management-go/testing/testdata/app_key_public_keys.json")
	require.NoError(t, err)

	var keys []cm.AppKeyJWK
	require.NoError(t, json.Unmarshal(data, &keys))
	require.GreaterOrEqual(t, index, 0)
	require.Less(t, index, len(keys))
	jwk := keys[index]
	require.Len(t, jwk.X5c, 1)

	return testAccAppKeyJWKData{
		kid: jwk.Kid,
		x5c: jwk.X5c[0],
		x5t: jwk.X5t,
	}
}

func testAccAppKeyJWKFromDER(publicKeyDER []byte) testAccAppKeyJWKData {
	digest := sha256.Sum256(publicKeyDER)
	fingerprint := base64.RawURLEncoding.EncodeToString(digest[:])

	return testAccAppKeyJWKData{
		kid: fingerprint,
		x5c: base64.StdEncoding.EncodeToString(publicKeyDER),
		x5t: fingerprint,
	}
}

func testAccAppKeyConfig(organizationID, appDefinitionID string, jwk testAccAppKeyJWKData, extra string) string {
	return fmt.Sprintf(`
resource "contentful_app_key" "test" {
  organization_id   = %q
  app_definition_id = %q

  jwk = {
    alg = "RS256"
    kty = "RSA"
    use = "sig"
    kid = %q
    x5c = [%q]
    x5t = %q
  }

  %s
}
`, organizationID, appDefinitionID, jwk.kid, jwk.x5c, jwk.x5t, extra)
}

func testAccAppKeyJWKConfig(jwk testAccAppKeyJWKData) string {
	return testAccAppKeyConfig(
		testAccAppKeyOrganizationID,
		testAccAppKeyAppDefinitionID,
		jwk,
		testAccAppKeyCreateBeforeDestroyHCL,
	)
}

func setTestAccAppKeyAppDefinitions(server *cmt.Server) {
	server.SetAppDefinition(testAccAppKeyOrganizationID, testAccAppKeyAppDefinitionID, cm.AppDefinitionData{
		Name: "Test App",
	})
	server.SetAppDefinition(testAccAppKeyOrganizationID, testAccAppKeyOtherAppDefinitionID, cm.AppDefinitionData{
		Name: "Other Test App",
	})
	server.SetAppDefinition(testAccAppKeyOtherOrganizationID, testAccAppKeyThirdAppDefinitionID, cm.AppDefinitionData{
		Name: "Third Test App",
	})
}
