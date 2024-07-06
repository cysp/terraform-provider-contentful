package util_test

import (
	"errors"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
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
					Message: "message",
				},
			},
			expected: "message",
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
