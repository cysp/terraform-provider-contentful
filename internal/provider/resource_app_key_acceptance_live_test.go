package provider_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

//nolint:paralleltest
func TestAccAppKeyResourceLiveLifecycle(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for live App Key tests")
	}

	if os.Getenv("TF_ACC_MOCKED") != "" {
		t.Skip("live App Key lifecycle is covered by the equal mock acceptance sibling")
	}

	jwk := testAccAppKeyJWK(t)
	replacementJWK := testAccAppKeyJWK(t)

	cleanupLiveAppKeyFixture(t, jwk.kid, replacementJWK.kid)
	client := newLiveAppKeyClient(t)

	ContentfulProviderMockableResourceTest(t, nil, resource.TestCase{
		CheckDestroy: testAccAppKeyDestroyCheck(
			func(ctx context.Context, params cm.GetAppKeyParams) (cm.GetAppKeyRes, error) {
				return client.GetAppKey(ctx, params)
			},
			jwk.kid,
			replacementJWK.kid,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccAppKeyJWKConfig(jwk),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "key_kid", jwk.kid),
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.kid", jwk.kid),
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.x5c.0", jwk.x5c),
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.x5t", jwk.x5t),
				),
			},
			{
				ResourceName:    testAccAppKeyResourceAddress,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
			{
				Config: testAccAppKeyJWKConfig(replacementJWK),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(testAccAppKeyResourceAddress, plancheck.ResourceActionReplace),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "key_kid", replacementJWK.kid),
					resource.TestCheckResourceAttr(testAccAppKeyResourceAddress, "jwk.kid", replacementJWK.kid),
				),
			},
		},
	})
}

func cleanupLiveAppKeyFixture(t *testing.T, ownedKeyIDs ...string) {
	t.Helper()

	initialKeyIDs := liveAppKeyIDs(listLiveAppKeys(t))

	for _, keyID := range ownedKeyIDs {
		if _, alreadyExists := initialKeyIDs[keyID]; alreadyExists {
			t.Fatalf("generated App Key fingerprint %q already exists before the test", keyID)
		}
	}

	const (
		maximumKeysPerAppDefinition = 3
		requiredFreeSlots           = 2
	)

	if len(initialKeyIDs) > maximumKeysPerAppDefinition-requiredFreeSlots {
		t.Fatalf(
			"App Key fixture has %d keys; the live lifecycle requires %d free slots",
			len(initialKeyIDs),
			requiredFreeSlots,
		)
	}

	t.Cleanup(func() {
		deleteLiveAppKeys(t, ownedKeyIDs)
		assertLiveAppKeyCleanup(t, initialKeyIDs, ownedKeyIDs, liveAppKeyIDs(listLiveAppKeys(t)))
	})
}

func listLiveAppKeys(t *testing.T) []cm.AppKey {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	client := newLiveAppKeyClient(t)

	response, err := client.GetAppKeys(ctx, cm.GetAppKeysParams{
		OrganizationID:  testAccAppKeyOrganizationID,
		AppDefinitionID: testAccAppKeyAppDefinitionID,
	})
	if err != nil {
		t.Fatalf("list App Keys during cleanup: %v", err)
	}

	keys, ok := response.(*cm.AppKeyCollection)
	if !ok {
		t.Fatalf("unexpected App Key cleanup list response: %T", response)
	}

	return keys.Items
}

func liveAppKeyIDs(keys []cm.AppKey) map[string]struct{} {
	keyIDs := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		keyIDs[key.Sys.ID] = struct{}{}
	}

	return keyIDs
}

func deleteLiveAppKeys(t *testing.T, keyIDs []string) {
	t.Helper()

	client := newLiveAppKeyClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for _, keyID := range keyIDs {
		deleteResponse, err := client.DeleteAppKey(ctx, cm.DeleteAppKeyParams{
			OrganizationID:  testAccAppKeyOrganizationID,
			AppDefinitionID: testAccAppKeyAppDefinitionID,
			KeyKid:          keyID,
		})
		if err != nil {
			t.Errorf("delete App Key %q during cleanup: %v", keyID, err)

			continue
		}

		if _, ok := deleteResponse.(*cm.NoContent); ok {
			continue
		}

		status, ok := deleteResponse.(cm.StatusCodeResponse)
		if !ok || status.GetStatusCode() != http.StatusNotFound {
			t.Errorf("unexpected App Key cleanup delete response for %q: %T", keyID, deleteResponse)
		}
	}
}

func assertLiveAppKeyCleanup(t *testing.T, initialKeyIDs map[string]struct{}, ownedKeyIDs []string, remainingKeyIDs map[string]struct{}) {
	t.Helper()

	for _, keyID := range ownedKeyIDs {
		if _, remains := remainingKeyIDs[keyID]; remains {
			t.Errorf("App Key cleanup left test-owned key %q", keyID)
		}
	}

	for keyID := range initialKeyIDs {
		if _, remains := remainingKeyIDs[keyID]; !remains {
			t.Errorf("App Key cleanup removed pre-existing key %q", keyID)
		}
	}
}

func newLiveAppKeyClient(t *testing.T) *cm.Client {
	t.Helper()

	accessToken := os.Getenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN")
	if accessToken == "" {
		t.Fatal("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN must be set for live App Key tests")
	}

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil
	retryClient.RetryMax = 5

	client, err := cm.NewClient(
		cm.DefaultServerURL,
		cm.NewAccessTokenSecuritySource(accessToken),
		cm.WithClient(cm.NewTransportClient(retryClient.StandardClient(), cm.DefaultUserAgent)),
	)
	if err != nil {
		t.Fatalf("create live App Key client: %v", err)
	}

	return client
}
