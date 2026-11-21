package cmtesting

import (
	"reflect"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
)

func NewWebhookDefinitionFromFields(spaceID, webhookDefinitionID string, webhookDefinitionFields cm.WebhookDefinitionData) cm.WebhookDefinition {
	webhookDefinition := cm.WebhookDefinition{
		Sys: cm.NewWebhookDefinitionSys(spaceID, webhookDefinitionID),
	}

	setWebhookDefinitionFields(&webhookDefinition, webhookDefinitionFields)

	return webhookDefinition
}

func UpdateWebhookDefinitionFromFields(webhook *cm.WebhookDefinition, fields cm.WebhookDefinitionData) *cm.ErrorStatusCode {
	updated := *webhook
	setWebhookDefinitionFields(&updated, fields)

	for i, header := range updated.Headers {
		if header.Value.IsSet() || !header.Secret.Or(false) {
			continue
		}

		index := slices.IndexFunc(webhook.Headers, func(stored cm.WebhookDefinitionHeader) bool {
			return stored.Key == header.Key
		})
		if index < 0 {
			return missingWebhookHeaderValueError(header.Key)
		}

		updated.Headers[i] = webhook.Headers[index]
	}

	if !reflect.DeepEqual(*webhook, updated) {
		updated.Sys.Version++
		*webhook = updated
	}

	return nil
}

func missingWebhookHeaderValueError(key string) *cm.ErrorStatusCode {
	var details jx.Encoder
	details.ObjStart()
	details.FieldStart("errors")
	details.Str("Header " + key + " is missing a value and cannot be set")
	details.ObjEnd()

	return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), details.Bytes())
}

func setWebhookDefinitionFields(webhookDefinition *cm.WebhookDefinition, webhookDefinitionFields cm.WebhookDefinitionData) {
	webhookDefinition.Name = webhookDefinitionFields.Name
	webhookDefinition.URL = webhookDefinitionFields.URL
	webhookDefinition.HttpBasicUsername = webhookDefinitionFields.HttpBasicUsername
	webhookDefinition.HttpBasicPassword = webhookDefinitionFields.HttpBasicPassword

	webhookDefinition.Headers = append(cm.WebhookDefinitionHeaders{}, webhookDefinitionFields.Headers...)
	webhookDefinition.Topics = webhookDefinitionFields.Topics
	webhookDefinition.Filters = webhookDefinitionFields.Filters
	webhookDefinition.Active = cm.NewOptBool(webhookDefinitionFields.Active.Or(true))
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
