package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewRoleFromFields(spaceID, roleID string, roleFields cm.RoleData) cm.Role {
	role := cm.Role{
		Sys: NewRoleSys(spaceID, roleID),
	}

	UpdateRoleFromFields(&role, roleFields)

	return role
}

func NewRoleSys(spaceID, roleID string) cm.RoleSys {
	return cm.RoleSys{
		Type: cm.RoleSysTypeRole,
		Space: cm.SpaceLink{
			Sys: cm.SpaceLinkSys{
				Type:     cm.SpaceLinkSysTypeLink,
				LinkType: cm.SpaceLinkSysLinkTypeSpace,
				ID:       spaceID,
			},
		},
		ID: roleID,
	}
}

func UpdateRoleFromFields(role *cm.Role, roleFields cm.RoleData) {
	role.Name = roleFields.Name
	role.Permissions = convertMap(roleFields.Permissions, func(permission cm.RoleDataPermissionsItem) cm.RolePermissionsItem {
		return cm.RolePermissionsItem{
			Type:        cm.RolePermissionsItemType(permission.Type),
			String:      permission.String,
			StringArray: permission.StringArray,
		}
	})
	role.Policies = convertSlice(roleFields.Policies, func(policy cm.RoleDataPoliciesItem) cm.RolePoliciesItem {
		return cm.RolePoliciesItem{
			Effect: cm.RolePoliciesItemEffect(policy.Effect),
			Actions: cm.RolePoliciesItemActions{
				Type:        cm.RolePoliciesItemActionsType(policy.Actions.Type),
				String:      policy.Actions.String,
				StringArray: policy.Actions.StringArray,
			},
			Constraint: policy.Constraint,
		}
	})
}
