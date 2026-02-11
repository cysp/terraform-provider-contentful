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

	appDefinitions                 map[string]*cm.AppDefinition
	appDefinitionResourceProviders map[string]*cm.ResourceProvider
	appDefinitionResourceTypes     map[string]*cm.ResourceType
	appSigningSecrets              map[string]*cm.AppSigningSecret

	appInstallations cm.SpaceEnvironmentMap[*cm.AppInstallation]

	contentTypes     cm.SpaceEnvironmentMap[*cm.ContentType]
	editorInterfaces cm.SpaceEnvironmentMap[*cm.EditorInterface]

	entries cm.SpaceEnvironmentMap[*cm.Entry]

	extensions cm.SpaceEnvironmentMap[*cm.Extension]

	roles cm.SpaceMap[*cm.Role]

	tags cm.SpaceEnvironmentMap[*cm.Tag]

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
		appDefinitions:                 make(map[string]*cm.AppDefinition),
		appDefinitionResourceProviders: make(map[string]*cm.ResourceProvider),
		appDefinitionResourceTypes:     make(map[string]*cm.ResourceType),
		appSigningSecrets:              make(map[string]*cm.AppSigningSecret),
		appInstallations:               cm.NewSpaceEnvironmentMap[*cm.AppInstallation](),
		contentTypes:                   cm.NewSpaceEnvironmentMap[*cm.ContentType](),
		editorInterfaces:               cm.NewSpaceEnvironmentMap[*cm.EditorInterface](),
		entries:                        cm.NewSpaceEnvironmentMap[*cm.Entry](),
		extensions:                     cm.NewSpaceEnvironmentMap[*cm.Extension](),
		roles:                          cm.NewSpaceMap[*cm.Role](),
		tags:                           cm.NewSpaceEnvironmentMap[*cm.Tag](),
		webhookDefinitions:             cm.NewSpaceMap[*cm.WebhookDefinition](),
	}
}

func (h *Handler) RegisterSpaceEnvironment(spaceID, environmentID, status string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.registerSpaceEnvironment(spaceID, environmentID, status)
}

func (h *Handler) registerSpaceEnvironment(spaceID, environmentID, status string) {
	if h.environments.Get(spaceID, environmentID) != nil {
		return
	}

	environment := NewEnvironmentFromEnvironmentData(spaceID, environmentID, status, cm.EnvironmentData{
		Name: environmentID,
	})
	h.environments.Set(spaceID, environmentID, &environment)
}
