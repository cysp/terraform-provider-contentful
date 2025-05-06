package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m *AppDefinitionResourceProviderResourceModel) ReadFromResponse(_ context.Context, res *cm.ResourceProvider) diag.Diagnostics {
	diags := diag.Diagnostics{}

	organizationID := res.Sys.Organization.Sys.ID
	appDefinitionID := res.Sys.AppDefinition.Sys.ID
	resourceProviderID := res.Sys.ID

	m.ID = types.StringValue(organizationID + "/" + appDefinitionID)

	m.OrganizationID = types.StringValue(organizationID)
	m.AppDefinitionID = types.StringValue(appDefinitionID)
	m.ResourceProviderID = types.StringValue(resourceProviderID)

	m.FunctionID = types.StringValue(res.Function.Sys.ID)

	return diags
}
