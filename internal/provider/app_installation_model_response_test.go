package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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
				IDIdentityModel: NewIDIdentityModelFromMultipartID("", "", ""),
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
				Parameters: []byte(testJSON(map[string]any{})),
			},
			expectedModel: AppInstallationModel{
				IDIdentityModel: NewIDIdentityModelFromMultipartID("", "", ""),
				AppInstallationIdentityModel: AppInstallationIdentityModel{
					SpaceID:         types.StringValue(""),
					EnvironmentID:   types.StringValue(""),
					AppDefinitionID: types.StringValue(""),
				},
				Marketplace: types.SetNull(types.StringType),
				Parameters:  NewNormalizedJSONValue([]byte(testJSON(map[string]any{}))),
			},
		},
		"foo=bar": {
			appInstallation: cm.AppInstallation{
				Parameters: []byte(testJSON(map[string]any{"foo": "bar"})),
			},
			expectedModel: AppInstallationModel{
				IDIdentityModel: NewIDIdentityModelFromMultipartID("", "", ""),
				AppInstallationIdentityModel: AppInstallationIdentityModel{
					SpaceID:         types.StringValue(""),
					EnvironmentID:   types.StringValue(""),
					AppDefinitionID: types.StringValue(""),
				},
				Marketplace: types.SetNull(types.StringType),
				Parameters:  NewNormalizedJSONValue([]byte(testJSON(map[string]any{"foo": "bar"}))),
			},
		},
		"large integer": {
			appInstallation: cm.AppInstallation{
				Parameters: []byte(testJSON(map[string]any{"externalId": int64(9007199254740993)})),
			},
			expectedModel: AppInstallationModel{
				IDIdentityModel: NewIDIdentityModelFromMultipartID("", "", ""),
				AppInstallationIdentityModel: AppInstallationIdentityModel{
					SpaceID:         types.StringValue(""),
					EnvironmentID:   types.StringValue(""),
					AppDefinitionID: types.StringValue(""),
				},
				Marketplace: types.SetNull(types.StringType),
				Parameters:  jsontypes.NewNormalizedValue(testJSON(map[string]any{"externalId": int64(9007199254740993)})),
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
