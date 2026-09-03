package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewEnvironmentAliasResourceModelFromResponse(environmentAlias cm.EnvironmentAlias) EnvironmentAliasModel {
	spaceID := environmentAlias.Sys.Space.Sys.ID
	environmentAliasID := environmentAlias.Sys.ID
	targetEnvironmentID := environmentAlias.Environment.Sys.ID

	model := EnvironmentAliasModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentAliasID),
		EnvironmentAliasIdentityModel: EnvironmentAliasIdentityModel{
			SpaceID:            types.StringValue(spaceID),
			EnvironmentAliasID: types.StringValue(environmentAliasID),
		},
	}

	model.TargetEnvironmentID = types.StringValue(targetEnvironmentID)

	return model
}
