package util

import (
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ErrorDetailFromContentfulManagementResponse(response interface{}, err error) string {
	if response, ok := response.(*cm.ErrorStatusCode); ok {
		responseType, err := response.Response.Sys.Type.MarshalText()

		if err == nil {
			detail := string(responseType) + ": " + response.Response.Sys.ID

			if responseMessage, ok := response.Response.Message.Get(); ok {
				detail += ": " + responseMessage
			}

			return detail
		}
	}

	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%v", response)
}

func OptBoolToBoolValue(b cm.OptBool) types.Bool {
	return types.BoolPointerValue(b.ValueBoolPointer())
}

func BoolValueToOptBool(b types.Bool) cm.OptBool {
	return cm.NewOptPointerBool(b.ValueBoolPointer())
}

func OptStringToStringValue(s cm.OptString) types.String {
	return types.StringPointerValue(s.ValueStringPointer())
}

func OptNilStringToStringValue(s cm.OptNilString) types.String {
	return types.StringPointerValue(s.ValueStringPointer())
}

func StringValueToOptString(s types.String) cm.OptString {
	return cm.NewOptPointerString(s.ValueStringPointer())
}

func StringValueToOptNilString(value types.String) cm.OptNilString {
	ons := cm.OptNilString{}

	if !value.IsUnknown() {
		if value.IsNull() {
			ons.SetToNull()
		} else {
			ons.SetTo(value.ValueString())
		}
	}

	return ons
}
