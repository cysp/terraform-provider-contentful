package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// NewPreviewEnvironmentResourceWithClient is available only in test builds.
//
//nolint:ireturn
func NewPreviewEnvironmentResourceWithClient(client *cm.Client) resource.Resource {
	return &previewEnvironmentResource{providerData: ContentfulProviderData{client: client}}
}
