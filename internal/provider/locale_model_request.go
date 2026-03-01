package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
)

func (model *LocaleModel) ToGetLocaleParams() cm.GetLocaleParams {
	return cm.GetLocaleParams{
		SpaceID:  model.SpaceID.ValueString(),
		LocaleID: model.LocaleID.ValueString(),
	}
}

func (model *LocaleModel) ToPutLocaleParams() cm.PutLocaleParams {
	return cm.PutLocaleParams{
		SpaceID:  model.SpaceID.ValueString(),
		LocaleID: model.LocaleID.ValueString(),
	}
}

func (model *LocaleModel) ToDeleteLocaleParams() cm.DeleteLocaleParams {
	return cm.DeleteLocaleParams{
		SpaceID:  model.SpaceID.ValueString(),
		LocaleID: model.LocaleID.ValueString(),
	}
}

func (model *LocaleModel) ToLocaleRequest() cm.LocaleRequest {
	request := cm.LocaleRequest{
		Name: model.Name.ValueString(),
		Code: model.Code.ValueString(),
	}

	if !model.FallbackCode.IsUnknown() {
		request.FallbackCode = util.StringValueToOptNilString(model.FallbackCode)
	}

	if !model.Optional.IsUnknown() && !model.Optional.IsNull() {
		request.Optional = util.BoolValueToOptBool(model.Optional)
	}

	if !model.ContentDeliveryAPI.IsUnknown() && !model.ContentDeliveryAPI.IsNull() {
		request.ContentDeliveryAPI = util.BoolValueToOptBool(model.ContentDeliveryAPI)
	}

	if !model.ContentManagementAPI.IsUnknown() && !model.ContentManagementAPI.IsNull() {
		request.ContentManagementAPI = util.BoolValueToOptBool(model.ContentManagementAPI)
	}

	return request
}
