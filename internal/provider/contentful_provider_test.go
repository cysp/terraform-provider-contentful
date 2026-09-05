package provider_test

import (
	"fmt"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

var providerConfigType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"url":          tftypes.String,
	"access_token": tftypes.String,
}}

func providerConfigValue(config map[string]any) tftypes.Value {
	return tftypes.NewValue(providerConfigType, map[string]tftypes.Value{
		"url":          tftypes.NewValue(tftypes.String, config["url"]),
		"access_token": tftypes.NewValue(tftypes.String, config["access_token"]),
	})
}

func providerConfigDynamicValue(config map[string]any) (tfprotov6.DynamicValue, error) {
	value, err := tfprotov6.NewDynamicValue(providerConfigType, providerConfigValue(config))
	if err != nil {
		err = fmt.Errorf("failed to create dynamic value: %w", err)
	}

	return value, err
}

type providerDiagnosticExpectation struct {
	summary   string
	detail    string
	attribute *tftypes.AttributePath
}

func providerAttributeError(summary, detail, attribute string) providerDiagnosticExpectation {
	return providerDiagnosticExpectation{
		summary:   summary,
		detail:    detail,
		attribute: tftypes.NewAttributePath().WithAttributeName(attribute),
	}
}

const (
	missingAccessTokenDetail   = "Set the access_token provider attribute or the CONTENTFUL_MANAGEMENT_ACCESS_TOKEN environment variable."
	invalidContentfulURLDetail = "The url provider attribute must be an absolute HTTP or HTTPS URL, such as https://api.contentful.com. " +
		"It can also be set with the CONTENTFUL_URL environment variable."
	unsupportedContentfulURLSchemeDetail = "The url provider attribute must use the http or https scheme."
	unknownContentfulURLDetail           = "The provider cannot create the Contentful client because the configured API URL is unknown. " +
		"Apply the source of the value first or use a known URL."
	unknownAccessTokenDetail = "The provider cannot create the Contentful client because the configured management access token is unknown. " +
		"Apply the source of the value first or use a known access token."
)

func TestProtocol6ProviderServerSchemaVersion(t *testing.T) {
	t.Parallel()

	providerServer, err := testAccProtoV6ProviderFactories["contentful"]()
	require.NotNil(t, providerServer)
	require.NoError(t, err)

	resp, err := providerServer.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Provider)
	assert.Empty(t, resp.Diagnostics)

	assert.EqualValues(t, 0, resp.Provider.Version)
}

