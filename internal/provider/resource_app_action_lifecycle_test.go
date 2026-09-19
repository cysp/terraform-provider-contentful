package provider_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

// This mocked/live lifecycle creates only a temporary App Definition and actions.
// Function deployment/invocation and spaces/environments are not touched.
//
//nolint:paralleltest // Live acceptance serializes access to the shared account.
func TestAccAppActionResourceLifecycle(t *testing.T) {
	parallelWhenMocked(t)

	organization := os.Getenv("CONTENTFUL_ORGANIZATION_ID")
	if os.Getenv("TF_ACC_MOCKED") != "" {
		organization = "organization"
	}

	if organization == "" {
		t.Skip("requires CONTENTFUL_ORGANIZATION_ID")
	}

	config := fmt.Sprintf(`resource "contentful_app_definition" "test" {
 organization_id=%q
 name="Terraform App Action acceptance"
 locations=[]
}
resource "contentful_app_action" "test" {
 organization_id=contentful_app_definition.test.organization_id
 app_definition_id=contentful_app_definition.test.app_definition_id
 name="Action"
 category="Custom"
 type="endpoint"
 url="https://example.invalid/terraform-action"
 parameters_schema=jsonencode({type="object",properties={message={type="string"}}})
 result_schema=jsonencode({type="object"})
 description="Temporary acceptance action"
}
resource "contentful_app_action" "builtin" {
 organization_id=contentful_app_definition.test.organization_id
 app_definition_id=contentful_app_definition.test.app_definition_id
 name="Builtin"
 category="Entries.v1.0"
 type="endpoint"
 url="https://example.invalid/terraform-action"
}
# CMA accepts the link without a deployed Function. This checks definition
# management only; it must not be interpreted as an execution test.
resource "contentful_app_action" "function" {
 organization_id=contentful_app_definition.test.organization_id
 app_definition_id=contentful_app_definition.test.app_definition_id
 name="Function action"
 category="Custom"
 type="function-invocation"
 function_id="pr755probefunction"
 parameters=jsonencode([
  {id="text",name="Text",type="Symbol"},
  {id="number",name="Number",type="Number",required=false,default=0},
  {id="flag",name="Flag",type="Boolean",default=false},
  {id="choice",name="Choice",type="Enum",options=["a","b"]}
 ])
 description=""
}
data "contentful_app_action" "one" {
 organization_id=contentful_app_action.test.organization_id
 app_definition_id=contentful_app_action.test.app_definition_id
 app_action_id=contentful_app_action.test.app_action_id
}
data "contentful_app_actions" "all" {
 organization_id=contentful_app_definition.test.organization_id
 app_definition_id=contentful_app_definition.test.app_definition_id
 depends_on=[contentful_app_action.test,contentful_app_action.builtin,contentful_app_action.function]
}`, organization)
	updated := strings.NewReplacer(
		`parameters_schema=jsonencode({type="object",properties={message={type="string"}}})`, `parameters=jsonencode([])`,
		` result_schema=jsonencode({type="object"})`, "",
		` description="Temporary acceptance action"`, "",
		`category="Entries.v1.0"`, `category="Notification.v1.0"`,
		`type="function-invocation"
 function_id="pr755probefunction"`, `type="endpoint"
 url="https://example.invalid/function-transition"`,
	).Replace(config)
	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	testAccMockableResource(t, server, resource.TestCase{Steps: []resource.TestStep{
		{Config: config, ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue("data.contentful_app_actions.all", tfjsonpath.New("app_actions"), knownvalue.ListSizeExact(3))}},
		{ResourceName: "contentful_app_action.test", ImportState: true, ImportStateVerify: true},
		{ResourceName: "contentful_app_action.builtin", ImportState: true, ImportStateVerify: true},
		{ResourceName: "contentful_app_action.function", ImportState: true, ImportStateVerify: true},
		{Config: updated, ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue("contentful_app_action.test", tfjsonpath.New("parameters"), knownvalue.StringExact("[]")), statecheck.ExpectKnownValue("contentful_app_action.builtin", tfjsonpath.New("category"), knownvalue.StringExact("Notification.v1.0")), statecheck.ExpectKnownValue("contentful_app_action.function", tfjsonpath.New("function_id"), knownvalue.Null())}},
		{Config: strings.Replace(updated, `type="endpoint"
 url="https://example.invalid/function-transition"`, `type="function-invocation"
 function_id="pr755probefunction"`, 1), ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("contentful_app_action.function", tfjsonpath.New("url"), knownvalue.Null()),
		}},
	}})
}
