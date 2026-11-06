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

	return readContentfulCollection(ctx, errorTitle,
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
		func(index int, entity cm.Role) (RoleDataSourceItemModel, diag.Diagnostics) {
			if entity.Sys.Space.Sys.ID != spaceID {
				return RoleDataSourceItemModel{}, discoveryResponseIdentityError(fmt.Sprintf("Role %q has a Space link that differs from the requested space %q.", entity.Sys.ID, spaceID))
			}

			return newRoleDataSourceItem(ctx, entity, path.Root("roles").AtListIndex(index))
		},
	)
}
