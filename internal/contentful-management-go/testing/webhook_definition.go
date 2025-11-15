package testing

import cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"

func NewWebhookDefinitionFromFields(spaceID, webhookDefinitionID string, webhookDefinitionFields cm.WebhookDefinitionData) cm.WebhookDefinition {
	webhookDefinition := cm.WebhookDefinition{
		Sys: cm.NewWebhookDefinitionSys(spaceID, webhookDefinitionID),
	}

	UpdateWebhookDefinitionFromFields(&webhookDefinition, webhookDefinitionFields)

	return webhookDefinition
}

func UpdateWebhookDefinitionFromFields(webhookDefinition *cm.WebhookDefinition, webhookDefinitionFields cm.WebhookDefinitionData) {
	webhookDefinition.Sys.Version++
	webhookDefinition.Name = webhookDefinitionFields.Name
	webhookDefinition.URL = webhookDefinitionFields.URL
	webhookDefinition.HttpBasicUsername = webhookDefinitionFields.HttpBasicUsername
	webhookDefinition.HttpBasicPassword = webhookDefinitionFields.HttpBasicPassword
	webhookDefinition.Headers = webhookDefinitionFields.Headers
	webhookDefinition.Topics = webhookDefinitionFields.Topics
	webhookDefinition.Filters = webhookDefinitionFields.Filters
	webhookDefinition.Active = webhookDefinitionFields.Active
	convertOptNil(&webhookDefinition.Transformation, &webhookDefinitionFields.Transformation, func(transformation cm.WebhookDefinitionDataTransformation) cm.WebhookDefinitionTransformation {
		return cm.WebhookDefinitionTransformation(transformation)
	})
}
