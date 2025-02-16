package util

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

func OptBoolToBoolValue(b cm.OptBool) basetypes.BoolValue {
	return types.BoolPointerValue(b.ValueBoolPointer())
}

func BoolValueToOptBool(b basetypes.BoolValue) cm.OptBool {
	return cm.NewOptPointerBool(b.ValueBoolPointer())
}

func OptStringToStringValue(s cm.OptString) basetypes.StringValue {
	return types.StringPointerValue(s.ValueStringPointer())
}

func OptNilStringToStringValue(s cm.OptNilString) basetypes.StringValue {
	return types.StringPointerValue(s.ValueStringPointer())
}

func StringValueToOptString(s basetypes.StringValue) cm.OptString {
	return cm.NewOptPointerString(s.ValueStringPointer())
}

func StringValueToOptNilString(value basetypes.StringValue) cm.OptNilString {
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

func NewStringListValueFromStringSlice(ctx context.Context, slice []string) (basetypes.ListValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	list, listDiags := types.ListValueFrom(ctx, types.String{}.Type(ctx), slice)
	diags.Append(listDiags...)

	return list, diags
}
