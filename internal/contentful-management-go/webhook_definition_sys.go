package contentfulmanagement

func NewWebhookDefinitionSys(spaceID, webhookDefinitionID string) WebhookDefinitionSys {
	return WebhookDefinitionSys{
		Type:  WebhookDefinitionSysTypeWebhookDefinition,
		ID:    webhookDefinitionID,
		Space: NewSpaceLink(spaceID),
	}
}
