package contentfulmanagementtestserver

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

type ContentfulManagementTestServer struct {
	mu *sync.Mutex

	httpTestServer *httptest.Server
	serveMux       *http.ServeMux

	me *cm.User

	personalAccessTokenIDsToCreate []string
	personalAccessTokens           map[string]*cm.PersonalAccessToken

	roleIdsToCreate []string
	roles           SpacedMap[*cm.Role]
}

func NewContentfulManagementTestServer() *ContentfulManagementTestServer {
	testserver := &ContentfulManagementTestServer{
		mu:                   &sync.Mutex{},
		personalAccessTokens: make(map[string]*cm.PersonalAccessToken),
		roleIdsToCreate:      make([]string, 0),
		roles:                NewSpacedMap[*cm.Role](),
	}

	testserver.serveMux = http.NewServeMux()
	testserver.httpTestServer = httptest.NewServer(testserver.serveMux)

	testserver.setupUserHandler()
	testserver.setupPersonalAccessTokenHandlers()
	testserver.setupSpaceRoleHandlers()

	return testserver
}

func (ts *ContentfulManagementTestServer) Server() *httptest.Server {
	return ts.httpTestServer
}

func (ts *ContentfulManagementTestServer) Reset() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.me = nil

	ts.personalAccessTokenIDsToCreate = nil
	ts.personalAccessTokens = make(map[string]*cm.PersonalAccessToken)

	ts.roles.Clear()
}

func (ts *ContentfulManagementTestServer) generateResourceID() string {
	return RandStringBytes(8) //nolint:mnd
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	bytes := make([]byte, n)

	for i := range bytes {
		//nolint:gosec
		bytes[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(bytes)
}
