package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTagResourceModelFromResponse(tag cm.Tag) TagModel {
	spaceID := tag.Sys.Space.Sys.ID
	environmentID := tag.Sys.Environment.Sys.ID
	tagID := tag.Sys.ID

	model := TagModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID, tagID),
		TagIdentityModel: TagIdentityModel{
			SpaceID:       types.StringValue(spaceID),
			EnvironmentID: types.StringValue(environmentID),
			TagID:         types.StringValue(tagID),
		},
		Name:       types.StringValue(tag.Name),
		Visibility: types.StringValue(tag.Sys.Visibility),
	}

	return model
}
