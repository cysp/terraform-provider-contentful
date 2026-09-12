package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Parents are seeded outside Terraform so their cleanup cannot hide a skipped
// child deletion. The sibling stays managed while only the target is removed.
func TestAccResourceDeleteRemoteState(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		resourceType string
		attributes   string
		remotePath   string
		version      string
	}{
		{resourceType: "app_key", remotePath: "/organizations/organization/app_definitions/{app_definition_id}/keys/{key_kid}"},
		{resourceType: "app_definition", attributes: `organization_id = "organization"
name = "@name@"
locations = []`, remotePath: "/organizations/organization/app_definitions/{app_definition_id}", version: ""},
		{resourceType: "app_installation", attributes: `space_id = "space"
environment_id = "master"
app_definition_id = "@name@"`, remotePath: "/spaces/space/environments/master/app_installations/{app_definition_id}", version: ""},
		{resourceType: "app_signing_secret", attributes: `organization_id = "organization"
app_definition_id = "@name@"
value = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`, remotePath: "/organizations/organization/app_definitions/{app_definition_id}/signing_secret", version: ""},
		{resourceType: "delivery_api_key", attributes: `space_id = "space"
name = "@name@"
environments = ["master"]`, remotePath: "/spaces/space/api_keys/{api_key_id}", version: ""},
		{resourceType: "content_type", attributes: `space_id = "space"
environment_id = "master"
content_type_id = "@name@"
name = "@name@"
description = ""
display_field = ""
fields = []`, remotePath: "/spaces/space/environments/master/content_types/{content_type_id}", version: ""},
		{resourceType: "entry", attributes: `space_id = "space"
environment_id = "master"
content_type_id = "entry-type"
entry_id = "@name@"
fields = {}`, remotePath: "/spaces/space/environments/master/entries/{entry_id}", version: ""},
		{resourceType: "environment", attributes: `space_id = "space"
environment_id = "@name@"
name = "@name@"`, remotePath: "/spaces/space/environments/{environment_id}", version: ""},
		{resourceType: "environment_alias", attributes: `space_id = "space"
environment_alias_id = "@name@"
target_environment_id = "master"`, remotePath: "/spaces/space/environment_aliases/{environment_alias_id}", version: ""},
		{resourceType: "extension", attributes: `space_id = "space"
environment_id = "master"
extension_id = "@name@"
extension = {
name = "@name@"
srcdoc = "<title>Extension</title>"
field_types = [{type = "Symbol"}]
}`, remotePath: "/spaces/space/environments/master/extensions/{extension_id}", version: ""},
		{resourceType: "personal_access_token", attributes: `name = "@name@"
scopes = ["content_management_read"]
expires_in = 300`, remotePath: "/users/me/access_tokens/{id}", version: ""},
		{resourceType: "preview_environment", attributes: `space_id = "space"
preview_environment_id = "@name@"
name = "@name@"
content_type_configurations = {}`, remotePath: "/spaces/space/preview_environments/{preview_environment_id}", version: ""},
		{resourceType: "resource_provider", attributes: `organization_id = "organization"
app_definition_id = "@name@"
resource_provider_id = "provider"
function_id = "function"`, remotePath: "/organizations/organization/app_definitions/{app_definition_id}/resource_provider", version: ""},
		{resourceType: "resource_type", attributes: `organization_id = "organization"
app_definition_id = "types"
default_field_mapping = {title = "{ /name }"}
resource_type_id = "provider:@name@"
name = "@name@"`, remotePath: "/organizations/organization/app_definitions/types/resource_provider/resource_types/{resource_type_id}", version: ""},
		{resourceType: "role", attributes: `space_id = "space"
name = "@name@"
permissions = {}
policies = []`, remotePath: "/spaces/space/roles/{role_id}", version: ""},
		{resourceType: "tag", attributes: `space_id = "space"
environment_id = "master"
visibility = "private"
tag_id = "@name@"
name = "@name@"`, remotePath: "/spaces/space/environments/master/tags/{tag_id}", version: "1"},
		{resourceType: "taxonomy_concept", attributes: `organization_id = "organization"
concept_id = "@name@"
pref_label = {"en-US" = "@name@"}`, remotePath: "/organizations/organization/taxonomy/concepts/{concept_id}", version: "1"},
		{resourceType: "taxonomy_concept_scheme", attributes: `organization_id = "organization"
concept_scheme_id = "@name@"
pref_label = {"en-US" = "@name@"}`, remotePath: "/organizations/organization/taxonomy/concept-schemes/{concept_scheme_id}", version: "1"},
		{resourceType: "team", attributes: `organization_id = "organization"
name = "@name@"`, remotePath: "/organizations/organization/teams/{team_id}", version: ""},
		{resourceType: "team_space_membership", attributes: `space_id = "space"
team_id = "@name@"
admin = true
roles = []`, remotePath: "/spaces/space/team_space_memberships/{team_space_membership_id}", version: ""},
		{resourceType: "webhook", attributes: `space_id = "space"
name = "@name@"
url = "https://example.com/@name@"
topics = ["Entry.publish"]`, remotePath: "/spaces/space/webhook_definitions/{webhook_id}", version: ""},
	} {
		t.Run(test.resourceType, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")

			for _, id := range []string{"target", "sibling", "types"} {
				server.SetAppDefinition("organization", id, cm.AppDefinitionData{Name: id})
			}

			server.SetResourceProvider("organization", "types", cm.ResourceProviderRequest{
				Sys: cm.NewResourceProviderRequestSys("provider"), Type: cm.ResourceProviderRequestTypeFunction, Function: cm.NewFunctionLink("function"),
			})
			server.SetContentType("space", "master", "entry-type", cm.ContentTypeRequestData{Name: "Entry", Fields: []cm.ContentTypeRequestDataFieldsItem{}})
			recorder := &remoteDeleteRecorder{next: server}

			keys := map[string]testAccAppKeyJWKData{}
			if test.resourceType == "app_key" {
				keys["target"], keys["sibling"] = testAccAppKeyJWK(t, 0), testAccAppKeyJWK(t, 1)
			}

			configFor := func(name string) string {
				if test.resourceType == "app_key" {
					return strings.Replace(testAccAppKeyConfig("organization", name, keys[name], ""), `"test"`, fmt.Sprintf("%q", name), 1)
				}

				return fmt.Sprintf("resource %q %q {\n%s\n}\n", "contentful_"+test.resourceType, name, strings.ReplaceAll(test.attributes, "@name@", name))
			}

			var (
				targetPath, siblingPath string
				siblingBefore           map[string]any
			)

			testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{
				{Config: configFor("target") + configFor("sibling"), Check: func(state *terraform.State) error {
					targetPath = remoteDeletePath(t, state, test.resourceType, "target", test.remotePath)
					siblingPath = remoteDeletePath(t, state, test.resourceType, "sibling", test.remotePath)
					require.NotEqual(t, targetPath, siblingPath)
					remoteDeleteGet(t, server, targetPath, http.StatusOK)
					siblingBefore = remoteDeleteGet(t, server, siblingPath, http.StatusOK)

					recorder.reset()

					return nil
				}},
				{Config: configFor("sibling"), Check: func(_ *terraform.State) error {
					assert.Equal(t, siblingBefore, remoteDeleteGet(t, server, siblingPath, http.StatusOK), "sibling must remain unchanged")

					method, path := http.MethodDelete, targetPath
					if test.resourceType == "personal_access_token" {
						method, path = http.MethodPut, targetPath+"/revoked"
						remote := remoteDeleteGet(t, server, targetPath, http.StatusOK)
						assert.NotEmpty(t, remote["revokedAt"], "remote token must be revoked after removal")
					} else {
						remoteDeleteGet(t, server, targetPath, http.StatusNotFound)
					}

					expected := []remoteDeleteRequest{{method: method, path: path, version: test.version}}
					if test.resourceType == "entry" || test.resourceType == "content_type" {
						expected = append([]remoteDeleteRequest{{method: http.MethodDelete, path: targetPath + "/published"}}, expected...)
					}

					assert.Equal(t, expected, recorder.requests(), "exact deletion target and version")

					return nil
				}},
			}})
		})
	}
}

