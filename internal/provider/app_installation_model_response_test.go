package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestAppInstallationModelReadFromResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		appInstallation cm.AppInstallation
		expectedModel   AppInstallationModel
	}{
		"null": {
			appInstallation: cm.AppInstallation{},
			expectedModel: AppInstallationModel{
				IDIdentityModel: NewIDIdentityModelFromMultipartID([]string{"", "", ""}),
				AppInstallationIdentityModel: AppInstallationIdentityModel{
					SpaceID:         types.StringValue(""),
					EnvironmentID:   types.StringValue(""),
					AppDefinitionID: types.StringValue(""),
				},
				Marketplace: types.SetNull(types.StringType),
			},
		},
		"empty": {
			appInstallation: cm.AppInstallation{
				Parameters: []byte("{}"),
			},
			expectedModel: AppInstallationModel{
				IDIdentityModel: NewIDIdentityModelFromMultipartID([]string{"", "", ""}),
				AppInstallationIdentityModel: AppInstallationIdentityModel{
					SpaceID:         types.StringValue(""),
					EnvironmentID:   types.StringValue(""),
					AppDefinitionID: types.StringValue(""),
				},
				Marketplace: types.SetNull(types.StringType),
				Parameters:  NewNormalizedJSONTypesNormalizedValue([]byte("{}")),
			},
		},
		"foo=bar": {
			appInstallation: cm.AppInstallation{
				Parameters: []byte("{\"foo\":\"bar\"}"),
			},
			expectedModel: AppInstallationModel{
				IDIdentityModel: NewIDIdentityModelFromMultipartID([]string{"", "", ""}),
				AppInstallationIdentityModel: AppInstallationIdentityModel{
					SpaceID:         types.StringValue(""),
					EnvironmentID:   types.StringValue(""),
					AppDefinitionID: types.StringValue(""),
				},
				Marketplace: types.SetNull(types.StringType),
				Parameters:  NewNormalizedJSONTypesNormalizedValue([]byte("{\"foo\":\"bar\"}")),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model, diags := NewAppInstallationResourceModelFromResponse(test.appInstallation, types.SetNull(types.StringType))

			assert.Equal(t, test.expectedModel, model)

			assert.Empty(t, diags)
		})
	}
}
