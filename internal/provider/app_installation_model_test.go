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
		model        provider.AppInstallationModel
		expectErrors bool
		expected     cm.OptString
	}{
		"absent": {
			model:    provider.AppInstallationModel{},
			expected: cm.OptString{},
		},
		"null": {
			model: provider.AppInstallationModel{
				Marketplace: types.SetNull(types.StringType),
			},
			expected: cm.OptString{},
		},
		"unknown": {
			model: provider.AppInstallationModel{
				Marketplace: types.SetUnknown(types.StringType),
			},
			expected: cm.OptString{},
		},
		"empty": {
			model: provider.AppInstallationModel{
				Marketplace: provider.NewEmptySetMust(types.StringType),
			},
			expected: cm.OptString{},
		},
		"foo": {
			model: provider.AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("foo")}),
			},
			expected: cm.NewOptString("foo"),
		},
		"foo,bar": {
			model: provider.AppInstallationModel{
				Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("foo"), types.StringValue("bar")}),
			},
			expected: cm.NewOptString("bar,foo"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, diags := test.model.ToXContentfulMarketplaceHeaderValue(t.Context())

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
			expectedRequestBody: "{\"parameters\":invalid}",
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
		appInstallation cm.AppInstallation
		expectedModel   provider.AppInstallationModel
	}{
		"null": {
			appInstallation: cm.AppInstallation{},
			expectedModel:   provider.AppInstallationModel{},
		},
		"empty": {
			appInstallation: cm.AppInstallation{
				Parameters: []byte("{}"),
			},
			expectedModel: provider.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedValue("{}"),
			},
		},
		"foo=bar": {
			appInstallation: cm.AppInstallation{
				Parameters: []byte("{\"foo\":\"bar\"}"),
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
