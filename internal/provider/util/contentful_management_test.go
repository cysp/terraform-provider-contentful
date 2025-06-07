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
		response interface{}
		err      error
		expected string
	}{
		"Error": {
			response: cm.Error{
				Sys: cm.ErrorSys{
					Type: cm.ErrorSysTypeError,
					ID:   "UnknownError",
				},
			},
			expected: "Error: UnknownError",
		},
		"Error: *GetAuthenticatedUserApplicationVndContentfulManagementV1JSONUnauthorized": {
			response: &cm.GetAuthenticatedUserApplicationVndContentfulManagementV1JSONUnauthorized{
				Sys: cm.ErrorSys{
					Type: cm.ErrorSysTypeError,
					ID:   "Unauthorized",
				},
				Message: cm.NewOptString("Unauthorized"),
			},
			expected: "Error: Unauthorized: Unauthorized",
		},
		"ErrorStatusCode": {
			response: &cm.ErrorStatusCode{
				Response: cm.Error{
					Sys: cm.ErrorSys{
						Type: cm.ErrorSysTypeError,
						ID:   "UnknownError",
					},
				},
			},
			expected: "Error: UnknownError",
		},
		"ErrorStatusCodeWithMessage": {
			response: &cm.ErrorStatusCode{
				Response: cm.Error{
					Sys: cm.ErrorSys{
						Type: cm.ErrorSysTypeError,
						ID:   "UnknownError",
					},
					Message: cm.NewOptString("Error message"),
				},
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
		"ValidationFailed with detailed errors": {
			response: &cm.ErrorStatusCode{
				StatusCode: 422,
				Response: cm.Error{
					Sys: cm.ErrorSys{
						Type: cm.ErrorSysTypeError,
						ID:   "ValidationFailed",
					},
					Message: cm.NewOptString("Validation error"),
					Details: []byte("{\"errors\":[{\"name\":\"required\",\"details\":\"The property \\\"annotations\\\" is required here\",\"path\":[\"metadata\",\"annotations\"]},{\"name\":\"required\",\"details\":\"The property \\\"taxonomy\\\" is required here\",\"path\":[\"metadata\",\"taxonomy\"]},{\"name\":\"in\",\"details\":\"Value must be one of expected values\",\"path\":[\"metadata\"],\"value\": {},\"expected\":[{\"required\":[\"annotations\"]},{\"required\":[\"taxonomy\"]}]}]}"),
				},
			},
			expected: "Error: ValidationFailed: Validation error\n  metadata.annotations: The property \"annotations\" is required here\n  metadata.taxonomy: The property \"taxonomy\" is required here\n  metadata: Value must be one of expected values",
		},
		"ValidationFailed with detailed errors in fields list item": {
			response: &cm.ErrorStatusCode{
				StatusCode: 422,
				Response: cm.Error{
					Sys: cm.ErrorSys{
						Type: cm.ErrorSysTypeError,
						ID:   "ValidationFailed",
					},
					Message: cm.NewOptString("Validation error"),
					Details: []byte("{\"errors\":[{\"name\":\"type\",\"details\":\"The type of \\\"required\\\" is incorrect, expected type: Boolean\",\"path\":[\"fields\",0,\"required\"],\"type\":\"Boolean\",\"value\":\"true\"}]}"),
				},
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

func TestStringValueToOptString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    types.String
		expected cm.OptString
	}{
		"set": {
			input:    types.StringValue("string"),
			expected: cm.NewOptString("string"),
		},
		"set: empty": {
			input:    types.StringValue(""),
			expected: cm.NewOptString(""),
		},
		"null": {
			input:    types.StringNull(),
			expected: cm.OptString{},
		},
		"unknown": {
			input:    types.StringUnknown(),
			expected: cm.NewOptString(""),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := util.StringValueToOptString(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStringValueToOptNilString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    types.String
		expected cm.OptNilString
	}{
		"set": {
			input:    types.StringValue("string"),
			expected: cm.NewOptNilString("string"),
		},
		"set: empty": {
			input:    types.StringValue(""),
			expected: cm.NewOptNilString(""),
		},
		"null": {
			input:    types.StringNull(),
			expected: cm.NewOptNilStringNull(),
		},
		"unknown": {
			input:    types.StringUnknown(),
			expected: cm.OptNilString{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := util.StringValueToOptNilString(test.input)

			assert.Equal(t, test.expected, actual)
		})
	}
}
