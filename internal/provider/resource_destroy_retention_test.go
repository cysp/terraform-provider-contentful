package provider_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccResourceDestroyRetainsConfiguration(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ name, config, path, parentPath, importID string }{
		{name: "editor_interface", config: `
resource "contentful_editor_interface" "test" {
 space_id = "space"
 environment_id = "master"
 content_type_id = "type"
 controls = []
}
`, path: "/spaces/space/environments/master/content_types/type/editor_interface", parentPath: "/spaces/space/environments/master/content_types/type", importID: "space/master/type"},
		{name: "space_enablements", config: `
resource "contentful_space_enablements" "test" {
 space_id = "space"
 cross_space_links = true
 space_templates = true
}
`, path: "/spaces/space/enablements", parentPath: "/spaces/space/environments/master", importID: "space"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")
			server.SetContentType("space", "master", "type", cm.ContentTypeRequestData{Name: "Type"})
			server.SetEditorInterface("space", "master", "type", cm.EditorInterfaceData{})

			var before map[string]any

			unchanged := func(_ *terraform.State) error {
				require.Equal(t, before, remoteDeleteGet(t, server, test.path, http.StatusOK), "destroy must retain remote configuration and version")
				remoteDeleteGet(t, server, test.parentPath, http.StatusOK)

				return nil
			}
			ContentfulProviderMockedResourceTest(t, server, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.config, Check: func(_ *terraform.State) error {
					before = remoteDeleteGet(t, server, test.path, http.StatusOK)

					return nil
				}},
				{Config: test.config, Destroy: true, Check: unchanged},
				{Config: test.config, ResourceName: "contentful_" + test.name + ".test", ImportState: true, ImportStateId: test.importID, ImportStatePersist: true},
				{Config: test.config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}, Check: unchanged},
			}})
		})
	}
}
