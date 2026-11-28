package provider

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppAccessTokenOpen(t *testing.T) {
	t.Parallel()

	const (
		jwt          = "signed-jwt-secret"
		issued       = "issued-token-secret"
		issuedResult = `{"token":"issued-token-secret","sys":{"expiresAt":"2026-11-30T12:34:56.789+02:00"}}`
	)

	for _, testcase := range []struct {
		name   string
		status int
		body   string
		err    error
		detail string
	}{
		{
			name:   "ok",
			status: http.StatusOK,
			body:   issuedResult,
		},
		{
			name:   "created",
			status: http.StatusCreated,
			body:   issuedResult,
		},
		{
			name:   "invalid signature",
			status: http.StatusUnauthorized,
			body:   `{"sys":{"type":"Error","id":"AccessTokenInvalid"},"message":"signed-jwt-secret"}`,
			detail: "HTTP 401",
		},
		{
			name:   "forbidden",
			status: http.StatusForbidden,
			body:   `{"sys":{"type":"Error","id":"AccessDenied"},"message":"issued-token-secret"}`,
			detail: "HTTP 403",
		},
		{
			name:   "unreadable gateway response",
			status: http.StatusBadGateway,
			body:   issued,
			detail: "HTTP 502",
		},
		{
			name:   "canceled request",
			err:    fmt.Errorf("%s: %w", jwt, context.Canceled),
			detail: "canceled",
		},
		{
			name:   "request deadline",
			err:    fmt.Errorf("%s: %w", jwt, context.DeadlineExceeded),
			detail: "timed out",
		},
		{
			name:   "network timeout",
			err:    &net.DNSError{Err: jwt, IsTimeout: true},
			detail: "timed out",
		},
		{
			name:   "transport failure",
			err:    fmt.Errorf("%s: %w", jwt, io.ErrUnexpectedEOF),
			detail: "before receiving an HTTP response",
		},
		{
			name:   "missing expiry",
			status: http.StatusOK,
			body:   `{"token":"issued-token-secret","sys":{}}`,
			detail: "HTTP 200",
		},
		{
			name:   "malformed expiry",
			status: http.StatusCreated,
			body:   `{"token":"issued-token-secret","sys":{"expiresAt":"signed-jwt-secret"}}`,
			detail: "HTTP 201",
		},
		{
			name:   "empty token",
			status: http.StatusCreated,
			body:   `{"token":"","sys":{"expiresAt":"2026-11-30T10:34:56Z"}}`,
			detail: "valid token and expiration time",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if testcase.err != nil {
					return nil, testcase.err
				}

				return discoveryHTTPResponse(req, testcase.status, testcase.body), nil
			})
			client, err := cm.NewClient(cm.DefaultServerURL, cm.NewAccessTokenSecuritySource("management-token"), cm.WithClient(&http.Client{Transport: transport}))
			require.NoError(t, err)

			ctx := t.Context()
			r := appAccessTokenEphemeralResource{providerData: ContentfulProviderData{client: client}}

			var schema ephemeral.SchemaResponse
			r.Schema(ctx, ephemeral.SchemaRequest{}, &schema)
			config := tfsdk.Config{
				Schema: schema.Schema,
				Raw: tftypes.NewValue(schema.Schema.Type().TerraformType(ctx), map[string]tftypes.Value{
					"space_id":          tftypes.NewValue(tftypes.String, "space"),
					"environment_id":    tftypes.NewValue(tftypes.String, "environment"),
					"app_definition_id": tftypes.NewValue(tftypes.String, "app"),
					"jwt":               tftypes.NewValue(tftypes.String, jwt),
					"token":             tftypes.NewValue(tftypes.String, nil),
					"expires_at":        tftypes.NewValue(tftypes.String, nil),
				}),
			}
			resp := ephemeral.OpenResponse{Result: tfsdk.EphemeralResultData(config)}
			r.Open(ctx, ephemeral.OpenRequest{Config: config}, &resp)

			if testcase.detail != "" {
				require.Len(t, resp.Diagnostics.Errors(), 1)
				d := resp.Diagnostics.Errors()[0]
				assert.Contains(t, d.Detail(), testcase.detail)
				assert.NotContains(t, d.Summary()+d.Detail(), jwt)
				assert.NotContains(t, d.Summary()+d.Detail(), issued)
				assert.True(t, resp.Result.Raw.Equal(config.Raw))

				return
			}

			require.Empty(t, resp.Diagnostics)

			var values map[string]tftypes.Value
			require.NoError(t, resp.Result.Raw.As(&values))
			assert.Equal(t, tftypes.NewValue(tftypes.String, issued), values["token"])
			assert.Equal(t, tftypes.NewValue(tftypes.String, "2026-11-30T10:34:56.789Z"), values["expires_at"])
		})
	}
}
