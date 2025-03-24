package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestAppInstallationModelReadFromResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		appInstallation cm.AppInstallation
		expectedModel   provider.AppInstallationResourceModel
	}{
		"null": {
			appInstallation: cm.AppInstallation{},
			expectedModel: provider.AppInstallationResourceModel{
				ID:              types.StringValue("//"),
				SpaceID:         types.StringValue(""),
				EnvironmentID:   types.StringValue(""),
				AppDefinitionID: types.StringValue(""),
			},
		},
		"empty": {
			appInstallation: cm.AppInstallation{
				Parameters: []byte("{}"),
			},
			expectedModel: provider.AppInstallationResourceModel{
				ID:              types.StringValue("//"),
				SpaceID:         types.StringValue(""),
				EnvironmentID:   types.StringValue(""),
				AppDefinitionID: types.StringValue(""),
				Parameters:      jsontypes.NewNormalizedValue("{}"),
			},
		},
		"foo=bar": {
			appInstallation: cm.AppInstallation{
				Parameters: []byte("{\"foo\":\"bar\"}"),
			},
			expectedModel: provider.AppInstallationResourceModel{
				ID:              types.StringValue("//"),
				SpaceID:         types.StringValue(""),
				EnvironmentID:   types.StringValue(""),
				AppDefinitionID: types.StringValue(""),
				Parameters:      jsontypes.NewNormalizedValue("{\"foo\":\"bar\"}"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := provider.AppInstallationResourceModel{}

			diags := model.ReadFromResponse(&test.appInstallation)

			assert.Equal(t, test.expectedModel, model)

			assert.Empty(t, diags)
		})
	}
}
