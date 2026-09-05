//nolint:testpackage // Verify the configured client and shared counter at the actual Configure entrypoints.
package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderDataConfiguration(t *testing.T) {
	t.Parallel()

	prior := ContentfulProviderData{client: &cm.Client{}, editorInterfaceVersionOffset: &ContentfulContentTypeCounter{}}
	configured := ContentfulProviderData{client: &cm.Client{}, editorInterfaceVersionOffset: &ContentfulContentTypeCounter{}}

	var nilPointer *ContentfulProviderData

	for name, test := range map[string]struct {
		input     any
		want      ContentfulProviderData
		wantError bool
	}{
		"nil preserves prior configuration":        {input: nil, want: prior},
		"configured value replaces both fields":    {input: configured, want: configured},
		"zero value replaces prior configuration":  {input: ContentfulProviderData{}, want: ContentfulProviderData{}},
		"wrong type preserves prior configuration": {input: "invalid", want: prior, wantError: true},
		"pointer is not the configured value type": {input: &configured, want: prior, wantError: true},
		"typed nil is not absent provider data":    {input: nilPointer, want: prior, wantError: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			check := func(t *testing.T, got ContentfulProviderData, diagnostics diag.Diagnostics) {
				t.Helper()

				if test.want.client == nil {
					assert.Nil(t, got.client)
				} else {
					assert.Same(t, test.want.client, got.client)
				}

				if test.want.editorInterfaceVersionOffset == nil {
					assert.Nil(t, got.editorInterfaceVersionOffset)
				} else {
					assert.Same(t, test.want.editorInterfaceVersionOffset, got.editorInterfaceVersionOffset)
				}

				if test.wantError {
					require.Len(t, diagnostics, 1)
					assert.Equal(t, diag.SeverityError, diagnostics[0].Severity())
					assert.Equal(t, "Invalid provider data", diagnostics[0].Summary())
					assert.Empty(t, diagnostics[0].Detail())
				} else {
					assert.Empty(t, diagnostics)
				}
			}

			t.Run("helper", func(t *testing.T) {
				t.Parallel()

				got := prior
				diagnostics := setProviderData(test.input, &got)
				check(t, got, diagnostics)
			})

			t.Run("resource", func(t *testing.T) {
				t.Parallel()

				implementation := appDefinitionResource{providerData: prior}

				var response resource.ConfigureResponse
				implementation.Configure(t.Context(), resource.ConfigureRequest{ProviderData: test.input}, &response)
				check(t, implementation.providerData, response.Diagnostics)
			})

			t.Run("data source", func(t *testing.T) {
				t.Parallel()

				implementation := appDefinitionDataSource{providerData: prior}

				var response datasource.ConfigureResponse
				implementation.Configure(t.Context(), datasource.ConfigureRequest{ProviderData: test.input}, &response)
				check(t, implementation.providerData, response.Diagnostics)
			})

			t.Run("list resource", func(t *testing.T) {
				t.Parallel()

				implementation := entryListResource{providerData: prior}

				var response resource.ConfigureResponse
				implementation.Configure(t.Context(), resource.ConfigureRequest{ProviderData: test.input}, &response)
				check(t, implementation.providerData, response.Diagnostics)
			})
		})
	}
}

func TestProviderDataConfigureRequestsPreserveGenericType(t *testing.T) {
	t.Parallel()

	var nilString *string

	for name, test := range map[string]struct {
		input     any
		want      string
		wantError bool
	}{
		"nil leaves output unchanged":        {input: nil, want: "prior"},
		"known string replaces output":       {input: "configured", want: "configured"},
		"empty string replaces output":       {input: "", want: ""},
		"other type leaves output unchanged": {input: 42, want: "prior", wantError: true},
		"typed nil is a different type":      {input: nilString, want: "prior", wantError: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resourceData, dataSourceData := "prior", "prior"
			resourceDiagnostics := SetProviderDataFromResourceConfigureRequest(resource.ConfigureRequest{ProviderData: test.input}, &resourceData)
			dataSourceDiagnostics := SetProviderDataFromDataSourceConfigureRequest(datasource.ConfigureRequest{ProviderData: test.input}, &dataSourceData)

			assert.Equal(t, test.want, resourceData)
			assert.Equal(t, test.want, dataSourceData)

			for _, diagnostics := range []diag.Diagnostics{resourceDiagnostics, dataSourceDiagnostics} {
				if test.wantError {
					require.Len(t, diagnostics, 1)
					assert.Equal(t, diag.SeverityError, diagnostics[0].Severity())
					assert.Equal(t, "Invalid provider data", diagnostics[0].Summary())
					assert.Empty(t, diagnostics[0].Detail())
				} else {
					assert.Empty(t, diagnostics)
				}
			}
		})
	}
}
