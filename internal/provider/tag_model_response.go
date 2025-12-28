package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTagResourceModelFromResponse(_ context.Context, tag cm.Tag) (TagModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

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

	return model, diags
}