type remoteDeleteRequest struct{ method, path, version string }
type remoteDeleteRecorder struct {
	next      http.Handler
	mu        sync.Mutex
	mutations []remoteDeleteRequest
}

func (r *remoteDeleteRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodDelete || strings.HasSuffix(req.URL.Path, "/revoked") {
		r.mu.Lock()
		r.mutations = append(r.mutations, remoteDeleteRequest{method: req.Method, path: req.URL.Path, version: req.Header.Get("X-Contentful-Version")})
		r.mu.Unlock()
	}

	r.next.ServeHTTP(w, req)
}
func (r *remoteDeleteRecorder) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mutations = nil
}
func (r *remoteDeleteRecorder) requests() []remoteDeleteRequest {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]remoteDeleteRequest(nil), r.mutations...)
}
func remoteDeletePath(t *testing.T, state *terraform.State, resourceType, name, pattern string) string {
	t.Helper()

	object := state.RootModule().Resources["contentful_"+resourceType+"."+name]
	require.NotNil(t, object)

	for key, value := range object.Primary.Attributes {
		pattern = strings.ReplaceAll(pattern, "{"+key+"}", value)
	}

	require.NotContains(t, pattern, "{")

	return pattern
}
func remoteDeleteGet(t *testing.T, server http.Handler, path string, status int) map[string]any {
	t.Helper()

	response := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	server.ServeHTTP(response, request)
	require.Equal(t, status, response.Code, "remote state at %s", path)

	var body map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))

	return body
}
