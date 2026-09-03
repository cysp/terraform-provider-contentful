package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEnvironmentResourceModelFromResponse(environment cm.Environment) EnvironmentModel {
	spaceID := environment.Sys.Space.Sys.ID
	environmentID := environment.Sys.ID

	model := EnvironmentModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID),
		EnvironmentIdentityModel: EnvironmentIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
		},
		Status: types.StringValue(environment.Sys.Status.Sys.ID),
	}

	model.Name = types.StringValue(environment.Name)

	return model
}
