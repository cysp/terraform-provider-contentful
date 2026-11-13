package provider_test

import (
	"slices"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestRoleMutationReconciliationProperties(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	rapid.Check(t, func(t *rapid.T) {
		plannedJSON, returnedJSON, reversedJSON := drawReconciliationJSON(t)
		actions := drawReconciliationDuplicates(t, "read", "create")
		effects := drawReconciliationDuplicates(t, "allow", "deny")
		plannedActions := rapid.Permutation(actions).Draw(t, "planned actions")
		returnedActions := rapid.Permutation(actions).Draw(t, "returned actions")
		plannedEffects := rapid.Permutation(effects).Draw(t, "planned policies")
		returnedEffects := rapid.Permutation(effects).Draw(t, "returned policies")

		plan := completeRoleMutationPlan()
		plan.Permissions = NewTypedMap(map[string]TypedList[types.String]{"Entry": NewTypedListFromStringSlice(plannedActions)})

		policies := make([]TypedObject[RolePolicyValue], len(plannedEffects))
		for index, effect := range plannedEffects {
			policies[index] = rolePolicyValue(effect, plannedActions, jsontypes.NewNormalizedValue(plannedJSON))
		}

		plan.Policies = NewTypedList(policies)
		roleID := "role"

		if rapid.Bool().Draw(t, "generated identity") {
			plan.ID = types.StringUnknown()
			plan.RoleID = types.StringUnknown()
			roleID = rapid.StringMatching(`[a-z][a-z0-9]{0,7}`).Draw(t, "created role ID")
		}

		response := completeRoleMutationResponse()
		response.Sys.ID = roleID
		response.Permissions = cm.RolePermissions{"Entry": cm.NewStringArrayRolePermissionsItem(returnedActions)}
		response.Policies = make([]cm.RolePoliciesItem, len(returnedEffects))

		responsePolicies := make([]TypedObject[RolePolicyValue], len(returnedEffects))
		for index, effect := range returnedEffects {
			response.Policies[index] = rolePolicyResponse(effect, returnedActions, returnedJSON)
			responsePolicies[index] = rolePolicyValue(effect, returnedActions, jsontypes.NewNormalizedValue(returnedJSON))
		}

		state, responseDiags, consistencyDiags := ReconcileRoleMutationResponse(ctx, response, plan)
		require.Empty(t, responseDiags)
		requireReconciliationPaths(t, consistencyDiags)
		assert.Equal(t, plan.Permissions, state.Permissions, "equivalent action permutations must restore exact Plan order")
		assert.Equal(t, plan.Policies, state.Policies, "equivalent policy permutations must restore exact Plan JSON and order")
		assert.Equal(t, "space/"+roleID, state.ID.ValueString())
		assert.Equal(t, roleID, state.RoleID.ValueString())

		// Every failure below must suppress every otherwise equivalent overlay.
		// Expected state is built directly, including the returned policy order
		// and JSON representation, independently of response projection.
		expected := RoleModel{
			IDIdentityModel: NewIDIdentityModelFromMultipartID("space", roleID),
			SpaceID:         types.StringValue("space"), RoleID: types.StringValue(roleID),
			Name:        plan.Name,
			Description: plan.Description,
			Permissions: NewTypedMap(map[string]TypedList[types.String]{"Entry": NewTypedListFromStringSlice(returnedActions)}),
			Policies:    NewTypedList(responsePolicies),
		}

		var expectedPath string

		switch rapid.IntRange(0, 5).Draw(t, "contradiction") {
		case 0:
			changed := slices.Clone(actions)
			changed[0] = "create"
			changed = rapid.Permutation(changed).Draw(t, "changed permissions")
			response.Permissions["Entry"] = cm.NewStringArrayRolePermissionsItem(changed)
			expected.Permissions = NewTypedMap(map[string]TypedList[types.String]{"Entry": NewTypedListFromStringSlice(changed)})
			expectedPath = "permissions"
		case 1:
			// Change one duplicate policy, preserving length and the set of effects.
			changed := slices.Clone(effects)
			changed[0] = "deny"

			changed = rapid.Permutation(changed).Draw(t, "changed policies")
			for index, effect := range changed {
				response.Policies[index] = rolePolicyResponse(effect, returnedActions, returnedJSON)
				responsePolicies[index] = rolePolicyValue(effect, returnedActions, jsontypes.NewNormalizedValue(returnedJSON))
			}

			expected.Policies = NewTypedList(responsePolicies)
			expectedPath = "policies"
		case 2:
			index := rapid.IntRange(0, len(response.Policies)-1).Draw(t, "policy index")
			changed := slices.Clone(actions)
			changed[0] = "create"
			response.Policies[index].Actions = cm.NewStringArrayRolePoliciesItemActions(changed)
			responsePolicies[index] = rolePolicyValue(returnedEffects[index], changed, jsontypes.NewNormalizedValue(returnedJSON))
			expected.Policies = NewTypedList(responsePolicies)
			expectedPath = "policies"
		case 3:
			index := rapid.IntRange(0, len(response.Policies)-1).Draw(t, "policy index")
			response.Policies[index].Constraint = []byte(reversedJSON)
			responsePolicies[index] = rolePolicyValue(returnedEffects[index], returnedActions, jsontypes.NewNormalizedValue(reversedJSON))
			expected.Policies = NewTypedList(responsePolicies)
			expectedPath = "policies"
		case 4:
			response.Name += " changed"
			expected.Name = types.StringValue(response.Name)
			expectedPath = "name"
		case 5:
			response.Sys.Space.Sys.ID = "other-space"
			expectedPath = "space_id"
		}

		state, responseDiags, consistencyDiags = ReconcileRoleMutationResponse(ctx, response, plan)
		require.Empty(t, responseDiags)
		requireReconciliationPaths(t, consistencyDiags, expectedPath)
		assert.Equal(t, expected, state, "one contradiction must retain the complete response except endpoint identity")
	})
}
