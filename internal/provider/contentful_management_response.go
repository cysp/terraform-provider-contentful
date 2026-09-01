package provider

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func contentfulResponseIsNotFound(response any) bool {
	status, ok := response.(cm.StatusCodeResponse)

	return ok && status.GetStatusCode() == http.StatusNotFound
}
