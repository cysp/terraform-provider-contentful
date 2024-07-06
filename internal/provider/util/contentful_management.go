package util

import (
	"fmt"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func ErrorDetailFromContentfulManagementResponse(response interface{}, err error) string {
	if response, ok := response.(*contentfulManagement.ErrorStatusCode); ok {
		return response.Response.Message
	}

	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%v", response)
}
