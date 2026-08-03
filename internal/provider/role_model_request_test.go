package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleModelRoundTripToRoleData(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	model := DiagsNoErrorsMust(NewRoleResourceModelFromResponse(ctx, cm.Role{
		Sys:         cm.NewRoleSys("space", "abcdef"),
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
	}))

	assert.Equal(t, "Reader", model.Name.ValueString())
	assert.Equal(t, "Read access to content", model.Description.ValueString())
	assert.Equal(t, "abcdef", model.RoleID.ValueString())

	req, diags := model.ToRoleData()

	assert.Equal(t, "Reader", req.Name)
	assert.True(t, req.Description.Set)
	assert.Equal(t, "Read access to content", req.Description.Value)

	assert.Len(t, req.Permissions, 2)
	assert.Equal(t, cm.NewStringRoleDataPermissionsItem("all"), req.Permissions["ContentDelivery"])
	assert.Equal(t, cm.NewStringArrayRoleDataPermissionsItem([]string{"read"}), req.Permissions["ContentManagement"])

	assert.Len(t, req.Policies, 3)
	assert.Equal(t, cm.RoleDataPoliciesItem{
		Effect:     "allow",
		Actions:    cm.NewStringRoleDataPoliciesItemActions("all"),
		Constraint: []byte("{\"sys.type\":\"Entry\"}"),
	}, req.Policies[0])
	assert.Equal(t, cm.RoleDataPoliciesItem{
		Effect:     "deny",
		Actions:    cm.NewStringArrayRoleDataPoliciesItemActions([]string{"delete"}),
		Constraint: []byte("{\"sys.type\":\"Entry\"}"),
	}, req.Policies[1])
	assert.Equal(t, cm.RoleDataPoliciesItem{
		Effect:  "allow",
		Actions: cm.NewStringRoleDataPoliciesItemActions("all"),
	}, req.Policies[2])

	assert.Empty(t, diags)
}

func TestRoleRequestRejectsUnresolvedValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mutate        func(*RoleModel)
		expectedPaths []string
	}{
		"null name": {
			mutate: func(model *RoleModel) {
				model.Name = types.StringNull()
			},
			expectedPaths: []string{"name"},
		},
		"unknown name": {
			mutate: func(model *RoleModel) {
				model.Name = types.StringUnknown()
			},
			expectedPaths: []string{"name"},
		},
		"unknown description": {
			mutate: func(model *RoleModel) {
				model.Description = types.StringUnknown()
			},
			expectedPaths: []string{"description"},
		},
		"null permissions": {
			mutate: func(model *RoleModel) {
				model.Permissions = NewTypedMapNull[TypedList[types.String]]()
			},
			expectedPaths: []string{"permissions"},
		},
		"unknown permissions": {
			mutate: func(model *RoleModel) {
				model.Permissions = NewTypedMapUnknown[TypedList[types.String]]()
			},
			expectedPaths: []string{"permissions"},
		},
		"null permission actions": {
			mutate: func(model *RoleModel) {
				model.Permissions = NewTypedMap(map[string]TypedList[types.String]{
					"Entry": NewTypedListNull[types.String](),
				})
			},
			expectedPaths: []string{`permissions["Entry"]`},
		},
		"unknown permission actions": {
			mutate: func(model *RoleModel) {
				model.Permissions = NewTypedMap(map[string]TypedList[types.String]{
					"Entry": NewTypedListUnknown[types.String](),
				})
			},
			expectedPaths: []string{`permissions["Entry"]`},
		},
		"null permission action": {
			mutate: func(model *RoleModel) {
				model.Permissions = NewTypedMap(map[string]TypedList[types.String]{
					"Entry": NewTypedList([]types.String{types.StringValue("read"), types.StringNull()}),
				})
			},
			expectedPaths: []string{`permissions["Entry"][1]`},
		},
		"unknown permission action": {
			mutate: func(model *RoleModel) {
				model.Permissions = NewTypedMap(map[string]TypedList[types.String]{
					"Entry": NewTypedList([]types.String{types.StringValue("read"), types.StringUnknown()}),
				})
			},
			expectedPaths: []string{`permissions["Entry"][1]`},
		},
		"null policies": {
			mutate: func(model *RoleModel) {
				model.Policies = NewTypedListNull[TypedObject[RolePolicyValue]]()
			},
			expectedPaths: []string{"policies"},
		},
		"unknown policies": {
			mutate: func(model *RoleModel) {
				model.Policies = NewTypedListUnknown[TypedObject[RolePolicyValue]]()
			},
			expectedPaths: []string{"policies"},
		},
		"null policy": {
			mutate: func(model *RoleModel) {
				model.Policies = NewTypedList([]TypedObject[RolePolicyValue]{NewTypedObjectNull[RolePolicyValue]()})
			},
			expectedPaths: []string{"policies[0]"},
		},
		"unknown policy": {
			mutate: func(model *RoleModel) {
				model.Policies = NewTypedList([]TypedObject[RolePolicyValue]{NewTypedObjectUnknown[RolePolicyValue]()})
			},
			expectedPaths: []string{"policies[0]"},
		},
		"null policy effect": {
			mutate: func(model *RoleModel) {
				model.Policies = rolePoliciesWith(RolePolicyValue{
					Actions:    NewTypedList([]types.String{types.StringValue("read")}),
					Constraint: jsontypes.NewNormalizedNull(),
					Effect:     types.StringNull(),
				})
			},
			expectedPaths: []string{"policies[0].effect"},
		},
		"unknown policy effect": {
			mutate: func(model *RoleModel) {
				model.Policies = rolePoliciesWith(RolePolicyValue{
					Actions:    NewTypedList([]types.String{types.StringValue("read")}),
					Constraint: jsontypes.NewNormalizedNull(),
					Effect:     types.StringUnknown(),
				})
			},
			expectedPaths: []string{"policies[0].effect"},
		},
		"null policy actions": {
			mutate: func(model *RoleModel) {
				model.Policies = rolePoliciesWith(RolePolicyValue{
					Actions:    NewTypedListNull[types.String](),
					Constraint: jsontypes.NewNormalizedNull(),
					Effect:     types.StringValue("allow"),
				})
			},
			expectedPaths: []string{"policies[0].actions"},
		},
		"unknown policy actions": {
			mutate: func(model *RoleModel) {
				model.Policies = rolePoliciesWith(RolePolicyValue{
					Actions:    NewTypedListUnknown[types.String](),
					Constraint: jsontypes.NewNormalizedNull(),
					Effect:     types.StringValue("allow"),
				})
			},
			expectedPaths: []string{"policies[0].actions"},
		},
		"null policy action": {
			mutate: func(model *RoleModel) {
				model.Policies = rolePoliciesWith(RolePolicyValue{
					Actions:    NewTypedList([]types.String{types.StringValue("read"), types.StringNull()}),
					Constraint: jsontypes.NewNormalizedNull(),
					Effect:     types.StringValue("allow"),
				})
			},
			expectedPaths: []string{"policies[0].actions[1]"},
		},
		"unknown policy action": {
			mutate: func(model *RoleModel) {
				model.Policies = rolePoliciesWith(RolePolicyValue{
					Actions:    NewTypedList([]types.String{types.StringValue("read"), types.StringUnknown()}),
					Constraint: jsontypes.NewNormalizedNull(),
					Effect:     types.StringValue("allow"),
				})
			},
			expectedPaths: []string{"policies[0].actions[1]"},
		},
		"unknown policy constraint": {
			mutate: func(model *RoleModel) {
				model.Policies = rolePoliciesWith(RolePolicyValue{
					Actions:    NewTypedList([]types.String{types.StringValue("read")}),
					Constraint: jsontypes.NewNormalizedUnknown(),
					Effect:     types.StringValue("allow"),
				})
			},
			expectedPaths: []string{"policies[0].constraint"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := validRoleRequestModel()
			test.mutate(&model)

			request, diags := model.ToRoleData()

			assert.Equal(t, cm.RoleData{}, request)
			require.True(t, diags.HasError())
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
		})
	}
}

