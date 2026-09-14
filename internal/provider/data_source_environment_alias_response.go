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

func newEnvironmentAliasDataSourceItem(entity cm.EnvironmentAlias, spaceID string) (EnvironmentAliasDataSourceItemModel, diag.Diagnostics) {
	var item EnvironmentAliasDataSourceItemModel

	if diagnostics := validateDiscoverySpace(entity.Sys.Space.Sys.ID, spaceID); diagnostics.HasError() {
		return item, diagnostics
	}

	item.EnvironmentAliasID = types.StringValue(entity.Sys.ID)
	item.TargetEnvironmentID = types.StringValue(entity.Environment.Sys.ID)

	return item, nil
}

//nolint:dupl // Concrete success response types enforce the collection element type at each fetch boundary.
func readEnvironmentAliases(ctx context.Context, client *cm.Client, spaceID string) ([]EnvironmentAliasDataSourceItemModel, diag.Diagnostics) {
	const errorTitle = "Failed to read environment aliases"

	items, diagnostics := readContentfulCollection(ctx, errorTitle,
		func(ctx context.Context, skip int64) (contentfulCollection[cm.EnvironmentAlias], diag.Diagnostics) {
			response, err := client.GetEnvironmentAliases(ctx, cm.GetEnvironmentAliasesParams{SpaceID: spaceID, Skip: cm.NewOptInt64(skip), Limit: cm.NewOptInt64(defaultPageLimit)})
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, util.ErrorDetailFromContentfulManagementResponse(response, err))}
			}

			switch response := response.(type) {
			case *cm.EnvironmentAliasCollection:
				return response, nil
			default:
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, contentfulListNonCollectionResponseDetail(response))}
			}
		},
		func(entity cm.EnvironmentAlias) (EnvironmentAliasDataSourceItemModel, diag.Diagnostics) {
			return newEnvironmentAliasDataSourceItem(entity, spaceID)
		},
	)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	slices.SortFunc(items, func(a, b EnvironmentAliasDataSourceItemModel) int {
		return cmp.Compare(a.EnvironmentAliasID.ValueString(), b.EnvironmentAliasID.ValueString())
	})

	return items, nil
}
