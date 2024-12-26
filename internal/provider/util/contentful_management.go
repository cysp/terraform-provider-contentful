package util

import (
	"context"
	"fmt"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func ErrorDetailFromContentfulManagementResponse(response interface{}, err error) string {
	if response, ok := response.(*contentfulManagement.ErrorStatusCode); ok {
		if detail := ErrorDetailFromContentfulManagementErrorStatusCode(response); detail != "" {
			return detail
		}
	}

	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%v", response)
}

func ErrorDetailFromContentfulManagementErrorStatusCode(response *contentfulManagement.ErrorStatusCode) string {
	if response == nil {
		return ""
	}

	responseType, err := response.Response.Sys.Type.MarshalText()
	if err != nil {
		return ""
	}

	detail := string(responseType) + ": " + response.Response.Sys.ID

	if responseMessage, ok := response.Response.Message.Get(); ok {
		detail += ": " + responseMessage
	}

	if responseReasons := ErrorReasonsFromContentfulManagementErrorStatusCodeDetails(response.Response.Details); responseReasons != "" {
		detail += ": " + responseReasons
	}

	return detail
}

func ErrorReasonsFromContentfulManagementErrorStatusCodeDetails(details jx.Raw) string {
	if details == nil {
		return ""
	}

	responseDetailsDecoder := jx.DecodeBytes(details)

	detailsObject := contentfulManagement.ErrorWithDetailsWithReasonsString{}

	if err := detailsObject.Decode(responseDetailsDecoder); err == nil {
		if reasons, ok := detailsObject.Reasons.Get(); ok {
			return reasons
		}
	}

	return ""
}

func OptBoolToBoolValue(b contentfulManagement.OptBool) basetypes.BoolValue {
	return types.BoolPointerValue(b.ValueBoolPointer())
}

func BoolValueToOptBool(b basetypes.BoolValue) contentfulManagement.OptBool {
	return contentfulManagement.NewOptPointerBool(b.ValueBoolPointer())
}

func OptStringToStringValue(s contentfulManagement.OptString) basetypes.StringValue {
	return types.StringPointerValue(s.ValueStringPointer())
}

func OptNilStringToStringValue(s contentfulManagement.OptNilString) basetypes.StringValue {
	return types.StringPointerValue(s.ValueStringPointer())
}

func StringValueToOptString(s basetypes.StringValue) contentfulManagement.OptString {
	return contentfulManagement.NewOptPointerString(s.ValueStringPointer())
}

func StringValueToOptNilString(value basetypes.StringValue) contentfulManagement.OptNilString {
	ons := contentfulManagement.OptNilString{}

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
