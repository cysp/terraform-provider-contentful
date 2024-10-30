package util

import (
	"context"
	"fmt"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ErrorDetailFromContentfulManagementResponse generates a detailed error message from a Contentful Management response and error.
func ErrorDetailFromContentfulManagementResponse(response interface{}, err error) string {
	if response, ok := response.(*contentfulManagement.ErrorStatusCode); ok {
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

// OptBoolToBoolValue converts a Contentful Management OptBool to a Terraform Plugin Framework BoolValue.
func OptBoolToBoolValue(b contentfulManagement.OptBool) basetypes.BoolValue {
	return types.BoolPointerValue(b.ValueBoolPointer())
}

// BoolValueToOptBool converts a Terraform Plugin Framework BoolValue to a Contentful Management OptBool.
func BoolValueToOptBool(b basetypes.BoolValue) contentfulManagement.OptBool {
	return contentfulManagement.NewOptPointerBool(b.ValueBoolPointer())
}

// OptStringToStringValue converts a Contentful Management OptString to a Terraform Plugin Framework StringValue.
func OptStringToStringValue(s contentfulManagement.OptString) basetypes.StringValue {
	return types.StringPointerValue(s.ValueStringPointer())
}

// OptNilStringToStringValue converts a Contentful Management OptNilString to a Terraform Plugin Framework StringValue.
func OptNilStringToStringValue(s contentfulManagement.OptNilString) basetypes.StringValue {
	return types.StringPointerValue(s.ValueStringPointer())
}

// StringValueToOptString converts a Terraform Plugin Framework StringValue to a Contentful Management OptString.
func StringValueToOptString(s basetypes.StringValue) contentfulManagement.OptString {
	return contentfulManagement.NewOptPointerString(s.ValueStringPointer())
}

// StringValueToOptNilString converts a Terraform Plugin Framework StringValue to a Contentful Management OptNilString.
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

// NewStringListValueFromStringSlice creates a new Terraform Plugin Framework ListValue from a slice of strings.
func NewStringListValueFromStringSlice(ctx context.Context, slice []string) (basetypes.ListValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	list, listDiags := types.ListValueFrom(ctx, types.String{}.Type(ctx), slice)
	diags.Append(listDiags...)

	return list, diags
}
