package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *TagModel) ToGetTagParams() cm.GetTagParams {
	return cm.GetTagParams{
		SpaceID:       model.SpaceID.ValueString(),
		EnvironmentID: model.EnvironmentID.ValueString(),
		TagID:         model.TagID.ValueString(),
	}
}

func (model *TagModel) ToPutTagRequest(modelPath path.Path) (cm.PutTagParams, cm.TagRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID, spaceIDDiags := requestRequiredString(model.SpaceID, modelPath.AtName("space_id"))
	diags.Append(spaceIDDiags...)

	environmentID, environmentIDDiags := requestRequiredString(model.EnvironmentID, modelPath.AtName("environment_id"))
	diags.Append(environmentIDDiags...)

	tagID, tagIDDiags := requestRequiredString(model.TagID, modelPath.AtName("tag_id"))
	diags.Append(tagIDDiags...)

	name, nameDiags := requestRequiredString(model.Name, modelPath.AtName("name"))
	diags.Append(nameDiags...)

	visibility, visibilityDiags := requestRequiredString(model.Visibility, modelPath.AtName("visibility"))
	diags.Append(visibilityDiags...)

	if diags.HasError() {
		return cm.PutTagParams{}, cm.TagRequest{}, diags
	}

	params := cm.PutTagParams{
		SpaceID:       spaceID,
		EnvironmentID: environmentID,
		TagID:         tagID,
	}
	params.XContentfulTagVisibility.SetTo(visibility)

	request := cm.TagRequest{
		Sys: cm.TagRequestSys{
			Type:       cm.TagRequestSysTypeTag,
			ID:         cm.NewOptString(tagID),
			Visibility: cm.NewOptString(visibility),
		},
		Name: name,
	}

	return params, request, diags
}

func (model *TagModel) ToDeleteTagParams() cm.DeleteTagParams {
	return cm.DeleteTagParams{
		SpaceID:       model.SpaceID.ValueString(),
		EnvironmentID: model.EnvironmentID.ValueString(),
		TagID:         model.TagID.ValueString(),
	}
}
