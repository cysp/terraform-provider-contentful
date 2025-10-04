package provider

import (
	"context"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewRoleResourceModelFromResponse(ctx context.Context, role cm.Role) (RoleModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := role.Sys.Space.Sys.ID
	roleID := role.Sys.ID

	model := RoleModel{
		IDIdentityModel: IDIdentityModel{
			ID: types.StringValue(strings.Join([]string{spaceID, roleID}, "/")),
		},
		RoleIdentityModel: RoleIdentityModel{
			SpaceID: types.StringValue(spaceID),
			RoleID:  types.StringValue(roleID),
		},
	}

	model.Name = types.StringValue(role.Name)
	model.Description = types.StringPointerValue(role.Description.ValueStringPointer())

	permissionsMapValue, permissionsMapValueDiags := NewPermissionsMapValueFromResponse(ctx, path.Root("permissions"), role.Permissions)
	diags.Append(permissionsMapValueDiags...)

	model.Permissions = permissionsMapValue

	policiesListValue, policiesListValueDiags := NewPoliciesListValueFromResponse(ctx, path.Root("policies"), role.Policies)
	diags.Append(policiesListValueDiags...)

	model.Policies = policiesListValue

	return model, diags
}
