package contentfulmanagementtestserver

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

func NewWebhookDefinitionFromFields(spaceID, webhookDefinitionID string, webhookDefinitionFields cm.WebhookDefinitionFields) cm.WebhookDefinition {
	webhookDefinition := cm.WebhookDefinition{
		Sys: NewWebhookDefinitionSys(spaceID, webhookDefinitionID),
	}

	UpdateWebhookDefinitionFromFields(&webhookDefinition, webhookDefinitionFields)

	return webhookDefinition
}

func NewWebhookDefinitionSys(spaceID, webhookDefinitionID string) cm.WebhookDefinitionSys {
	return cm.WebhookDefinitionSys{
		Type: cm.WebhookDefinitionSysTypeWebhookDefinition,
		ID:   webhookDefinitionID,
		Space: cm.SpaceLink{
			Sys: cm.SpaceLinkSys{
				Type:     cm.SpaceLinkSysTypeLink,
				LinkType: cm.SpaceLinkSysLinkTypeSpace,
				ID:       spaceID,
			},
		},
	}
}

func UpdateWebhookDefinitionFromFields(webhookDefinition *cm.WebhookDefinition, webhookDefinitionFields cm.WebhookDefinitionFields) {
	webhookDefinition.Sys.Version++
	webhookDefinition.Name = webhookDefinitionFields.Name
	webhookDefinition.URL = webhookDefinitionFields.URL
	webhookDefinition.HttpBasicUsername = webhookDefinitionFields.HttpBasicUsername
	webhookDefinition.HttpBasicPassword = webhookDefinitionFields.HttpBasicPassword
	webhookDefinition.Headers = webhookDefinitionFields.Headers
	webhookDefinition.Topics = webhookDefinitionFields.Topics
	webhookDefinition.Filters = webhookDefinitionFields.Filters
	webhookDefinition.Active = webhookDefinitionFields.Active
	convertOptNil(&webhookDefinition.Transformation, &webhookDefinitionFields.Transformation, func(transformation cm.WebhookDefinitionFieldsTransformation) cm.WebhookDefinitionTransformation {
		return cm.WebhookDefinitionTransformation(transformation)
	})
}
