package provider

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

func (model *LocaleModel) ToCreateLocaleParams() cm.CreateLocaleParams {
	return cm.CreateLocaleParams{
		SpaceID:       model.SpaceID.ValueString(),
		EnvironmentID: model.EnvironmentID.ValueString(),
	}
}

func (model *LocaleModel) ToGetLocaleParams() cm.GetLocaleParams {
	return cm.GetLocaleParams{
		SpaceID:       model.SpaceID.ValueString(),
		EnvironmentID: model.EnvironmentID.ValueString(),
		LocaleID:      model.LocaleID.ValueString(),
	}
}

func (model *LocaleModel) ToPutLocaleParams(version int) cm.PutLocaleParams {
	return cm.PutLocaleParams{
		SpaceID:            model.SpaceID.ValueString(),
		EnvironmentID:      model.EnvironmentID.ValueString(),
		LocaleID:           model.LocaleID.ValueString(),
		XContentfulVersion: version,
	}
}

func (model *LocaleModel) ToDeleteLocaleParams() cm.DeleteLocaleParams {
	return cm.DeleteLocaleParams{
		SpaceID:       model.SpaceID.ValueString(),
		EnvironmentID: model.EnvironmentID.ValueString(),
		LocaleID:      model.LocaleID.ValueString(),
	}
}

func (model *LocaleModel) ToLocaleData() cm.LocaleData {
	request := cm.LocaleData{
		Name:                 model.Name.ValueString(),
		Code:                 model.Code.ValueString(),
		FallbackCode:         cm.NewNilStringNull(),
		ContentDeliveryApi:   model.ContentDeliveryAPI.ValueBool(),
		ContentManagementApi: model.ContentManagementAPI.ValueBool(),
		Optional:             model.Optional.ValueBool(),
	}

	if !model.FallbackCode.IsNull() && !model.FallbackCode.IsUnknown() {
		request.FallbackCode = cm.NewNilString(model.FallbackCode.ValueString())
	}

	return request
}
