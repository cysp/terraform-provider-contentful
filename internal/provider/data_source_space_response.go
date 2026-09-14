package provider

import (
	"cmp"
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newSpaceDataSourceItem(entity cm.Space) SpaceDataSourceItemModel {
	var item SpaceDataSourceItemModel

	item.SpaceID = types.StringValue(entity.Sys.ID)
	item.OrganizationID = types.StringValue(entity.Sys.Organization.Sys.ID)
	item.Name = types.StringValue(entity.Name)

	return item
}

func readSpaces(ctx context.Context, client *cm.Client, organizationID types.String) ([]SpaceDataSourceItemModel, diag.Diagnostics) {
	const errorTitle = "Failed to read spaces"

	items, diagnostics := readContentfulCollection(ctx, errorTitle,
		func(ctx context.Context, skip int64) (contentfulCollection[cm.Space], diag.Diagnostics) {
			response, err := client.GetSpaces(ctx, cm.GetSpacesParams{XContentfulOrganization: cm.OptString{Value: organizationID.ValueString(), Set: !organizationID.IsNull()}, Skip: cm.NewOptInt64(skip), Limit: cm.NewOptInt64(defaultPageLimit)})
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, util.ErrorDetailFromContentfulManagementResponse(response, err))}
			}

			switch response := response.(type) {
			case *cm.SpaceCollection:
				return response, nil
			default:
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, contentfulListNonCollectionResponseDetail(response))}
			}
		},
		func(entity cm.Space) (SpaceDataSourceItemModel, diag.Diagnostics) {
			if !organizationID.IsNull() && entity.Sys.Organization.Sys.ID != organizationID.ValueString() {
				return SpaceDataSourceItemModel{}, discoveryResponseIdentityError("The returned organization ID differs from the requested scope.")
			}

			return newSpaceDataSourceItem(entity), nil
		},
	)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	slices.SortFunc(items, func(a, b SpaceDataSourceItemModel) int {
		return cmp.Compare(a.SpaceID.ValueString(), b.SpaceID.ValueString())
	})

	return items, nil
}
