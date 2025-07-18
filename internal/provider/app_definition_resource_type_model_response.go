package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewAppDefinitionResourceTypeResourceModelFromResponse(_ context.Context, response cm.ResourceType) (AppDefinitionResourceTypeModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	organizationID := response.Sys.Organization.Sys.ID
	appDefinitionID := response.Sys.AppDefinition.Sys.ID
	resourceProviderID := response.Sys.ResourceProvider.Sys.ID
	resourceTypeID := response.Sys.ID

	model := AppDefinitionResourceTypeModel{
		ID:                 types.StringValue(organizationID + "/" + appDefinitionID + "/" + resourceTypeID),
		OrganizationID:     types.StringValue(organizationID),
		AppDefinitionID:    types.StringValue(appDefinitionID),
		ResourceProviderID: types.StringValue(resourceProviderID),
		ResourceTypeID:     types.StringValue(resourceTypeID),
	}

	model.Name = types.StringValue(response.Name)

	defaultFieldMapping := ResourceTypeFieldMapping{
		Title:       types.StringValue(response.DefaultFieldMapping.Title),
		Description: types.StringPointerValue(response.DefaultFieldMapping.Description.ValueStringPointer()),
		Subtitle:    types.StringPointerValue(response.DefaultFieldMapping.Subtitle.ValueStringPointer()),
		ExternalURL: types.StringPointerValue(response.DefaultFieldMapping.ExternalUrl.ValueStringPointer()),
	}

	if image, ok := response.DefaultFieldMapping.Image.Get(); ok {
		defaultFieldMapping.Image = &ResourceTypeFieldMappingImage{
			URL:     types.StringValue(image.URL),
			AltText: types.StringPointerValue(image.AltText.ValueStringPointer()),
		}
	}

	if badge, ok := response.DefaultFieldMapping.Badge.Get(); ok {
		defaultFieldMapping.Badge = &ResourceTypeFieldMappingBadge{
			Label:   types.StringValue(badge.Label),
			Variant: types.StringValue(badge.Variant),
		}
	}

	model.DefaultFieldMapping = &defaultFieldMapping

	return model, diags
}
