package provider_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
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

func testAccAppKeyJWK(t *testing.T) testAccAppKeyJWKData {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)

	return testAccAppKeyJWKFromDER(publicKeyDER)
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
