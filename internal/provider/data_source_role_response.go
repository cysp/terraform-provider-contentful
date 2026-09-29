package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"

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

	roles := make([]cm.Role, 0)
	params := cm.GetRolesParams{SpaceID: spaceID, Limit: cm.NewOptInt64(defaultPageLimit)}
	seen := make(map[string]struct{})

	var skip int64

	for {
		err := ctx.Err()
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, err.Error())}
		}

		response, err := client.GetRoles(ctx, params)
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, util.ErrorDetailFromContentfulManagementResponse(response, err))}
		}

		collection, ok := response.(*cm.RoleCollection)
		if !ok {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, contentfulListNonCollectionResponseDetail(response))}
		}

		for _, entity := range collection.Items {
			if entity.Sys.Space.Sys.ID != spaceID {
				return nil, discoveryResponseIdentityError(fmt.Sprintf("Role %q has a Space link that differs from the requested space %q.", entity.Sys.ID, spaceID))
			}

			roles = append(roles, entity)
		}

		// An offset request retains legacy traversal when later metadata is omitted.
		// On the initial request, offset metadata distinguishes legacy responses
		// from terminal cursor pages, where the SDK permits pages to be absent.
		offset := params.Skip.IsSet() || collection.Total.IsSet() || collection.Skip.IsSet()
		if offset && (collection.Pages.IsSet() || params.PageNext.IsSet()) {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, "Role collection mixes offset and cursor pagination.")}
		}

		if offset {
			skip += int64(len(collection.Items))
			if len(collection.Items) == 0 || (collection.Total.IsSet() && skip >= int64(collection.Total.Value)) {
				break
			}

			params.Skip = cm.NewOptInt64(skip)

			continue
		}

		next := collection.Pages.Value.Next
		if !next.IsSet() {
			break
		}

		cursor, err := rolePageNext(next.Value, spaceID)
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, err.Error())}
		}

		if _, exists := seen[cursor]; exists {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, "Role collection repeats a pageNext cursor.")}
		}

		seen[cursor] = struct{}{}

		params.Skip.Reset()
		params.PageNext = cm.NewOptString(cursor)
	}

	// Project only the complete collection, preserving final diagnostic indexes.
	return projectRoles(ctx, roles)
}

var errRolePageNext = errors.New("invalid Role pages.next")

// Extract only the cursor. Never request a response-provided URL or copy its
// authority, credentials, or other query parameters into the configured client.
func rolePageNext(link, spaceID string) (string, error) {
	navigation, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("%w URL", errRolePageNext)
	}

	if navigation.User != nil || navigation.Fragment != "" || navigation.Opaque != "" ||
		(navigation.Scheme != "" && navigation.Host == "") || (navigation.Host != "" && navigation.Scheme == "") ||
		(navigation.Scheme != "" && navigation.Scheme != "https" && navigation.Scheme != "http") ||
		navigation.Path != "/spaces/"+spaceID+"/roles" {
		return "", fmt.Errorf("%w endpoint", errRolePageNext)
	}

	query, err := url.ParseQuery(navigation.RawQuery)
	if err != nil {
		return "", fmt.Errorf("%w query", errRolePageNext)
	}

	cursors := query["pageNext"]
	if len(cursors) != 1 || cursors[0] == "" || query.Has("pagePrev") || query.Has("skip") {
		return "", fmt.Errorf("%w: expected one nonempty pageNext cursor and no pagePrev or skip", errRolePageNext)
	}

	return cursors[0], nil
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
