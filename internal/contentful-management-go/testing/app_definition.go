package cmtesting

import (
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewAppDefinitionFromFields(organizationID, appDefinitionID string, fields cm.AppDefinitionData) cm.AppDefinition {
	appDefinition := cm.AppDefinition{
		Sys: cm.NewAppDefinitionSys(organizationID, appDefinitionID),
	}

	UpdateAppDefinitionFromFields(&appDefinition, organizationID, appDefinitionID, fields)

	return appDefinition
}

func UpdateAppDefinitionFromFields(appDefinition *cm.AppDefinition, organizationID, appDefinitionID string, fields cm.AppDefinitionData) {
	appDefinition.Sys.ID = appDefinitionID
	appDefinition.Sys.Organization.Sys.ID = organizationID

	appDefinition.Name = fields.Name

	appDefinition.Src = fields.Src
	appDefinition.Bundle = fields.Bundle

	appDefinition.Locations = slices.Clone(fields.Locations)
	for index := range appDefinition.Locations {
		appDefinition.Locations[index].FieldTypes = slices.Clone(fields.Locations[index].FieldTypes)
	}

	appDefinition.Parameters = fields.Parameters
}
