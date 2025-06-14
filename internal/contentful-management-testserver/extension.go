package contentfulmanagementtestserver

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewExtensionFromFields(spaceID, environmentID, appDefinitionID string, ExtensionFields cm.ExtensionFields) cm.Extension {
	Extension := cm.Extension{
		Sys: NewExtensionSys(spaceID, environmentID, appDefinitionID),
	}

	UpdateExtensionFromFields(&Extension, ExtensionFields)

	return Extension
}

func NewExtensionSys(spaceID, environmentID, appDefinitionID string) cm.ExtensionSys {
	return cm.ExtensionSys{
		Type: cm.ExtensionSysTypeExtension,
		Space: cm.SpaceLink{
			Sys: cm.SpaceLinkSys{
				Type:     cm.SpaceLinkSysTypeLink,
				LinkType: cm.SpaceLinkSysLinkTypeSpace,
				ID:       spaceID,
			},
		},
		Environment: cm.EnvironmentLink{
			Sys: cm.EnvironmentLinkSys{
				Type:     cm.EnvironmentLinkSysTypeLink,
				LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
				ID:       environmentID,
			},
		},
	}
}

func UpdateExtensionFromFields(Extension *cm.Extension, ExtensionFields cm.ExtensionFields) {
	Extension.Parameters = ExtensionFields.Parameters
}
