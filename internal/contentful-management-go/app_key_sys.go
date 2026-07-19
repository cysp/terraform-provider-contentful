package contentfulmanagement

import "time"

func NewAppKeySys(organizationID, appDefinitionID, keyKID, userID string) AppKeySys {
	now := time.Now().UTC()

	return AppKeySys{
		Type:          AppKeySysTypeAppKey,
		Organization:  NewOrganizationLink(organizationID),
		AppDefinition: NewAppDefinitionLink(appDefinitionID),
		CreatedBy:     NewUserLink(userID),
		UpdatedBy:     NewUserLink(userID),
		ID:            keyKID,
		CreatedAt:     NewOptDateTime(now),
		UpdatedAt:     NewOptDateTime(now),
		LastUsedAt:    NewOptNilDateTimeNull(),
	}
}
