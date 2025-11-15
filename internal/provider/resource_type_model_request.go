package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (r *ResourceTypeModel) ToResourceTypeData(_ context.Context, _ path.Path) (cm.ResourceTypeData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	resourceTypeFields := cm.ResourceTypeData{
		Name: r.Name.ValueString(),
		DefaultFieldMapping: cm.ResourceTypeDefaultFieldMapping{
			Title:       r.DefaultFieldMapping.Title.ValueString(),
			Subtitle:    cm.NewOptPointerString(r.DefaultFieldMapping.Subtitle.ValueStringPointer()),
			Description: cm.NewOptPointerString(r.DefaultFieldMapping.Description.ValueStringPointer()),
			ExternalUrl: cm.NewOptPointerString(r.DefaultFieldMapping.ExternalURL.ValueStringPointer()),
		},
	}

	if r.DefaultFieldMapping.Image != nil {
		resourceTypeFields.DefaultFieldMapping.Image.SetTo(cm.ResourceTypeDefaultFieldMappingImage{
			URL:     r.DefaultFieldMapping.Image.URL.ValueString(),
			AltText: cm.NewOptPointerString(r.DefaultFieldMapping.Image.AltText.ValueStringPointer()),
		})
	}

	if r.DefaultFieldMapping.Badge != nil {
		resourceTypeFields.DefaultFieldMapping.Badge.SetTo(cm.ResourceTypeDefaultFieldMappingBadge{
			Label:   r.DefaultFieldMapping.Badge.Label.ValueString(),
			Variant: r.DefaultFieldMapping.Badge.Variant.ValueString(),
		})
	}

	return resourceTypeFields, diags
}
