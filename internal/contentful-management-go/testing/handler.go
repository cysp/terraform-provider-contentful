package testing

import (
	"sync"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type Handler struct {
	mu sync.Mutex

	me *cm.User

	teams OrganizationMap[*cm.Team]

	teamSpaceMemberships cm.SpaceMap[*cm.TeamSpaceMembership]

	personalAccessTokens map[string]*cm.PersonalAccessToken

	enablements map[string]*cm.SpaceEnablement

	apiKeys        cm.SpaceMap[*cm.ApiKey]
	previewAPIKeys cm.SpaceMap[*cm.PreviewApiKey]

	environments       cm.SpaceMap[*cm.Environment]
	environmentAliases cm.SpaceMap[*cm.EnvironmentAlias]

	marketplaceAppDefinitions map[string]*cm.AppDefinition

	appDefinitions                 OrganizationMap[*cm.AppDefinition]
	appDefinitionResourceProviders OrganizationMap[*cm.ResourceProvider]
	appDefinitionResourceTypes     OrganizationMap[*cm.ResourceType]
	appInstallations               cm.SpaceEnvironmentMap[*cm.AppInstallation]
	appSigningSecrets              OrganizationMap[*cm.AppSigningSecret]

	contentTypes     cm.SpaceEnvironmentMap[*cm.ContentType]
	editorInterfaces cm.SpaceEnvironmentMap[*cm.EditorInterface]

	entries cm.SpaceEnvironmentMap[*cm.Entry]

	extensions cm.SpaceEnvironmentMap[*cm.Extension]

	roles cm.SpaceMap[*cm.Role]

	webhookDefinitions cm.SpaceMap[*cm.WebhookDefinition]
}

var _ cm.Handler = (*Handler)(nil)

func NewHandler() *Handler {
	return &Handler{
		mu: sync.Mutex{},

		teams:                          NewOrganizationMap[*cm.Team](),
		teamSpaceMemberships:           cm.NewSpaceMap[*cm.TeamSpaceMembership](),
		personalAccessTokens:           make(map[string]*cm.PersonalAccessToken),
		enablements:                    make(map[string]*cm.SpaceEnablement),
		apiKeys:                        cm.NewSpaceMap[*cm.ApiKey](),
		previewAPIKeys:                 cm.NewSpaceMap[*cm.PreviewApiKey](),
		environments:                   cm.NewSpaceMap[*cm.Environment](),
		environmentAliases:             cm.NewSpaceMap[*cm.EnvironmentAlias](),
		marketplaceAppDefinitions:      make(map[string]*cm.AppDefinition),
		appDefinitions:                 NewOrganizationMap[*cm.AppDefinition](),
		appDefinitionResourceProviders: NewOrganizationMap[*cm.ResourceProvider](),
		appDefinitionResourceTypes:     NewOrganizationMap[*cm.ResourceType](),
		appInstallations:               cm.NewSpaceEnvironmentMap[*cm.AppInstallation](),
		appSigningSecrets:              NewOrganizationMap[*cm.AppSigningSecret](),
		contentTypes:                   cm.NewSpaceEnvironmentMap[*cm.ContentType](),
		editorInterfaces:               cm.NewSpaceEnvironmentMap[*cm.EditorInterface](),
		entries:                        cm.NewSpaceEnvironmentMap[*cm.Entry](),
		extensions:                     cm.NewSpaceEnvironmentMap[*cm.Extension](),
		roles:                          cm.NewSpaceMap[*cm.Role](),
		webhookDefinitions:             cm.NewSpaceMap[*cm.WebhookDefinition](),
	}
}