func TestRoleRequestAggregatesParentDiagnostics(t *testing.T) {
	t.Parallel()

	model := validRoleRequestModel()
	model.Name = types.StringUnknown()
	model.Description = types.StringUnknown()
	model.Permissions = NewTypedMapUnknown[TypedList[types.String]]()
	model.Policies = NewTypedListUnknown[TypedObject[RolePolicyValue]]()

	request, diags := model.ToRoleData()

	assert.Equal(t, cm.RoleData{}, request)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{"name", "description", "permissions", "policies"}, attributeDiagnosticPaths(t, diags))
}

func TestRoleRequestDescriptionStates(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		description types.String
		expected    cm.OptNilString
	}{
		"null": {
			description: types.StringNull(),
			expected:    cm.NewOptNilStringNull(),
		},
		"empty": {
			description: types.StringValue(""),
			expected:    cm.NewOptNilString(""),
		},
		"non-empty": {
			description: types.StringValue("description"),
			expected:    cm.NewOptNilString("description"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := validRoleRequestModel()
			model.Description = test.description

			request, diags := model.ToRoleData()

			require.False(t, diags.HasError(), diags.Errors())
			assert.Equal(t, test.expected, request.Description)
		})
	}
}

func TestRoleRequestOmitsNullPolicyConstraint(t *testing.T) {
	t.Parallel()

	model := validRoleRequestModel()
	request, diags := model.ToRoleData()

	require.False(t, diags.HasError(), diags.Errors())
	require.Len(t, request.Policies, 1)
	assert.Nil(t, request.Policies[0].Constraint)
}

func validRoleRequestModel() RoleModel {
	return RoleModel{
		Name: types.StringValue("role"),
		Permissions: NewTypedMap(map[string]TypedList[types.String]{
			"Entry": NewTypedList([]types.String{types.StringValue("read")}),
		}),
		Policies: rolePoliciesWith(RolePolicyValue{
			Actions:    NewTypedList([]types.String{types.StringValue("read")}),
			Constraint: jsontypes.NewNormalizedNull(),
			Effect:     types.StringValue("allow"),
		}),
	}
}

func rolePoliciesWith(policy RolePolicyValue) TypedList[TypedObject[RolePolicyValue]] {
	return NewTypedList([]TypedObject[RolePolicyValue]{NewTypedObject(policy)})
}