func TestProtocol6ProviderServerSchemaDocumentsProviderConfiguration(t *testing.T) {
	t.Parallel()

	providerServer, err := testAccProtoV6ProviderFactories["contentful"]()
	require.NotNil(t, providerServer)
	require.NoError(t, err)

	resp, err := providerServer.GetProviderSchema(t.Context(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Provider)
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
	tests := map[string]struct {
		config              map[string]any
		env                 map[string]string
		options             []Option
		expectedDiagnostics []providerDiagnosticExpectation
	}{
		"config: url": {
			config: map[string]any{
				"url": "https://api.test.contentful.com",
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Missing Contentful management access token", missingAccessTokenDetail, "access_token"),
			},
		},
		"config: access_token": {
			config: map[string]any{
				"access_token": "CFPAT-12345",
			},
		},
		"config: url(invalid),access_token": {
			config: map[string]any{
				"url":          "url://an invalid url %/",
				"access_token": "CFPAT-12345",
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Invalid Contentful API URL", invalidContentfulURLDetail, "url"),
			},
		},
		"config: url(relative),access_token": {
			config: map[string]any{
				"url":          "/relative",
				"access_token": "CFPAT-12345",
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Invalid Contentful API URL", invalidContentfulURLDetail, "url"),
			},
		},
		"config: url(empty-host),access_token": {
			config: map[string]any{
				"url":          "http://:80",
				"access_token": "CFPAT-12345",
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Invalid Contentful API URL", invalidContentfulURLDetail, "url"),
			},
		},
		"config: url(unsupported-scheme),access_token": {
			config: map[string]any{
				"url":          "ftp://api.test.contentful.com",
				"access_token": "CFPAT-12345",
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Invalid Contentful API URL", unsupportedContentfulURLSchemeDetail, "url"),
			},
		},
		"env: url": {
			env: map[string]string{
				"CONTENTFUL_URL": "https://api.test.contentful.com",
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Missing Contentful management access token", missingAccessTokenDetail, "access_token"),
			},
		},
		"env: url,access_token": {
			env: map[string]string{
				"CONTENTFUL_URL":                     "https://api.test.contentful.com",
				"CONTENTFUL_MANAGEMENT_ACCESS_TOKEN": "CFPAT-12345",
			},
		},
		"env: url(invalid),access_token": {
			env: map[string]string{
				"CONTENTFUL_URL":                     "http://:80",
				"CONTENTFUL_MANAGEMENT_ACCESS_TOKEN": "CFPAT-12345",
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Invalid Contentful API URL", invalidContentfulURLDetail, "url"),
			},
		},
		"config: url env: access_token": {
			config: map[string]any{
				"url": "https://api.test.contentful.com",
			},
			env: map[string]string{
				"CONTENTFUL_MANAGEMENT_ACCESS_TOKEN": "CFPAT-12345",
			},
		},
		"config takes precedence over env": {
			config: map[string]any{
				"url":          "https://api.test.contentful.com",
				"access_token": "CFPAT-12345",
			},
			env: map[string]string{
				"CONTENTFUL_URL": "url://an invalid url %/",
			},
		},
		"known empty config takes precedence over env": {
			config: map[string]any{
				"url":          "",
				"access_token": "",
			},
			env: map[string]string{
				"CONTENTFUL_URL":                     "url://an invalid url %/",
				"CONTENTFUL_MANAGEMENT_ACCESS_TOKEN": "CFPAT-12345",
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Missing Contentful management access token", missingAccessTokenDetail, "access_token"),
			},
		},
		"unknown url": {
			config: map[string]any{
				"url":          tftypes.UnknownValue,
				"access_token": "CFPAT-12345",
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Unknown Contentful API URL", unknownContentfulURLDetail, "url"),
			},
		},
		"unknown access_token": {
			config: map[string]any{
				"url":          "https://api.test.contentful.com",
				"access_token": tftypes.UnknownValue,
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Unknown Contentful management access token", unknownAccessTokenDetail, "access_token"),
			},
		},
		"unknown config takes precedence over env": {
			config: map[string]any{
				"url":          tftypes.UnknownValue,
				"access_token": tftypes.UnknownValue,
			},
			env: map[string]string{
				"CONTENTFUL_URL":                     "https://api.test.contentful.com",
				"CONTENTFUL_MANAGEMENT_ACCESS_TOKEN": "CFPAT-12345",
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Unknown Contentful API URL", unknownContentfulURLDetail, "url"),
				providerAttributeError("Unknown Contentful management access token", unknownAccessTokenDetail, "access_token"),
			},
		},
		"options take precedence over unknown config and env": {
			config: map[string]any{
				"url":          tftypes.UnknownValue,
				"access_token": tftypes.UnknownValue,
			},
			env: map[string]string{
				"CONTENTFUL_URL": "url://an invalid url %/",
			},
			options: []Option{
				WithContentfulURL("https://api.test.contentful.com"),
				WithAccessToken("CFPAT-override"),
			},
		},
		"invalid url option takes precedence over config and env": {
			config: map[string]any{
				"url":          "https://api.test.contentful.com",
				"access_token": "CFPAT-12345",
			},
			env: map[string]string{
				"CONTENTFUL_URL": "https://api.test.contentful.com",
			},
			options: []Option{
				WithContentfulURL("url://an invalid url %/"),
			},
			expectedDiagnostics: []providerDiagnosticExpectation{
				providerAttributeError("Invalid Contentful API URL", invalidContentfulURLDetail, "url"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("CONTENTFUL_URL", "")
			t.Setenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN", "")

			for key, value := range test.env {
				t.Setenv(key, value)
			}

			providerServer, err := makeTestAccProtoV6ProviderFactories(test.options...)["contentful"]()
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

			require.Len(t, resp.Diagnostics, len(test.expectedDiagnostics))

			for i, diagnostic := range resp.Diagnostics {
				expected := test.expectedDiagnostics[i]

				assert.Equal(t, expected.summary, diagnostic.Summary)
				assert.Equal(t, expected.detail, diagnostic.Detail)
				assert.Equal(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity)
				assert.Equal(t, expected.attribute, diagnostic.Attribute)
			}
		})
	}
}

func TestProviderConfigureErrorsDoNotSetProviderData(t *testing.T) {
	t.Parallel()

	for name, url := range map[string]any{
		"unknown configuration": tftypes.UnknownValue,
		"invalid URL":           "url://an invalid url %/",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			providerImplementation := New("test")
			schemaResponse := &frameworkprovider.SchemaResponse{}
			providerImplementation.Schema(t.Context(), frameworkprovider.SchemaRequest{}, schemaResponse)
			config := tfsdk.Config{
				Schema: schemaResponse.Schema,
				Raw: providerConfigValue(map[string]any{
					"url":          url,
					"access_token": "CFPAT-12345",
				}),
			}
			response := &frameworkprovider.ConfigureResponse{}

			providerImplementation.Configure(
				t.Context(),
				frameworkprovider.ConfigureRequest{Config: config},
				response,
			)

			require.True(t, response.Diagnostics.HasError())
			assert.Nil(t, response.ActionData)
			assert.Nil(t, response.DataSourceData)
			assert.Nil(t, response.ListResourceData)
			assert.Nil(t, response.ResourceData)
		})
	}
}

func TestProtocol6ProviderServerConfigureOverridesUnknownValues(t *testing.T) {
	t.Parallel()

	providerFactories := makeTestAccProtoV6ProviderFactories(
		WithContentfulURL("https://api.test.contentful.com"),
		WithAccessToken("CFPAT-override"),
	)
	providerServer, err := providerFactories["contentful"]()
	require.NoError(t, err)

	providerConfigValue, err := providerConfigDynamicValue(map[string]any{
		"url":          tftypes.UnknownValue,
		"access_token": tftypes.UnknownValue,
	})
	require.NoError(t, err)

	resp, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{
		Config: &providerConfigValue,
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Diagnostics)
}
