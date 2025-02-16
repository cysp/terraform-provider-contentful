package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type ContentfulProviderData struct {
	client *cm.Client

	editorInterfaceVersionOffset *ContentfulContentTypeCounter
}
