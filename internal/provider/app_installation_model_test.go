package provider_test

import (
	"context"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestToXContentfulMarketplaceHeaderValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		model        provider.AppInstallationModel
		expectErrors bool
		expected     contentfulManagement.OptString
	}{
		"absent": {
			model:    provider.AppInstallationModel{},
			expected: contentfulManagement.OptString{},
		},
		"null": {
			model: provider.AppInstallationModel{
				Marketplace: types.SetNull(types.StringType),
			},
			expected: contentfulManagement.OptString{},
		},
		"unknown": {
			model: provider.AppInstallationModel{
				Marketplace: types.SetUnknown(types.StringType),
			},
			expected: contentfulManagement.OptString{},
		},
		"empty": {
			model: provider.AppInstallationModel{
				Marketplace: util.NewEmptySetMust(types.StringType),
			},
			expected: contentfulManagement.OptString{},
		},
		"foo": {
			model: provider.AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("foo")}),
			},
			expected: contentfulManagement.NewOptString("foo"),
		},
		"foo,bar": {
			model: provider.AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("foo"), types.StringValue("bar")}),
			},
			expected: contentfulManagement.NewOptString("bar,foo"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, diags := test.model.ToXContentfulMarketplaceHeaderValue(context.Background())

			assert.EqualValues(t, test.expected, value)

			if test.expectErrors {
				assert.NotEmpty(t, diags.Errors())
			} else {
				assert.Empty(t, diags.Errors())
			}
		})
	}
}

func TestToPutAppInstallationReq(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		model               provider.AppInstallationModel
		expectErrors        bool
		expectWarnings      bool
		expectedRequestBody string
	}{
		"null": {
			model: provider.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedNull(),
			},
			expectedRequestBody: "{}",
		},
		"unknown": {
			model: provider.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedUnknown(),
			},
			expectWarnings:      true,
			expectedRequestBody: "{}",
		},
		"empty": {
			model: provider.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedValue("{}"),
			},
			expectedRequestBody: "{\"parameters\":{}}",
		},
		"foo=bar": {
			model: provider.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedValue("{\"foo\":\"bar\"}"),
			},
			expectedRequestBody: "{\"parameters\":{\"foo\":\"bar\"}}",
		},
		"invalid": {
			model: provider.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedValue("invalid"),
			},
			expectErrors:        true,
			expectedRequestBody: "{\"parameters\":{}}",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req, diags := test.model.ToPutAppInstallationReq()

			requestBody, _ := req.MarshalJSON()

			assert.EqualValues(t, test.expectedRequestBody, string(requestBody))

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

func TestAppInstallationModelReadFromResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		appInstallation contentfulManagement.AppInstallation
		expectedModel   provider.AppInstallationModel
	}{
		"null": {
			appInstallation: contentfulManagement.AppInstallation{},
			expectedModel:   provider.AppInstallationModel{},
		},
		"empty": {
			appInstallation: contentfulManagement.AppInstallation{
				Parameters: contentfulManagement.NewOptAppInstallationParameters(contentfulManagement.AppInstallationParameters{}),
			},
			expectedModel: provider.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedValue("{}"),
			},
		},
		"foo=bar": {
			appInstallation: contentfulManagement.AppInstallation{
				Parameters: contentfulManagement.NewOptAppInstallationParameters(contentfulManagement.AppInstallationParameters{
					"foo": []byte{'"', 'b', 'a', 'r', '"'},
				}),
			},
			expectedModel: provider.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedValue("{\"foo\":\"bar\"}"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := provider.AppInstallationModel{}

			diags := model.ReadFromResponse(&test.appInstallation)

			assert.EqualValues(t, test.expectedModel, model)

			assert.Empty(t, diags)
		})
	}
}
