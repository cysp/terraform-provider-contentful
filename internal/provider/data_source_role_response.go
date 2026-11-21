package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"

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

	items := make([]RoleDataSourceItemModel, 0)
	diagnostics := diag.Diagnostics{}
	pagination := rolePagination{
		params:  cm.GetRolesParams{SpaceID: spaceID, Limit: cm.NewOptInt64(defaultPageLimit)},
		cursors: make(map[string]struct{}),
	}

	for {
		err := ctx.Err()
		if err != nil {
			diagnostics.AddError(errorTitle, err.Error())

			return nil, diagnostics
		}

		response, err := client.GetRoles(ctx, pagination.params)
		if err != nil {
			diagnostics.AddError(errorTitle, util.ErrorDetailFromContentfulManagementResponse(response, err))

			return nil, diagnostics
		}

		collection, ok := response.(*cm.RoleCollection)
		if !ok {
			diagnostics.AddError(errorTitle, contentfulListNonCollectionResponseDetail(response))

			return nil, diagnostics
		}

		for _, entity := range collection.Items {
			item, itemDiagnostics := newRoleCollectionItem(ctx, entity, spaceID, len(items))
			diagnostics.Append(itemDiagnostics...)

			if diagnostics.HasError() {
				return nil, diagnostics
			}

			items = append(items, item)
		}

		more, err := pagination.advance(collection, len(items))
		if err != nil {
			diagnostics.AddError(errorTitle, err.Error())

			return nil, diagnostics
		}

		if !more {
			return items, diagnostics
		}
	}
}

type rolePagination struct {
	params  cm.GetRolesParams
	cursors map[string]struct{}
	// Keep only the last page while the response could be ignoring skip.
	previous        []cm.Role
	repeated        bool
	confirmedOffset bool
}

var (
	errRolePaginationMixed = errors.New("response mixes offset and cursor pagination")
	errRoleCursorRepeated  = errors.New("response repeats a pageNext cursor")
	errRoleOffsetRepeated  = errors.New("the same Role page repeated at three consecutive offsets without pagination metadata, so the provider cannot confirm that traversal advanced")
)

func (p *rolePagination) advance(collection *cm.RoleCollection, count int) (bool, error) {
	offset := collection.Total.IsSet() || collection.Skip.IsSet()
	cursorPage := collection.Pages.IsSet() || p.params.PageNext.IsSet()

	if cursorPage && (offset || p.params.Skip.IsSet()) {
		return false, errRolePaginationMixed
	}

	if cursorPage {
		next, ok := collection.Pages.Value.Next.Get()
		if !ok {
			return false, nil
		}

		cursor, err := rolePageNext(next, p.params.SpaceID)
		if err != nil {
			return false, err
		}

		if _, exists := p.cursors[cursor]; exists {
			return false, errRoleCursorRepeated
		}

		p.cursors[cursor] = struct{}{}
		p.params.PageNext = cm.NewOptString(cursor)

		return true, nil
	}

	if len(collection.Items) == 0 || (collection.Total.IsSet() && count >= collection.Total.Value) {
		return false, nil
	}

	if offset {
		p.confirmedOffset = true
		p.previous = nil
	} else if !p.confirmedOffset {
		// One duplicate can precede a genuine empty offset sentinel.
		repeats := reflect.DeepEqual(p.previous, collection.Items)
		if repeats && p.repeated {
			return false, errRoleOffsetRepeated
		}

		p.previous, p.repeated = collection.Items, repeats
	}

	p.params.Skip = cm.NewOptInt64(int64(count))

	return true, nil
}

func newRoleCollectionItem(ctx context.Context, entity cm.Role, spaceID string, index int) (RoleDataSourceItemModel, diag.Diagnostics) {
	if entity.Sys.Space.Sys.ID != spaceID {
		return RoleDataSourceItemModel{}, discoveryResponseIdentityError(fmt.Sprintf("Role %q has a Space link that differs from the requested space %q.", entity.Sys.ID, spaceID))
	}

	return newRoleDataSourceItem(ctx, entity, path.Root("roles").AtListIndex(index))
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
