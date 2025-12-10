package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestToXContentfulMarketplaceHeaderValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		model        AppInstallationModel
		expectErrors bool
		expected     cm.OptString
	}{
		"absent": {
			model:    AppInstallationModel{},
			expected: cm.OptString{},
		},
		"null": {
			model: AppInstallationModel{
				Marketplace: types.SetNull(types.StringType),
			},
			expected: cm.OptString{},
		},
		"unknown": {
			model: AppInstallationModel{
				Marketplace: types.SetUnknown(types.StringType),
			},
			expected: cm.OptString{},
		},
		"empty": {
			model: AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{}),
			},
			expected: cm.OptString{},
		},
		"foo": {
			model: AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("foo")}),
			},
			expected: cm.NewOptString("foo"),
		},
		"foo,bar": {
			model: AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("foo"), types.StringValue("bar")}),
			},
			expected: cm.NewOptString("bar,foo"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, diags := test.model.ToXContentfulMarketplaceHeaderValue(t.Context())

			assert.Equal(t, test.expected, value)

			if test.expectErrors {
				assert.NotEmpty(t, diags.Errors())
			} else {
				assert.Empty(t, diags.Errors())
			}
		})
	}
}

func TestToAppInstallationData(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		model               AppInstallationModel
		expectErrors        bool
		expectWarnings      bool
		expectedRequestBody string
	}{
		"null": {
			model: AppInstallationModel{
				Parameters: jsontypes.NewNormalizedNull(),
			},
			expectedRequestBody: "{}",
		},
		"unknown": {
			model: AppInstallationModel{
				Parameters: jsontypes.NewNormalizedUnknown(),
			},
			expectWarnings:      true,
			expectedRequestBody: "{}",
		},
		"empty": {
			model: AppInstallationModel{
				Parameters: NewNormalizedJSONTypesNormalizedValue([]byte("{}")),
			},
			expectedRequestBody: "{\"parameters\":{}}",
		},
		"foo=bar": {
			model: AppInstallationModel{
				Parameters: NewNormalizedJSONTypesNormalizedValue([]byte("{\"foo\":\"bar\"}")),
			},
			expectedRequestBody: "{\"parameters\":{\"foo\":\"bar\"}}",
		},
		"invalid": {
			model: AppInstallationModel{
				Parameters: NewNormalizedJSONTypesNormalizedValue([]byte("invalid")),
			},
			expectedRequestBody: "{\"parameters\":invalid}",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, diags := test.model.ToAppInstallationData()

			requestBody, _ := req.MarshalJSON()

			assert.Equal(t, test.expectedRequestBody, string(requestBody))

			if test.expectErrors {
				assert.NotEmpty(t, diags.Errors())
			} else {
				assert.Empty(t, diags.Errors())
			}

			if test.expectWarnings {
				assert.NotEmpty(t, diags.Warnings())
			} else {
				assert.Empty(t, diags.Warnings())
			}
		})
	}
}
