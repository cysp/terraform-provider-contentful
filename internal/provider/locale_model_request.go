package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *LocaleModel) ToCreateLocaleParams(modelPath path.Path) (cm.CreateLocaleParams, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID, spaceIDDiags := requestRequiredString(model.SpaceID, modelPath.AtName("space_id"))
	diags.Append(spaceIDDiags...)

	environmentID, environmentIDDiags := requestRequiredString(model.EnvironmentID, modelPath.AtName("environment_id"))
	diags.Append(environmentIDDiags...)

	if diags.HasError() {
		return cm.CreateLocaleParams{}, diags
	}

	return cm.CreateLocaleParams{
		SpaceID:       spaceID,
		EnvironmentID: environmentID,
	}, diags
}

func (model *LocaleModel) ToGetLocaleParams() cm.GetLocaleParams {
	return cm.GetLocaleParams{
		SpaceID:       model.SpaceID.ValueString(),
		EnvironmentID: model.EnvironmentID.ValueString(),
		LocaleID:      model.LocaleID.ValueString(),
	}
}

func (model *LocaleModel) ToPutLocaleParams(modelPath path.Path) (cm.PutLocaleParams, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID, spaceIDDiags := requestRequiredString(model.SpaceID, modelPath.AtName("space_id"))
	diags.Append(spaceIDDiags...)

	environmentID, environmentIDDiags := requestRequiredString(model.EnvironmentID, modelPath.AtName("environment_id"))
	diags.Append(environmentIDDiags...)

	localeID, localeIDDiags := requestRequiredString(model.LocaleID, modelPath.AtName("locale_id"))
	diags.Append(localeIDDiags...)

	if diags.HasError() {
		return cm.PutLocaleParams{}, diags
	}

	return cm.PutLocaleParams{
		SpaceID:       spaceID,
		EnvironmentID: environmentID,
		LocaleID:      localeID,
	}, diags
}

func (model *LocaleModel) ToDeleteLocaleParams() cm.DeleteLocaleParams {
	return cm.DeleteLocaleParams{
		SpaceID:       model.SpaceID.ValueString(),
		EnvironmentID: model.EnvironmentID.ValueString(),
		LocaleID:      model.LocaleID.ValueString(),
	}
}

func (model *LocaleModel) ToLocaleData(modelPath path.Path) (cm.LocaleData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	name, nameDiags := requestRequiredString(model.Name, modelPath.AtName("name"))
	diags.Append(nameDiags...)

	code, codeDiags := requestRequiredString(model.Code, modelPath.AtName("code"))
	diags.Append(codeDiags...)

	fallbackCode, fallbackCodeDiags := requestNullableString(model.FallbackCode, modelPath.AtName("fallback_code"))
	diags.Append(fallbackCodeDiags...)

	contentDeliveryAPI, contentDeliveryAPIDiags := requestRequiredBool(model.ContentDeliveryAPI, modelPath.AtName("content_delivery_api"))
	diags.Append(contentDeliveryAPIDiags...)

	contentManagementAPI, contentManagementAPIDiags := requestRequiredBool(model.ContentManagementAPI, modelPath.AtName("content_management_api"))
	diags.Append(contentManagementAPIDiags...)

	optional, optionalDiags := requestRequiredBool(model.Optional, modelPath.AtName("optional"))
	diags.Append(optionalDiags...)

	if diags.HasError() {
		return cm.LocaleData{}, diags
	}

	return cm.LocaleData{
		Name:                 name,
		Code:                 code,
		FallbackCode:         cm.NewNilPointerString(fallbackCode.ValueStringPointer()),
		ContentDeliveryApi:   contentDeliveryAPI,
		ContentManagementApi: contentManagementAPI,
		Optional:             optional,
	}, diags
}
