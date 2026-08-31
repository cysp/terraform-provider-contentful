package provider_test

import (
	"encoding/json"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
			"ContentDelivery": cm.NewStringRolePermissionsItem("all"),
			"ContentModel":    cm.NewStringArrayRolePermissionsItem([]string{"read"}),
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
	assert.Equal(t, cm.NewStringArrayRoleDataPermissionsItem([]string{"read"}), req.Permissions["ContentModel"])

	assert.Len(t, req.Policies, 2)
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
		"null permission values": {
			mutate: func(model *RoleModel) {
				model.Permissions = NewTypedMap(map[string]TypedList[types.String]{
					"ContentModel": NewTypedListNull[types.String](),
				})
			},
			expectedPaths: []string{`permissions["ContentModel"]`},
		},
		"unknown permission values": {
			mutate: func(model *RoleModel) {
				model.Permissions = NewTypedMap(map[string]TypedList[types.String]{
					"ContentModel": NewTypedListUnknown[types.String](),
				})
			},
			expectedPaths: []string{`permissions["ContentModel"]`},
		},
		"null permission value": {
			mutate: func(model *RoleModel) {
				model.Permissions = NewTypedMap(map[string]TypedList[types.String]{
					"ContentModel": NewTypedList([]types.String{types.StringValue("read"), types.StringNull()}),
				})
			},
			expectedPaths: []string{`permissions["ContentModel"][1]`},
		},
		"unknown permission value": {
			mutate: func(model *RoleModel) {
				model.Permissions = NewTypedMap(map[string]TypedList[types.String]{
					"ContentModel": NewTypedList([]types.String{types.StringValue("read"), types.StringUnknown()}),
				})
			},
			expectedPaths: []string{`permissions["ContentModel"][1]`},
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

func TestRoleRequestRejectsAllCombinedWithOtherValues(t *testing.T) {
	t.Parallel()

	for name, values := range map[string][]string{
		"all and another string": {"all", "read"},
		"duplicate all":          {"all", "all"},
	} {
		t.Run(name+"/permission values", func(t *testing.T) {
			t.Parallel()

			_, diags := ToRoleDataPermissionsItem(
				path.Root("permissions").AtMapKey("ContentModel"),
				NewTypedListFromStringSlice(values),
			)

			require.True(t, diags.HasError())
			assert.Equal(t, []string{`permissions["ContentModel"]`}, attributeDiagnosticPaths(t, diags))
			assert.Equal(t, "Invalid permission values", diags.Errors()[0].Summary())
			assert.Equal(t, `"all" must be specified by itself. Remove "all" or the other permission values from this list.`, diags.Errors()[0].Detail())
		})

		t.Run(name+"/policy actions", func(t *testing.T) {
			t.Parallel()

			_, diags := ToRoleDataPoliciesItemActions(
				path.Root("policies").AtListIndex(0).AtName("actions"),
				NewTypedListFromStringSlice(values),
			)

			require.True(t, diags.HasError())
			assert.Equal(t, []string{"policies[0].actions"}, attributeDiagnosticPaths(t, diags))
			assert.Equal(t, "Invalid policy actions", diags.Errors()[0].Summary())
			assert.Equal(t, `"all" must be specified by itself. Remove "all" or the other policy actions from this list.`, diags.Errors()[0].Detail())
		})
	}
}

func TestRoleRequestEncodesPermissionValuesAndPolicyActions(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		permissionValues []string
		policyActions    []string
		expectedValues   string
		expectedActions  string
	}{
		"empty lists": {
			permissionValues: []string{},
			policyActions:    []string{},
			expectedValues:   `[]`,
			expectedActions:  `[]`,
		},
		"documented values": {
			permissionValues: []string{"read"},
			policyActions:    []string{"read", "create"},
			expectedValues:   `["read"]`,
			expectedActions:  `["read","create"]`,
		},
		"unrecognized and duplicate strings": {
			permissionValues: []string{"future-permission-value", "read", "read"},
			policyActions:    []string{"future-policy-action", "read", "read"},
			expectedValues:   `["future-permission-value","read","read"]`,
			expectedActions:  `["future-policy-action","read","read"]`,
		},
		"singleton all": {
			permissionValues: []string{"all"},
			policyActions:    []string{"all"},
			expectedValues:   `"all"`,
			expectedActions:  `"all"`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			permissionValues, permissionDiags := ToRoleDataPermissionsItem(
				path.Root("permissions").AtMapKey("ContentModel"),
				NewTypedListFromStringSlice(test.permissionValues),
			)
			policyActions, policyDiags := ToRoleDataPoliciesItemActions(
				path.Root("policies").AtListIndex(0).AtName("actions"),
				NewTypedListFromStringSlice(test.policyActions),
			)

			require.Empty(t, permissionDiags)
			require.Empty(t, policyDiags)

			encodedValues, err := json.Marshal(permissionValues)
			require.NoError(t, err)
			assert.JSONEq(t, test.expectedValues, string(encodedValues))

			encodedActions, err := json.Marshal(policyActions)
			require.NoError(t, err)
			assert.JSONEq(t, test.expectedActions, string(encodedActions))
		})
	}
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
			"ContentModel": NewTypedList([]types.String{types.StringValue("read")}),
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
