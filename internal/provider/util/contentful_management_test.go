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
