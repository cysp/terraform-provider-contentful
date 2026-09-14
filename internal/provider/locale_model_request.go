package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *LocaleModel) ToLocaleData() (cm.LocaleData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	name, nameDiags := requestRequiredString(model.Name, path.Root("name"))
	diags.Append(nameDiags...)

	code, codeDiags := requestRequiredString(model.Code, path.Root("code"))
	diags.Append(codeDiags...)

	fallbackCode, fallbackCodeDiags := requestNullableString(model.FallbackCode, path.Root("fallback_code"))
	diags.Append(fallbackCodeDiags...)

	contentDeliveryAPI, contentDeliveryAPIDiags := requestRequiredBool(model.ContentDeliveryAPI, path.Root("content_delivery_api"))
	diags.Append(contentDeliveryAPIDiags...)

	contentManagementAPI, contentManagementAPIDiags := requestRequiredBool(model.ContentManagementAPI, path.Root("content_management_api"))
	diags.Append(contentManagementAPIDiags...)

	optional, optionalDiags := requestRequiredBool(model.Optional, path.Root("optional"))
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
