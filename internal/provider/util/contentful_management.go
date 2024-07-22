package util

import (
	"fmt"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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

func OptStringToStringValue(s contentfulManagement.OptString) basetypes.StringValue {
	if value, ok := s.Get(); ok {
		return types.StringValue(value)
	}

	return types.StringNull()
}

func StringValueToOptString(s basetypes.StringValue) contentfulManagement.OptString {
	switch value := s.ValueStringPointer(); {
	case value == nil:
		return contentfulManagement.OptString{}
	default:
		return contentfulManagement.NewOptString(*value)
	}
}
