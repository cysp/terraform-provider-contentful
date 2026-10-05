package provider_test

import (
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccRoleDataSourcesComposition(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	// Preserve noncanonical key order so both data sources must normalize the response.
	constraint := testJSON(struct {
		Z int   `json:"z"`
		A []int `json:"a"`
	}{Z: 1, A: []int{1, 2}})

	server.SetRole("space", "item-b", cm.RoleData{
		Name: "Second",
		Permissions: cm.RoleDataPermissions{
			"ContentDelivery": cm.NewStringRoleDataPermissionsItem("all"),
			"ContentModel":    cm.NewStringArrayRoleDataPermissionsItem([]string{"read", "read"}),
		},
		Policies: []cm.RoleDataPoliciesItem{
			{Effect: cm.RoleDataPoliciesItemEffectAllow, Actions: cm.NewStringRoleDataPoliciesItemActions("all"), Constraint: []byte(constraint)},
			{Effect: cm.RoleDataPoliciesItemEffectDeny, Actions: cm.NewStringArrayRoleDataPoliciesItemActions([]string{"read", "read"})},
		},
	})
	server.SetRole("space", "item-a", cm.RoleData{Name: "First", Permissions: cm.RoleDataPermissions{}, Policies: []cm.RoleDataPoliciesItem{}})

	var roleReads, roleMutations atomic.Int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/spaces/space/roles") {
			if request.Method == http.MethodGet {
				roleReads.Add(1)
			} else {
				roleMutations.Add(1)
			}
		}

		server.ServeHTTP(w, request)
	})

	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{
			ConfigDirectory: config.TestNameDirectory(),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectUnknownValue("data.contentful_role.selected", tfjsonpath.New("name")),
				plancheck.ExpectUnknownValue("data.contentful_roles.all", tfjsonpath.New("roles")),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("data.contentful_role.selected", tfjsonpath.New("id"), knownvalue.StringExact("space/item-b")),
				statecheck.ExpectKnownValue("data.contentful_role.selected", tfjsonpath.New("role_id"), knownvalue.StringExact("item-b")),
				statecheck.ExpectKnownValue("data.contentful_role.selected", tfjsonpath.New("description"), knownvalue.Null()),
				statecheck.ExpectKnownValue("data.contentful_role.selected", tfjsonpath.New("permissions"), knownvalue.MapExact(map[string]knownvalue.Check{
					"ContentDelivery": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("all")}),
					"ContentModel":    knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("read"), knownvalue.StringExact("read")}),
				})),
				statecheck.ExpectKnownValue("data.contentful_role.selected", tfjsonpath.New("policies"), knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{"effect": knownvalue.StringExact("allow"), "actions": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("all")}), "constraint": knownvalue.StringExact(testJSON(map[string]any{"a": []any{1, 2}, "z": 1}))}),
					knownvalue.ObjectExact(map[string]knownvalue.Check{"effect": knownvalue.StringExact("deny"), "actions": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("read"), knownvalue.StringExact("read")}), "constraint": knownvalue.Null()}),
				})),
				statecheck.ExpectKnownValue("data.contentful_roles.all", tfjsonpath.New("id"), knownvalue.StringExact("space")),
				statecheck.ExpectKnownValue("data.contentful_roles.all", tfjsonpath.New("roles"), knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{"role_id": knownvalue.StringExact("item-a"), "name": knownvalue.StringExact("First"), "description": knownvalue.Null(), "permissions": knownvalue.MapExact(map[string]knownvalue.Check{}), "policies": knownvalue.ListExact([]knownvalue.Check{})}),
					knownvalue.ObjectExact(map[string]knownvalue.Check{"role_id": knownvalue.StringExact("item-b"), "name": knownvalue.StringExact("Second"), "description": knownvalue.Null(), "permissions": knownvalue.MapExact(map[string]knownvalue.Check{
						"ContentDelivery": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("all")}),
						"ContentModel":    knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("read"), knownvalue.StringExact("read")}),
					}), "policies": knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{"effect": knownvalue.StringExact("allow"), "actions": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("all")}), "constraint": knownvalue.StringExact(testJSON(map[string]any{"a": []any{1, 2}, "z": 1}))}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{"effect": knownvalue.StringExact("deny"), "actions": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("read"), knownvalue.StringExact("read")}), "constraint": knownvalue.Null()}),
					})}),
				})),
				statecheck.ExpectKnownValue("contentful_team_space_membership.assignment", tfjsonpath.New("roles"), knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("item-b")})),
			},
		},
		{ConfigDirectory: config.TestNameDirectory(), PlanOnly: true},
	}})

	assert.Positive(t, roleReads.Load())
	assert.Zero(t, roleMutations.Load())
}
