package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestToXContentfulMarketplaceHeaderValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		model        provider.AppInstallationResourceModel
		expectErrors bool
		expected     cm.OptString
	}{
		"absent": {
			model:    provider.AppInstallationResourceModel{},
			expected: cm.OptString{},
		},
		"null": {
			model: provider.AppInstallationResourceModel{
				Marketplace: types.SetNull(types.StringType),
			},
			expected: cm.OptString{},
		},
		"unknown": {
			model: provider.AppInstallationResourceModel{
				Marketplace: types.SetUnknown(types.StringType),
			},
			expected: cm.OptString{},
		},
		"empty": {
			model: provider.AppInstallationResourceModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{}),
			},
			expected: cm.OptString{},
		},
		"foo": {
			model: provider.AppInstallationResourceModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("foo")}),
			},
			expected: cm.NewOptString("foo"),
		},
		"foo,bar": {
			model: provider.AppInstallationResourceModel{
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

func TestToAppInstallationFields(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		model               provider.AppInstallationResourceModel
		expectErrors        bool
		expectWarnings      bool
		expectedRequestBody string
	}{
		"null": {
			model: provider.AppInstallationResourceModel{
				Parameters: jsontypes.NewNormalizedNull(),
			},
			expectedRequestBody: "{}",
		},
		"unknown": {
			model: provider.AppInstallationResourceModel{
				Parameters: jsontypes.NewNormalizedUnknown(),
			},
			expectWarnings:      true,
			expectedRequestBody: "{}",
		},
		"empty": {
			model: provider.AppInstallationResourceModel{
				Parameters: jsontypes.NewNormalizedValue("{}"),
			},
			expectedRequestBody: "{\"parameters\":{}}",
		},
		"foo=bar": {
			model: provider.AppInstallationResourceModel{
				Parameters: jsontypes.NewNormalizedValue("{\"foo\":\"bar\"}"),
			},
			expectedRequestBody: "{\"parameters\":{\"foo\":\"bar\"}}",
		},
		"invalid": {
			model: provider.AppInstallationResourceModel{
				Parameters: jsontypes.NewNormalizedValue("invalid"),
			},
			expectedRequestBody: "{\"parameters\":invalid}",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, diags := test.model.ToAppInstallationFields()

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
