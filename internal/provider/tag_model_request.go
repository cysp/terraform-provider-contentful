package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (model *TagModel) ToGetTagParams() cm.GetTagParams {
	return cm.GetTagParams{
		SpaceID:       model.SpaceID.ValueString(),
		EnvironmentID: model.EnvironmentID.ValueString(),
		TagID:         model.TagID.ValueString(),
	}
}

func (model *TagModel) ToPutTagParams() cm.PutTagParams {
	params := cm.PutTagParams{
		SpaceID:       model.SpaceID.ValueString(),
		EnvironmentID: model.EnvironmentID.ValueString(),
		TagID:         model.TagID.ValueString(),
	}

	if !model.Visibility.IsUnknown() && !model.Visibility.IsNull() {
		params.XContentfulTagVisibility.SetTo(model.Visibility.ValueString())
	}

	return params
}

func (model *TagModel) ToDeleteTagParams() cm.DeleteTagParams {
	return cm.DeleteTagParams{
		SpaceID:       model.SpaceID.ValueString(),
		EnvironmentID: model.EnvironmentID.ValueString(),
		TagID:         model.TagID.ValueString(),
	}
}

func (model *TagModel) ToTagRequest() cm.TagRequest {
	request := cm.TagRequest{
		Sys: cm.TagRequestSys{
			Type: cm.TagRequestSysTypeTag,
		},
		Name: model.Name.ValueString(),
	}

	if !model.TagID.IsUnknown() && !model.TagID.IsNull() {
		reqID := model.TagID.ValueString()
		request.Sys.ID = cm.NewOptString(reqID)
	}

	if !model.Visibility.IsUnknown() && !model.Visibility.IsNull() {
		visibility := model.Visibility.ValueString()
		request.Sys.Visibility = cm.NewOptString(visibility)
	}

	return request
}
