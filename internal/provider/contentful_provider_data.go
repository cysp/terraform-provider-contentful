package provider

import (
	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type ContentfulProviderData struct {
	client *contentfulManagement.Client

	editorInterfaceVersionOffset *ContentfulContentTypeCounter
}
