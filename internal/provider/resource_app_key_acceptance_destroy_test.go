package provider_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

type testAccAppKeyGetter func(context.Context, cm.GetAppKeyParams) (cm.GetAppKeyRes, error)

var (
	errAppKeyStillExists        = errors.New("app key still exists after destroy")
	errUnexpectedAppKeyResponse = errors.New("unexpected App Key response after destroy")
	errTestAppKeyRequest        = errors.New("request failed")
)

func testAccAppKeyDestroyCheck(getAppKey testAccAppKeyGetter, keyIDs ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		for _, keyID := range keyIDs {
			response, err := getAppKey(ctx, cm.GetAppKeyParams{
				OrganizationID:  testAccAppKeyOrganizationID,
				AppDefinitionID: testAccAppKeyAppDefinitionID,
				KeyKid:          keyID,
			})
			if err != nil {
				return fmt.Errorf("read App Key %q after destroy: %w", keyID, err)
			}

			if status, ok := response.(cm.StatusCodeResponse); ok && status.GetStatusCode() == http.StatusNotFound {
				continue
			}

			if _, ok := response.(*cm.AppKey); ok {
				return fmt.Errorf("%w: %q", errAppKeyStillExists, keyID)
			}

			return fmt.Errorf("%w: %q returned %T", errUnexpectedAppKeyResponse, keyID, response)
		}

		return nil
	}
}

func TestAppKeyDestroyCheck(t *testing.T) {
	t.Parallel()

	keyID := "key"

	t.Run("absent", func(t *testing.T) {
		t.Parallel()

		check := testAccAppKeyDestroyCheck(func(context.Context, cm.GetAppKeyParams) (cm.GetAppKeyRes, error) {
			return cmt.NewContentfulManagementErrorStatusCodeNotFound(new("not found"), nil), nil
		}, keyID)

		require.NoError(t, check(nil))
	})

	t.Run("retained", func(t *testing.T) {
		t.Parallel()

		check := testAccAppKeyDestroyCheck(func(context.Context, cm.GetAppKeyParams) (cm.GetAppKeyRes, error) {
			return &cm.AppKey{}, nil
		}, keyID)

		require.ErrorContains(t, check(nil), "still exists after destroy")
	})

	t.Run("request error", func(t *testing.T) {
		t.Parallel()

		check := testAccAppKeyDestroyCheck(func(context.Context, cm.GetAppKeyParams) (cm.GetAppKeyRes, error) {
			return nil, errTestAppKeyRequest
		}, keyID)

		require.ErrorContains(t, check(nil), "request failed")
	})

	t.Run("unexpected response", func(t *testing.T) {
		t.Parallel()

		check := testAccAppKeyDestroyCheck(func(context.Context, cm.GetAppKeyParams) (cm.GetAppKeyRes, error) {
			return cmt.NewContentfulManagementErrorStatusCodeBadRequest(new("bad request"), nil), nil
		}, keyID)

		require.ErrorContains(t, check(nil), "unexpected")
	})
}
