package contentfulmanagement

import "time"

func NewAppSigningSecretSys(organizationID, appDefinitionID, userID string) AppSigningSecretSys {
	now := time.Now().UTC()

	return AppSigningSecretSys{
		Type:          AppSigningSecretSysTypeAppSigningSecret,
		Organization:  NewOrganizationLink(organizationID),
		AppDefinition: NewAppDefinitionLink(appDefinitionID),
		CreatedAt:     now,
		CreatedBy:     NewUserLink(userID),
		UpdatedAt:     now,
		UpdatedBy:     NewUserLink(userID),
	}
}
