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

func newEnvironmentDataSourceItem(entity cm.Environment, spaceID string) (EnvironmentDataSourceItemModel, diag.Diagnostics) {
	var item EnvironmentDataSourceItemModel

	if diagnostics := validateDiscoverySpace(entity.Sys.Space.Sys.ID, spaceID); diagnostics.HasError() {
		return item, diagnostics
	}

	if target, ok := entity.Sys.AliasedEnvironment.Get(); ok {
		item.AliasedEnvironmentID = types.StringValue(target.Sys.ID)
	}

	item.EnvironmentID = types.StringValue(entity.Sys.ID)
	item.Name = types.StringValue(entity.Name)
	item.Status = types.StringValue(entity.Sys.Status.Sys.ID)

	return item, nil
}

//nolint:dupl // Concrete success response types enforce the collection element type at each fetch boundary.
func readEnvironments(ctx context.Context, client *cm.Client, spaceID string) ([]EnvironmentDataSourceItemModel, diag.Diagnostics) {
	const errorTitle = "Failed to read environments"

	items, diagnostics := readContentfulCollection(ctx, errorTitle,
		func(ctx context.Context, skip int64) (contentfulCollection[cm.Environment], diag.Diagnostics) {
			response, err := client.GetEnvironments(ctx, cm.GetEnvironmentsParams{SpaceID: spaceID, Skip: cm.NewOptInt64(skip), Limit: cm.NewOptInt64(defaultPageLimit)})
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, util.ErrorDetailFromContentfulManagementResponse(response, err))}
			}

			switch response := response.(type) {
			case *cm.EnvironmentCollection:
				return response, nil
			default:
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, contentfulListNonCollectionResponseDetail(response))}
			}
		},
		func(entity cm.Environment) (EnvironmentDataSourceItemModel, diag.Diagnostics) {
			return newEnvironmentDataSourceItem(entity, spaceID)
		},
	)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	slices.SortFunc(items, func(a, b EnvironmentDataSourceItemModel) int {
		return cmp.Compare(a.EnvironmentID.ValueString(), b.EnvironmentID.ValueString())
	})

	return items, nil
}
