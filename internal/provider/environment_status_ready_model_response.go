package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEnvironmentStatusReadyModelFromResponse(environment cm.Environment) EnvironmentStatusReadyModel {
	spaceID := environment.Sys.Space.Sys.ID
	environmentID := environment.Sys.ID

	model := EnvironmentStatusReadyModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID),
		EnvironmentIdentityModel: EnvironmentIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
		},
		Status: types.StringValue(environment.Sys.Status.Sys.ID),
	}

	return model
}
