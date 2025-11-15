package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ErrorDetailFromContentfulManagementResponse(response any, err error) string {
	if response, ok := response.(cm.ErrorStatusCodeResponse); ok {
		if detail := ErrorDetailFromContentfulManagementErrorStatusCode(response); detail != "" {
			return detail
		}
	}

	if response, ok := response.(cm.ErrorResponse); ok {
		responseError, responseErrorOk := response.GetError()
		if responseErrorOk {
			if detail := ErrorDetailFromContentfulManagementError(responseError); detail != "" {
				return detail
			}
		}
	}

	if response != nil {
		cmErrorType := reflect.TypeFor[cm.Error]()

		rValue := reflect.ValueOf(response)
		if rValue.Kind() == reflect.Ptr && !rValue.IsNil() {
			rValue = rValue.Elem()
		}

		if rValue.CanConvert(cmErrorType) {
			cmError, ok := rValue.Convert(cmErrorType).Interface().(cm.Error)
			if ok {
				return ErrorDetailFromContentfulManagementError(cmError)
			}
		}
	}

	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%v", response)
}

func ErrorDetailFromContentfulManagementErrorStatusCode(response cm.ErrorStatusCodeResponse) string {
	if response == nil {
		return ""
	}

	responseError, responseErrorOk := response.GetError()
	if !responseErrorOk {
		return ""
	}

	return ErrorDetailFromContentfulManagementError(responseError)
}

func ErrorDetailFromContentfulManagementError(response cm.Error) string {
	responseType, err := response.Sys.Type.MarshalText()
	if err != nil {
		return ""
	}

	var detailStringBuilder strings.Builder

	detailStringBuilder.WriteString(string(responseType) + ": " + response.Sys.ID)

	if responseMessage, ok := response.Message.Get(); ok {
		detailStringBuilder.WriteString(": " + responseMessage)
	}

	if reasons := ContentfulManagementErrorDetailReasons(response.Details); reasons != "" {
		detailStringBuilder.WriteString(": " + reasons)
	}

	if response.Sys.ID == "ValidationFailed" {
		if details, ok := ContentfulManagementValidationFailedErrorDetails(response.Details); ok {
			for _, s := range details {
				detailStringBuilder.WriteString("\n  " + s)
			}
		}
	}

	return detailStringBuilder.String()
}

func ContentfulManagementErrorDetailReasons(detailsJSONBytes []byte) string {
	type ContentfulManagementErrorDetails struct {
		Reasons *string `json:"reasons"`
	}

	details := ContentfulManagementErrorDetails{}
	jsonUnmarshalErr := json.Unmarshal(detailsJSONBytes, &details)

	if jsonUnmarshalErr == nil && details.Reasons != nil {
		return *details.Reasons
	}

	return ""
}

type ValidationFailedErrorDetails struct {
	Errors []ValidationFailedErrorDetailsError `json:"errors"`
}

type ValidationFailedErrorDetailsError struct {
	Name    string `json:"name"`
	Details string `json:"details"`
	Path    []any  `json:"path"`
}

func ContentfulManagementValidationFailedErrorDetails(detailsJSONBytes []byte) ([]string, bool) {
	details := ValidationFailedErrorDetails{}

	err := json.Unmarshal(detailsJSONBytes, &details)
	if err != nil {
		return []string{}, false
	}

	strings := make([]string, 0, len(details.Errors))

	for _, err := range details.Errors {
		pathString, pathStringOk := ContentfulManagementValidationFailedErrorDetailsErrorPathToString(err.Path)

		detailString := ""

		if pathStringOk && pathString != "" {
			detailString += pathString
		}

		if err.Details != "" {
			if detailString != "" {
				detailString += ": "
			}

			detailString += err.Details
		}

		strings = append(strings, detailString)
	}

	return strings, true
}

func ContentfulManagementValidationFailedErrorDetailsErrorPathToString(path []any) (string, bool) {
	if len(path) == 0 {
		return "", false
	}

	pathString := ""

	for _, pathComponent := range path {
		switch pathComponent := pathComponent.(type) {
		case string:
			if pathString != "" {
				pathString += "."
			}

			pathString += pathComponent
		case int64, int32, int16, int8, int, uint64, uint32, uint16, uint8, uint, float64, float32:
			pathString += fmt.Sprintf("[%v]", pathComponent)
		default:
			if pathString != "" {
				pathString += "."
			}

			pathString += "<unknown>"
		}
	}

	return pathString, true
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
