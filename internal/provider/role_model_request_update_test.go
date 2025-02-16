//nolint:dupl
package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestRoleModelRoundTripToUpdateRoleReq(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	model := provider.RoleModel{}
	model.ReadFromResponse(ctx, &cm.Role{
		Sys: cm.RoleSys{
			ID: "abcdef",
		},
		Name:        "Reader",
		Description: cm.NewOptNilString("Read access to content"),
		Permissions: map[string]cm.RolePermissionsItem{
			"ContentDelivery":   cm.NewStringRolePermissionsItem("all"),
			"ContentManagement": cm.NewStringArrayRolePermissionsItem([]string{"read"}),
		},
		Policies: []cm.RolePoliciesItem{
			{
				Effect:     "allow",
				Actions:    cm.NewStringRolePoliciesItemActions("all"),
				Constraint: []byte("{\"sys.type\":\"Entry\"}"),
			},
			{
				Effect:     "deny",
				Actions:    cm.NewStringArrayRolePoliciesItemActions([]string{"delete"}),
				Constraint: []byte("{\"sys.type\":\"Entry\"}"),
			},
			{
				Effect:  "allow",
				Actions: cm.NewStringArrayRolePoliciesItemActions([]string{"all"}),
			},
		},
	})

	assert.Equal(t, "Reader", model.Name.ValueString())
	assert.Equal(t, "Read access to content", model.Description.ValueString())
	assert.Equal(t, "abcdef", model.RoleId.ValueString())

	req, diags := model.ToUpdateRoleReq(ctx)

	assert.Equal(t, "Reader", req.Name)
	assert.True(t, req.Description.Set)
	assert.Equal(t, "Read access to content", req.Description.Value)

	assert.Len(t, req.Permissions, 2)
	assert.Equal(t, cm.NewStringUpdateRoleReqPermissionsItem("all"), req.Permissions["ContentDelivery"])
	assert.Equal(t, cm.NewStringArrayUpdateRoleReqPermissionsItem([]string{"read"}), req.Permissions["ContentManagement"])

	assert.Len(t, req.Policies, 3)
	assert.Equal(t, cm.UpdateRoleReqPoliciesItem{
		Effect:     "allow",
		Actions:    cm.NewStringUpdateRoleReqPoliciesItemActions("all"),
		Constraint: []byte("{\"sys.type\":\"Entry\"}"),
	}, req.Policies[0])
	assert.Equal(t, cm.UpdateRoleReqPoliciesItem{
		Effect:     "deny",
		Actions:    cm.NewStringArrayUpdateRoleReqPoliciesItemActions([]string{"delete"}),
		Constraint: []byte("{\"sys.type\":\"Entry\"}"),
	}, req.Policies[1])
	assert.Equal(t, cm.UpdateRoleReqPoliciesItem{
		Effect:  "allow",
		Actions: cm.NewStringUpdateRoleReqPoliciesItemActions("all"),
	}, req.Policies[2])

	assert.Empty(t, diags)
}
