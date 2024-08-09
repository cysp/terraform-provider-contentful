package util_test

import (
	"errors"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
)

func TestErrorDetailFromContentfulManagementResponse(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		response interface{}
		err      error
		expected string
	}{
		"ErrorStatusCode": {
			response: &contentfulManagement.ErrorStatusCode{
				Response: contentfulManagement.Error{
					Sys: contentfulManagement.ErrorSys{
						Type: contentfulManagement.ErrorSysTypeError,
						ID:   "UnknownError",
					},
				},
			},
			expected: "Error: UnknownError",
		},
		"ErrorStatusCodeWithMessage": {
			response: &contentfulManagement.ErrorStatusCode{
				Response: contentfulManagement.Error{
					Sys: contentfulManagement.ErrorSys{
						Type: contentfulManagement.ErrorSysTypeError,
						ID:   "UnknownError",
					},
					Message: contentfulManagement.NewOptString("Error message"),
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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := util.ErrorDetailFromContentfulManagementResponse(test.response, test.err)

			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestOptStringToStringValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    contentfulManagement.OptString
		expected basetypes.StringValue
	}{
		"set": {
			input:    contentfulManagement.NewOptString("string"),
			expected: types.StringValue("string"),
		},
		"set: empty": {
			input:    contentfulManagement.NewOptString(""),
			expected: types.StringValue(""),
		},
		"unset": {
			input:    contentfulManagement.OptString{},
			expected: types.StringNull(),
		},
		"unset: non-empty": {
			input:    contentfulManagement.OptString{Value: "string"},
			expected: types.StringNull(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := util.OptStringToStringValue(test.input)

			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestStringValueToOptString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    basetypes.StringValue
		expected contentfulManagement.OptString
	}{
		"set": {
			input:    types.StringValue("string"),
			expected: contentfulManagement.NewOptString("string"),
		},
		"set: empty": {
			input:    types.StringValue(""),
			expected: contentfulManagement.NewOptString(""),
		},
		"null": {
			input:    types.StringNull(),
			expected: contentfulManagement.OptString{},
		},
		"unknown": {
			input:    types.StringUnknown(),
			expected: contentfulManagement.NewOptString(""),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := util.StringValueToOptString(test.input)

			assert.EqualValues(t, test.expected, actual)
		})
	}
}

func TestStringValueToOptNilString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input    basetypes.StringValue
		expected contentfulManagement.OptNilString
	}{
		"set": {
			input:    types.StringValue("string"),
			expected: contentfulManagement.NewOptNilString("string"),
		},
		"set: empty": {
			input:    types.StringValue(""),
			expected: contentfulManagement.NewOptNilString(""),
		},
		"null": {
			input:    types.StringNull(),
			expected: contentfulManagement.NewOptNilStringNull(),
		},
		"unknown": {
			input:    types.StringUnknown(),
			expected: contentfulManagement.OptNilString{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := util.StringValueToOptNilString(test.input)

			assert.EqualValues(t, test.expected, actual)
		})
	}
}
