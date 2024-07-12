package provider_test

import (
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/cysp/terraform-provider-contentful/internal/provider/resource_app_installation"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/stretchr/testify/assert"
)

func TestReadModelFromAppInstallation(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		appInstallation contentfulManagement.AppInstallation
		expectedModel   resource_app_installation.AppInstallationModel
	}{
		"null": {
			appInstallation: contentfulManagement.AppInstallation{},
			expectedModel:   resource_app_installation.AppInstallationModel{},
		},
		"empty": {
			appInstallation: contentfulManagement.AppInstallation{
				Parameters: contentfulManagement.NewOptAppInstallationParameters(contentfulManagement.AppInstallationParameters{}),
			},
			expectedModel: resource_app_installation.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedValue("{}"),
			},
		},
		"foo=bar": {
			appInstallation: contentfulManagement.AppInstallation{
				Parameters: contentfulManagement.NewOptAppInstallationParameters(contentfulManagement.AppInstallationParameters{
					"foo": []byte{'"', 'b', 'a', 'r', '"'},
				}),
			},
			expectedModel: resource_app_installation.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedValue("{\"foo\":\"bar\"}"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := resource_app_installation.AppInstallationModel{}

			provider.ReadAppInstallationModel(&model, test.appInstallation)

			assert.EqualValues(t, test.expectedModel, model)
		})
	}
}

func TestCreatePutAppInstallationRequestBody(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		model               resource_app_installation.AppInstallationModel
		expectErrors        bool
		expectWarnings      bool
		expectedRequestBody string
	}{
		"null": {
			model: resource_app_installation.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedNull(),
			},
			expectedRequestBody: "{}",
		},
		"unknown": {
			model: resource_app_installation.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedUnknown(),
			},
			expectWarnings:      true,
			expectedRequestBody: "{}",
		},
		"empty": {
			model: resource_app_installation.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedValue("{}"),
			},
			expectedRequestBody: "{\"parameters\":{}}",
		},
		"foo=bar": {
			model: resource_app_installation.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedValue("{\"foo\":\"bar\"}"),
			},
			expectedRequestBody: "{\"parameters\":{\"foo\":\"bar\"}}",
		},
		"invalid": {
			model: resource_app_installation.AppInstallationModel{
				Parameters: jsontypes.NewNormalizedValue("invalid"),
			},
			expectErrors:        true,
			expectedRequestBody: "{\"parameters\":{}}",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := &contentfulManagement.PutAppInstallationReq{}

			diags := provider.CreatePutAppInstallationRequestBody(req, test.model)

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
