//nolint:dupl
package provider_test

import (
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestRoleModelRoundTripToCreateRoleReq(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	model := provider.RoleModel{}
	model.ReadFromResponse(ctx, &contentfulManagement.Role{
		Sys: contentfulManagement.RoleSys{
			ID: "abcdef",
		},
		Name:        "Reader",
		Description: contentfulManagement.NewOptNilString("Read access to content"),
		Permissions: map[string]contentfulManagement.RolePermissionsItem{
			"ContentDelivery":   contentfulManagement.NewStringRolePermissionsItem("all"),
			"ContentManagement": contentfulManagement.NewStringArrayRolePermissionsItem([]string{"read"}),
		},
		Policies: []contentfulManagement.RolePoliciesItem{
			{
				Effect:     "allow",
				Actions:    contentfulManagement.NewStringRolePoliciesItemActions("all"),
				Constraint: []byte("{\"sys.type\":\"Entry\"}"),
			},
			{
				Effect:     "deny",
				Actions:    contentfulManagement.NewStringArrayRolePoliciesItemActions([]string{"delete"}),
				Constraint: []byte("{\"sys.type\":\"Entry\"}"),
			},
			{
				Effect:  "allow",
				Actions: contentfulManagement.NewStringArrayRolePoliciesItemActions([]string{"all"}),
			},
		},
	})

	assert.Equal(t, "Reader", model.Name.ValueString())
	assert.Equal(t, "Read access to content", model.Description.ValueString())
	assert.Equal(t, "abcdef", model.RoleId.ValueString())

	req, diags := model.ToCreateRoleReq(ctx)

	assert.Equal(t, "Reader", req.Name)
	assert.True(t, req.Description.Set)
	assert.Equal(t, "Read access to content", req.Description.Value)

	assert.Len(t, req.Permissions, 2)
	assert.Equal(t, contentfulManagement.NewStringCreateRoleReqPermissionsItem("all"), req.Permissions["ContentDelivery"])
	assert.Equal(t, contentfulManagement.NewStringArrayCreateRoleReqPermissionsItem([]string{"read"}), req.Permissions["ContentManagement"])

	assert.Len(t, req.Policies, 3)
	assert.Equal(t, contentfulManagement.CreateRoleReqPoliciesItem{
		Effect:     "allow",
		Actions:    contentfulManagement.NewStringCreateRoleReqPoliciesItemActions("all"),
		Constraint: []byte("{\"sys.type\":\"Entry\"}"),
	}, req.Policies[0])
	assert.Equal(t, contentfulManagement.CreateRoleReqPoliciesItem{
		Effect:     "deny",
		Actions:    contentfulManagement.NewStringArrayCreateRoleReqPoliciesItemActions([]string{"delete"}),
		Constraint: []byte("{\"sys.type\":\"Entry\"}"),
	}, req.Policies[1])
	assert.Equal(t, contentfulManagement.CreateRoleReqPoliciesItem{
		Effect:  "allow",
		Actions: contentfulManagement.NewStringCreateRoleReqPoliciesItemActions("all"),
	}, req.Policies[2])

	assert.Empty(t, diags)
}
