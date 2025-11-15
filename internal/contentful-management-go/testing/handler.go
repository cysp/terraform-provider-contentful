package testing

import (
	"sync"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type Handler struct {
	mu sync.Mutex

	me *cm.User

	teams OrganizationMap[*cm.Team]

	teamSpaceMemberships SpaceMap[*cm.TeamSpaceMembership]

	personalAccessTokens map[string]*cm.PersonalAccessToken

	enablements map[string]*cm.SpaceEnablement

	apiKeys        SpaceMap[*cm.ApiKey]
	previewAPIKeys SpaceMap[*cm.PreviewApiKey]

	environments       SpaceMap[*cm.Environment]
	environmentAliases SpaceMap[*cm.EnvironmentAlias]

	marketplaceAppDefinitions map[string]*cm.AppDefinition

	appDefinitions                 OrganizationMap[*cm.AppDefinition]
	appDefinitionResourceProviders OrganizationMap[*cm.ResourceProvider]
	appDefinitionResourceTypes     OrganizationMap[*cm.ResourceType]
	appInstallations               SpaceEnvironmentMap[*cm.AppInstallation]
	appSigningSecrets              OrganizationMap[*cm.AppSigningSecret]

	contentTypes     SpaceEnvironmentMap[*cm.ContentType]
	editorInterfaces SpaceEnvironmentMap[*cm.EditorInterface]

	entries SpaceEnvironmentMap[*cm.Entry]

	extensions SpaceEnvironmentMap[*cm.Extension]

	roles SpaceMap[*cm.Role]

	webhookDefinitions SpaceMap[*cm.WebhookDefinition]
}

var _ cm.Handler = (*Handler)(nil)

func NewHandler() *Handler {
	return &Handler{
		mu: sync.Mutex{},

		teams:                          NewOrganizationMap[*cm.Team](),
		teamSpaceMemberships:           NewSpaceMap[*cm.TeamSpaceMembership](),
		personalAccessTokens:           make(map[string]*cm.PersonalAccessToken),
		enablements:                    make(map[string]*cm.SpaceEnablement),
		apiKeys:                        NewSpaceMap[*cm.ApiKey](),
		previewAPIKeys:                 NewSpaceMap[*cm.PreviewApiKey](),
		environments:                   NewSpaceMap[*cm.Environment](),
		environmentAliases:             NewSpaceMap[*cm.EnvironmentAlias](),
		marketplaceAppDefinitions:      make(map[string]*cm.AppDefinition),
		appDefinitions:                 NewOrganizationMap[*cm.AppDefinition](),
		appDefinitionResourceProviders: NewOrganizationMap[*cm.ResourceProvider](),
		appDefinitionResourceTypes:     NewOrganizationMap[*cm.ResourceType](),
		appInstallations:               NewSpaceEnvironmentMap[*cm.AppInstallation](),
		appSigningSecrets:              NewOrganizationMap[*cm.AppSigningSecret](),
		contentTypes:                   NewSpaceEnvironmentMap[*cm.ContentType](),
		editorInterfaces:               NewSpaceEnvironmentMap[*cm.EditorInterface](),
		entries:                        NewSpaceEnvironmentMap[*cm.Entry](),
		extensions:                     NewSpaceEnvironmentMap[*cm.Extension](),
		roles:                          NewSpaceMap[*cm.Role](),
		webhookDefinitions:             NewSpaceMap[*cm.WebhookDefinition](),
	}
}
