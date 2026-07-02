package provider_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestAccProtoV6ProviderFactories(options ...Option) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"contentful": providerserver.NewProtocol6WithError(Factory("test", options...)()),
	}
}

var testAccProtoV6ProviderFactories = makeTestAccProtoV6ProviderFactories()

func providerConfigDynamicValue(config map[string]any) (tfprotov6.DynamicValue, error) {
	providerConfigTypes := map[string]tftypes.Type{
		"url":          tftypes.String,
		"access_token": tftypes.String,
	}
	providerConfigObjectType := tftypes.Object{AttributeTypes: providerConfigTypes}

	providerConfigObjectValue := tftypes.NewValue(providerConfigObjectType, map[string]tftypes.Value{
		"url":          tftypes.NewValue(tftypes.String, config["url"]),
		"access_token": tftypes.NewValue(tftypes.String, config["access_token"]),
	})

	value, err := tfprotov6.NewDynamicValue(providerConfigObjectType, providerConfigObjectValue)
	if err != nil {
		err = fmt.Errorf("failed to create dynamic value: %w", err)
	}

	return value, err
}

func TestProtocol6ProviderServerSchemaVersion(t *testing.T) {
	t.Parallel()

	providerServer, err := testAccProtoV6ProviderFactories["contentful"]()
	require.NotNil(t, providerServer)
	require.NoError(t, err)

	resp, err := providerServer.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NotNil(t, resp.Provider)
	require.NoError(t, err)
	assert.Empty(t, resp.Diagnostics)

	assert.EqualValues(t, 0, resp.Provider.Version)
}

func TestProtocol6ProviderServerSchemaDocumentsProviderConfiguration(t *testing.T) {
	t.Parallel()

	providerServer, err := testAccProtoV6ProviderFactories["contentful"]()
	require.NotNil(t, providerServer)
	require.NoError(t, err)

	resp, err := providerServer.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NotNil(t, resp.Provider)
	require.NoError(t, err)
	require.Empty(t, resp.Diagnostics)

	attributes := map[string]*tfprotov6.SchemaAttribute{}
	for _, attribute := range resp.Provider.Block.Attributes {
		attributes[attribute.Name] = attribute
	}

	require.Contains(t, attributes, "url")
	assert.Contains(t, attributes["url"].Description, "CONTENTFUL_URL")
	assert.Contains(t, attributes["url"].Description, "public Contentful Management API")

	require.Contains(t, attributes, "access_token")
	assert.True(t, attributes["access_token"].Sensitive)
	assert.Contains(t, attributes["access_token"].Description, "CONTENTFUL_MANAGEMENT_ACCESS_TOKEN")
}

func TestProtocol6ProviderServerConfigure(t *testing.T) {
	if os.Getenv("TF_ACC") != "" {
		return
	}

	tests := map[string]struct {
		config                    map[string]any
		env                       map[string]string
		expectedSuccess           bool
		expectedDiagnosticSummary string
	}{
		"config: url": {
			config: map[string]any{
				"url": "https://api.test.contentful.com",
			},
			expectedSuccess:           false,
			expectedDiagnosticSummary: "Missing Contentful management access token",
		},
		"config: access_token": {
			config: map[string]any{
				"access_token": "CFPAT-12345",
			},
			expectedSuccess: true,
		},
		"config: url,access_token": {
			config: map[string]any{
				"url":          "https://api.test.contentful.com",
				"access_token": "CFPAT-12345",
			},
			expectedSuccess: true,
		},
		"config: url(invalid),access_token": {
			config: map[string]any{
				"url":          "url://an invalid url %/",
				"access_token": "CFPAT-12345",
			},
			expectedSuccess:           false,
			expectedDiagnosticSummary: "Invalid Contentful API URL",
		},
		"config: url(relative),access_token": {
			config: map[string]any{
				"url":          "/relative",
				"access_token": "CFPAT-12345",
			},
			expectedSuccess:           false,
			expectedDiagnosticSummary: "Invalid Contentful API URL",
		},
		"config: url(unsupported-scheme),access_token": {
			config: map[string]any{
				"url":          "ftp://api.test.contentful.com",
				"access_token": "CFPAT-12345",
			},
			expectedSuccess:           false,
			expectedDiagnosticSummary: "Invalid Contentful API URL",
		},
		"env: url": {
			env: map[string]string{
				"CONTENTFUL_URL": "https://api.test.contentful.com",
			},
			expectedSuccess:           false,
			expectedDiagnosticSummary: "Missing Contentful management access token",
		},
		"env: url,access_token": {
			env: map[string]string{
				"CONTENTFUL_URL":                     "https://api.test.contentful.com",
				"CONTENTFUL_MANAGEMENT_ACCESS_TOKEN": "CFPAT-12345",
			},
			expectedSuccess: true,
		},
		"config: url env: access_token": {
			config: map[string]any{
				"url": "https://api.test.contentful.com",
			},
			env: map[string]string{
				"CONTENTFUL_MANAGEMENT_ACCESS_TOKEN": "CFPAT-12345",
			},
			expectedSuccess: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for key, value := range test.env {
				t.Setenv(key, value)
			}

			providerServer, err := testAccProtoV6ProviderFactories["contentful"]()
			require.NotNil(t, providerServer)
			require.NoError(t, err)

			providerConfigValue, err := providerConfigDynamicValue(test.config)
			require.NotNil(t, providerConfigValue)
			require.NoError(t, err)

			resp, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{
				Config: &providerConfigValue,
			})
			require.NotNil(t, resp)
			require.NoError(t, err)

			if test.expectedSuccess {
				assert.Empty(t, resp.Diagnostics)
			} else {
				assert.NotEmpty(t, resp.Diagnostics)
				assertDiagnosticSummary(t, resp.Diagnostics, test.expectedDiagnosticSummary)
			}
		})
	}
}

func assertDiagnosticSummary(t *testing.T, diagnostics []*tfprotov6.Diagnostic, expectedSummary string) {
	t.Helper()

	if expectedSummary == "" {
		return
	}

	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic.Summary, expectedSummary) {
			return
		}
	}

	t.Fatalf("expected diagnostic summary %q in %#v", expectedSummary, diagnostics)
}
