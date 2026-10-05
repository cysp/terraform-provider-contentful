package util_test

import (
	"errors"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestErrorDetailFromContentfulManagementResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		response any
		err      error
		expected string
	}{
		"Error": {
			response: cm.Error{
				Sys: cm.NewErrorSys("UnknownError"),
			},
			expected: "Error: UnknownError",
		},
		"Error pointer": {
			response: &cm.Error{
				Sys: cm.NewErrorSys("UnknownError"),
			},
			expected: "Error: UnknownError",
		},
		"nil Error pointer": {
			response: (*cm.Error)(nil),
			expected: "<nil>",
		},
		"Error: *ApplicationJSONError": {
			response: &cm.ApplicationJSONError{
				Type: cm.ErrorApplicationJSONError,
				Error: cm.Error{
					Sys:     cm.NewErrorSys("Unauthorized"),
					Message: cm.NewOptString("Unauthorized"),
				},
			},
			expected: "Error: Unauthorized: Unauthorized",
		},
		"ErrorStatusCode": {
			response: &cm.ErrorStatusCode{
				Response: cm.NewErrorApplicationJSONError(cm.Error{
					Sys: cm.NewErrorSys("UnknownError"),
				}),
			},
			expected: "Error: UnknownError",
		},
		"ErrorStatusCodeWithMessage": {
			response: &cm.ErrorStatusCode{
				Response: cm.NewErrorApplicationJSONError(cm.Error{
					Sys:     cm.NewErrorSys("UnknownError"),
					Message: cm.NewOptString("Error message"),
				}),
			},
			expected: "Error: UnknownError: Error message",
		},
		"VersionMismatch": {
			response: &cm.ErrorStatusCode{
				Response: cm.NewErrorApplicationJSONError(cm.Error{
					Sys:     cm.NewErrorSys(cm.ErrorSysIDVersionMismatch),
					Message: cm.NewOptString("Version mismatch"),
				}),
			},
			expected: "Error: VersionMismatch: Version mismatch\n\nContentful rejected the request because its version precondition was not satisfied.",
		},
		"ErrorStatusCodeWithMessageAndUnsupportedDetails": {
			response: &cm.ErrorStatusCode{
				Response: cm.NewErrorApplicationJSONError(cm.Error{
					Sys:     cm.NewErrorSys("UnknownError"),
					Message: cm.NewOptString("Error message"),
					Details: []byte(testJSON("Detailed reason for error")),
				}),
			},
			expected: "Error: UnknownError: Error message",
		},
		"ErrorStatusCodeWithMessageAndReasons": {
			response: &cm.ErrorStatusCode{
				Response: cm.NewErrorApplicationJSONError(cm.Error{
					Sys:     cm.NewErrorSys("UnknownError"),
					Message: cm.NewOptString("Error message"),
					Details: []byte(testJSON(map[string]any{"reasons": "Detailed reason for error"})),
				}),
			},
			expected: "Error: UnknownError: Error message: Detailed reason for error",
		},
		"ErrorStatusCodeWithMessageAndUnsupportedReason": {
			response: &cm.ErrorStatusCode{
				Response: cm.NewErrorApplicationJSONError(cm.Error{
					Sys:     cm.NewErrorSys("UnknownError"),
					Message: cm.NewOptString("Error message"),
					Details: []byte(testJSON(map[string]any{"reasons": []any{"Reason 1", "Reason 2"}})),
				}),
			},
			expected: "Error: UnknownError: Error message",
		},
		"string": {
			response: "string",
			expected: "string",
		},
		"error": {
			err:      errors.ErrUnsupported,
			expected: "unsupported operation",
		},
		"ValidationFailed with string errors": {
			response: &cm.ErrorStatusCode{
				StatusCode: 422,
				Response: cm.NewErrorApplicationJSONError(cm.Error{
					Sys:     cm.NewErrorSys("ValidationFailed"),
					Message: cm.NewOptString("Validation error"),
					Details: []byte(testJSON(map[string]any{
						"errors": "AppAction cannot have both parametersSchema and parameters. Please provide just a parametersSchema.",
					})),
				}),
			},
			expected: "Error: ValidationFailed: Validation error\n  AppAction cannot have both parametersSchema and parameters. Please provide just a parametersSchema.",
		},
		"ValidationFailed with detailed errors": {
			response: &cm.ErrorStatusCode{
				StatusCode: 422,
				Response: cm.NewErrorApplicationJSONError(cm.Error{
					Sys:     cm.NewErrorSys("ValidationFailed"),
					Message: cm.NewOptString("Validation error"),
					Details: []byte(testJSON(map[string]any{
						"errors": []any{
							map[string]any{
								"name":    "required",
								"details": "The property \"annotations\" is required here",
								"path":    []any{"metadata", "annotations"},
							},
							map[string]any{
								"name":    "required",
								"details": "The property \"taxonomy\" is required here",
								"path":    []any{"metadata", "taxonomy"},
							},
							map[string]any{
								"name":     "in",
								"details":  "Value must be one of expected values",
								"path":     []any{"metadata"},
								"value":    map[string]any{},
								"expected": []any{map[string]any{"required": []any{"annotations"}}, map[string]any{"required": []any{"taxonomy"}}},
							},
						},
					})),
				}),
			},
			expected: "Error: ValidationFailed: Validation error\n  metadata.annotations: The property \"annotations\" is required here\n  metadata.taxonomy: The property \"taxonomy\" is required here\n  metadata: Value must be one of expected values",
		},
		"ValidationFailed with detailed errors in fields list item": {
			response: &cm.ErrorStatusCode{
				StatusCode: 422,
				Response: cm.NewErrorApplicationJSONError(cm.Error{
					Sys:     cm.NewErrorSys("ValidationFailed"),
					Message: cm.NewOptString("Validation error"),
					Details: []byte(testJSON(map[string]any{
						"errors": []any{
							map[string]any{
								"name":    "type",
								"details": "The type of \"required\" is incorrect, expected type: Boolean",
								"path":    []any{"fields", 0, "required"},
								"type":    "Boolean",
								"value":   "true",
							},
						},
					})),
				}),
			},
			expected: "Error: ValidationFailed: Validation error\n  fields[0].required: The type of \"required\" is incorrect, expected type: Boolean",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := util.ErrorDetailFromContentfulManagementResponse(test.response, test.err)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestContentfulManagementValidationFailedErrorDetails(t *testing.T) {
	t.Parallel()

	for name, details := range map[string]string{
		"missing errors":    testJSON(map[string]any{}),
		"null errors":       testJSON(map[string]any{"errors": nil}),
		"unsupported error": testJSON(map[string]any{"errors": 42}),
		"invalid JSON":      `{`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			messages, ok := util.ContentfulManagementValidationFailedErrorDetails([]byte(details))

			assert.False(t, ok)
			assert.Empty(t, messages)
		})
	}
}

func TestOptStringToStringValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    cm.OptString
		expected types.String
	}{
		"set": {
			input:    cm.NewOptString("string"),
			expected: types.StringValue("string"),
		},
		"set: empty": {
			input:    cm.NewOptString(""),
			expected: types.StringValue(""),
		},
		"unset": {
			input:    cm.OptString{},
			expected: types.StringNull(),
		},
		"unset: non-empty": {
			input:    cm.OptString{Value: "string"},
			expected: types.StringNull(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := util.OptStringToStringValue(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}
