package provider_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

//nolint:ireturn
func configuredProtocolServer(t *testing.T, handler http.Handler) tfprotov6.ProviderServer {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	providerServer, err := makeTestAccProtoV6ProviderFactories(WithContentfulURL(server.URL), WithAccessToken(cmt.ValidAccessToken))["contentful"]()
	require.NoError(t, err)
	config, err := providerConfigDynamicValue(map[string]any{"url": server.URL, "access_token": cmt.ValidAccessToken})
	require.NoError(t, err)
	response, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{Config: &config})
	require.NoError(t, err)
	require.Empty(t, response.Diagnostics)

	return providerServer
}
