package cmtesting

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

	headers := make(cm.WebhookDefinitionHeaders, len(webhookDefinitionFields.Headers))
	for index, header := range webhookDefinitionFields.Headers {
		headers[index] = header
		if header.Value.IsSet() || !header.Secret.Or(false) {
			continue
		}

		for _, existingHeader := range webhookDefinition.Headers {
			if existingHeader.Key == header.Key && existingHeader.Secret.Or(false) {
				headers[index].Value = existingHeader.Value

				break
			}
		}
	}

	webhookDefinition.Headers = headers
	webhookDefinition.Topics = webhookDefinitionFields.Topics
	webhookDefinition.Filters = webhookDefinitionFields.Filters
	webhookDefinition.Active = webhookDefinitionFields.Active
	convertOptNil(&webhookDefinition.Transformation, &webhookDefinitionFields.Transformation, func(transformation cm.WebhookDefinitionDataTransformation) cm.WebhookDefinitionTransformation {
		return cm.WebhookDefinitionTransformation(transformation)
	})
}

func redactWebhookDefinitionSecrets(webhookDefinition cm.WebhookDefinition) cm.WebhookDefinition {
	redacted := webhookDefinition
	redacted.HttpBasicPassword.Reset()
	redacted.Headers = make(cm.WebhookDefinitionHeaders, len(webhookDefinition.Headers))
	copy(redacted.Headers, webhookDefinition.Headers)

	for index := range redacted.Headers {
		if redacted.Headers[index].Secret.Or(false) {
			redacted.Headers[index].Value.Reset()
		}
	}

	return redacted
}
