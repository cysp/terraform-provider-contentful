package contentfulmanagementtestserver

import (
	"net/http"
	"net/http/httptest"
	"sync"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

const (
	NonexistentID = "nonexistent"
)

type ContentfulManagementTestServer struct {
	mu *sync.Mutex

	httpTestServer *httptest.Server
	serveMux       *http.ServeMux

	me *cm.User

	personalAccessTokens map[string]*cm.PersonalAccessToken

	enablements map[string]*cm.SpaceEnablement

	apiKeys        SpaceMap[*cm.ApiKey]
	previewAPIKeys SpaceMap[*cm.PreviewApiKey]

	appDefinitions                 OrganizationMap[*cm.AppDefinition]
	appDefinitionResourceProviders OrganizationMap[*cm.ResourceProvider]
	appDefinitionResourceTypes     OrganizationMap[*cm.ResourceType]
	appInstallations               SpaceEnvironmentMap[*cm.AppInstallation]

	contentTypes     SpaceEnvironmentMap[*cm.ContentType]
	editorInterfaces SpaceEnvironmentMap[*cm.EditorInterface]

	extensions SpaceEnvironmentMap[*cm.Extension]

	roles SpaceMap[*cm.Role]

	webhookDefinitions SpaceMap[*cm.WebhookDefinition]
}

func NewContentfulManagementTestServer() *ContentfulManagementTestServer {
	testserver := &ContentfulManagementTestServer{
		mu:                             &sync.Mutex{},
		personalAccessTokens:           make(map[string]*cm.PersonalAccessToken),
		enablements:                    make(map[string]*cm.SpaceEnablement),
		apiKeys:                        NewSpaceMap[*cm.ApiKey](),
		previewAPIKeys:                 NewSpaceMap[*cm.PreviewApiKey](),
		appDefinitions:                 NewOrganizationMap[*cm.AppDefinition](),
		appDefinitionResourceProviders: NewOrganizationMap[*cm.ResourceProvider](),
		appDefinitionResourceTypes:     NewOrganizationMap[*cm.ResourceType](),
		appInstallations:               NewSpaceEnvironmentMap[*cm.AppInstallation](),
		contentTypes:                   NewSpaceEnvironmentMap[*cm.ContentType](),
		editorInterfaces:               NewSpaceEnvironmentMap[*cm.EditorInterface](),
		extensions:                     NewSpaceEnvironmentMap[*cm.Extension](),
		roles:                          NewSpaceMap[*cm.Role](),
		webhookDefinitions:             NewSpaceMap[*cm.WebhookDefinition](),
	}

	testserver.serveMux = http.NewServeMux()
	testserver.httpTestServer = httptest.NewServer(testserver.serveMux)

	testserver.setupUserHandler()
	testserver.setupPersonalAccessTokenHandlers()
	testserver.setupOrganizationAppDefinitionHandlers()
	testserver.setupOrganizationAppDefinitionResourceProviderHandlers()
	testserver.setupOrganizationAppDefinitionResourceTypeHandlers()
	testserver.setupSpaceEnablementsHandlers()
	testserver.setupSpaceAPIKeyHandlers()
	testserver.SetupSpaceEnvironmentAppInstallationHandlers()
	testserver.setupSpaceEnvironmentContentTypeHandlers()
	testserver.SetupSpaceEnvironmentExtensionHandlers()
	testserver.setupSpacePreviewAPIKeyHandlers()
	testserver.setupSpaceRoleHandlers()
	testserver.setupSpaceWebhookDefinitionHandlers()

	return testserver
}

func (ts *ContentfulManagementTestServer) Server() *httptest.Server {
	return ts.httpTestServer
}
