package provider

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func newRoleDataSourceItem(ctx context.Context, entity cm.Role, root path.Path) (RoleDataSourceItemModel, diag.Diagnostics) {
	model, diagnostics, _, _ := newRoleResourceModelFromResponseAtPath(ctx, entity, root)

	return RoleDataSourceItemModel{
		RoleID:      model.RoleID,
		Name:        model.Name,
		Description: model.Description,
		Permissions: model.Permissions,
		Policies:    model.Policies,
	}, diagnostics
}

func readRoles(ctx context.Context, client *cm.Client, spaceID string) ([]RoleDataSourceItemModel, diag.Diagnostics) {
	const errorTitle = "Failed to read roles"

	// Project after pagination so diagnostics use the final roles[index] paths.
	roles, diagnostics := readContentfulCollection(ctx, errorTitle,
		func(ctx context.Context, skip int64) (contentfulCollection[cm.Role], diag.Diagnostics) {
			response, err := client.GetRoles(ctx, cm.GetRolesParams{SpaceID: spaceID, Skip: cm.NewOptInt64(skip), Limit: cm.NewOptInt64(defaultPageLimit)})
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, util.ErrorDetailFromContentfulManagementResponse(response, err))}
			}

			switch response := response.(type) {
			case *cm.RoleCollection:
				return response, nil
			default:
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, contentfulListNonCollectionResponseDetail(response))}
			}
		},
		func(entity cm.Role) (cm.Role, diag.Diagnostics) {
			if entity.Sys.Space.Sys.ID != spaceID {
				return cm.Role{}, discoveryResponseIdentityError(fmt.Sprintf("Role %q has a Space link that differs from the requested space %q.", entity.Sys.ID, spaceID))
			}

			return entity, nil
		},
	)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	return projectRoles(ctx, roles)
}

func projectRoles(ctx context.Context, roles []cm.Role) ([]RoleDataSourceItemModel, diag.Diagnostics) {
	items := make([]RoleDataSourceItemModel, 0, len(roles))
	diagnostics := diag.Diagnostics{}

	for index, role := range roles {
		item, itemDiagnostics := newRoleDataSourceItem(ctx, role, path.Root("roles").AtListIndex(index))
		diagnostics.Append(itemDiagnostics...)

		if diagnostics.HasError() {
			return nil, diagnostics
		}

		items = append(items, item)
	}

	return items, diagnostics
}
