package util_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

func TestProviderDataFromDataSourceConfigureRequest(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		providerData    any
		expectedSuccess bool
	}{
		"nil": {
			providerData: nil,
		},
		"string": {
			providerData:    "123",
			expectedSuccess: true,
		},
		"number": {
			providerData: 123,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var (
				providerData string
				resp         datasource.ConfigureResponse
			)

			result := util.ProviderDataFromDataSourceConfigureRequest(datasource.ConfigureRequest{
				ProviderData: test.providerData,
			}, &providerData, &resp)

			assert.EqualValues(t, test.expectedSuccess, result)

			if test.expectedSuccess {
				assert.Empty(t, resp.Diagnostics.Errors())
			}
		})
	}
}

func TestProviderDataFromResourceeConfigureRequest(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		providerData    any
		expectedSuccess bool
	}{
		"nil": {
			providerData: nil,
		},
		"string": {
			providerData:    "123",
			expectedSuccess: true,
		},
		"number": {
			providerData: 123,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var (
				providerData string
				resp         resource.ConfigureResponse
			)

			result := util.ProviderDataFromResourceConfigureRequest(resource.ConfigureRequest{
				ProviderData: test.providerData,
			}, &providerData, &resp)

			assert.EqualValues(t, test.expectedSuccess, result)

			if test.expectedSuccess {
				assert.EqualValues(t, test.providerData, providerData)
				assert.Empty(t, resp.Diagnostics.Errors())
			}
		})
	}
}
