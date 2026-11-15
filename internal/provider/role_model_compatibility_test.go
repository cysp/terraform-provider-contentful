package provider_test

import (
	"maps"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These literal wire types and values must remain independent of the Go model
// representation. The same fixture can be decoded from state written before a
// model refactor without requiring a Terraform state migration.
func roleCompatibilityPolicyType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"actions":    tftypes.List{ElementType: tftypes.String},
		"constraint": tftypes.String,
		"effect":     tftypes.String,
	}}
}

func roleCompatibilityItemType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"role_id":     tftypes.String,
		"name":        tftypes.String,
		"description": tftypes.String,
		"permissions": tftypes.Map{ElementType: tftypes.List{ElementType: tftypes.String}},
		"policies":    tftypes.List{ElementType: roleCompatibilityPolicyType()},
	}}
}

func roleCompatibilityResourceType() tftypes.Object {
	attributes := roleCompatibilityItemType().AttributeTypes
	attributes["id"] = tftypes.String
	attributes["space_id"] = tftypes.String
	attributes["timeouts"] = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"create": tftypes.String, "read": tftypes.String,
		"update": tftypes.String, "delete": tftypes.String,
	}}

	return tftypes.Object{AttributeTypes: attributes}
}

func roleCompatibilityItem(permissions, policies any) map[string]tftypes.Value {
	types := roleCompatibilityItemType().AttributeTypes

	return map[string]tftypes.Value{
		"role_id":     tftypes.NewValue(tftypes.String, "role"),
		"name":        tftypes.NewValue(tftypes.String, "Reader"),
		"description": tftypes.NewValue(tftypes.String, nil),
		"permissions": tftypes.NewValue(types["permissions"], permissions),
		"policies":    tftypes.NewValue(types["policies"], policies),
	}
}

func roleCompatibilityPolicy(actions, constraint, effect any) tftypes.Value {
	return tftypes.NewValue(roleCompatibilityPolicyType(), map[string]tftypes.Value{
		"actions":    tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, actions),
		"constraint": tftypes.NewValue(tftypes.String, constraint),
		"effect":     tftypes.NewValue(tftypes.String, effect),
	})
}

func TestRoleModelsPreserveWireValues(t *testing.T) {
	t.Parallel()

	stringList := tftypes.List{ElementType: tftypes.String}
	policyType := roleCompatibilityPolicyType()
	cases := map[string]struct {
		permissions, policies any
	}{
		"null containers": {},
		"empty containers": {
			permissions: map[string]tftypes.Value{},
			policies:    []tftypes.Value{},
		},
		"duplicates empty actions and exact JSON": {
			permissions: map[string]tftypes.Value{
				"ContentModel": tftypes.NewValue(stringList, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "read"), tftypes.NewValue(tftypes.String, "read"),
				}),
				"Disabled": tftypes.NewValue(stringList, []tftypes.Value{}),
			},
			policies: []tftypes.Value{
				roleCompatibilityPolicy([]tftypes.Value{}, nil, "deny"),
				roleCompatibilityPolicy([]tftypes.Value{
					tftypes.NewValue(tftypes.String, "read"), tftypes.NewValue(tftypes.String, "read"),
				}, `{ "z": 1, "a": [1, 2] }`, "allow"),
			},
		},
		"null nested containers objects and scalars": {
			permissions: map[string]tftypes.Value{
				"ContentModel": tftypes.NewValue(stringList, nil),
				"ContentDelivery": tftypes.NewValue(stringList, []tftypes.Value{
					tftypes.NewValue(tftypes.String, nil),
				}),
			},
			policies: []tftypes.Value{
				tftypes.NewValue(policyType, nil),
				roleCompatibilityPolicy(nil, nil, nil),
				roleCompatibilityPolicy([]tftypes.Value{tftypes.NewValue(tftypes.String, nil)}, nil, "allow"),
			},
		},
		"unknown scalar leaves": {
			permissions: map[string]tftypes.Value{
				"ContentModel": tftypes.NewValue(stringList, []tftypes.Value{
					tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				}),
			},
			policies: []tftypes.Value{roleCompatibilityPolicy(
				[]tftypes.Value{tftypes.NewValue(tftypes.String, tftypes.UnknownValue)},
				tftypes.UnknownValue, tftypes.UnknownValue,
			)},
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			item := roleCompatibilityItem(test.permissions, test.policies)
			readTimeoutType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"read": tftypes.String}}
			singularType := roleCompatibilityItemType()
			singularType.AttributeTypes["id"] = tftypes.String
			singularType.AttributeTypes["space_id"] = tftypes.String
			singularType.AttributeTypes["timeouts"] = readTimeoutType
			pluralType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"id": tftypes.String, "space_id": tftypes.String,
				"roles": tftypes.List{ElementType: roleCompatibilityItemType()}, "timeouts": readTimeoutType,
			}}
			resourceValues := maps.Clone(item)
			resourceValues["id"] = tftypes.NewValue(tftypes.String, "space/role")
			resourceValues["space_id"] = tftypes.NewValue(tftypes.String, "space")
			resourceValues["timeouts"] = tftypes.NewValue(roleCompatibilityResourceType().AttributeTypes["timeouts"], nil)
			singularValues := maps.Clone(resourceValues)
			singularValues["timeouts"] = tftypes.NewValue(readTimeoutType, nil)

			for _, model := range []struct {
				name   string
				state  tfsdk.State
				target any
			}{
				{
					name:   "resource",
					state:  tfsdk.State{Schema: RoleResourceSchema(t.Context()), Raw: tftypes.NewValue(roleCompatibilityResourceType(), resourceValues)},
					target: &RoleModel{},
				},
				{
					name:   "singular data source",
					state:  tfsdk.State{Schema: RoleDataSourceSchema(t.Context()), Raw: tftypes.NewValue(singularType, singularValues)},
					target: &RoleDataSourceModel{},
				},
				{
					name: "plural data source",
					state: tfsdk.State{Schema: RolesDataSourceSchema(t.Context()), Raw: tftypes.NewValue(pluralType, map[string]tftypes.Value{
						"id":       tftypes.NewValue(tftypes.String, "space"),
						"space_id": tftypes.NewValue(tftypes.String, "space"),
						"roles":    tftypes.NewValue(pluralType.AttributeTypes["roles"], []tftypes.Value{tftypes.NewValue(roleCompatibilityItemType(), item)}),
						"timeouts": tftypes.NewValue(readTimeoutType, nil),
					})},
					target: &RolesDataSourceModel{},
				},
			} {
				t.Run(model.name, func(t *testing.T) {
					t.Parallel()

					expected := model.state.Raw.Copy()
					assert.True(t, expected.Type().Equal(model.state.Schema.Type().TerraformType(t.Context())), "public Terraform type changed")
					require.Empty(t, model.state.Get(t.Context(), model.target))
					require.Empty(t, model.state.Set(t.Context(), model.target))
					assert.True(t, expected.Equal(model.state.Raw), "wire values changed during model round trip")
				})
			}
		})
	}
}
