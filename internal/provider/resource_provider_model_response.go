package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewResourceProviderResourceModelFromResponse(res cm.ResourceProvider) ResourceProviderModel {
	organizationID := res.Sys.Organization.Sys.ID
	appDefinitionID := res.Sys.AppDefinition.Sys.ID
	resourceProviderID := res.Sys.ID

	model := ResourceProviderModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(organizationID, appDefinitionID),
		ResourceProviderIdentityModel: ResourceProviderIdentityModel{
			OrganizationID:  types.StringValue(organizationID),
			AppDefinitionID: types.StringValue(appDefinitionID),
		},
		ResourceProviderID: types.StringValue(resourceProviderID),
	}

	model.FunctionID = types.StringValue(res.Function.Sys.ID)

	return model
}
