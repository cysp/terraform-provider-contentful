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

func newLocaleDataSourceItem(entity cm.Locale, spaceID, environmentID string) (LocaleDataSourceItemModel, diag.Diagnostics) {
	var item LocaleDataSourceItemModel

	if diagnostics := validateDiscoverySpace(entity.Sys.Space.Sys.ID, spaceID); diagnostics.HasError() {
		return item, diagnostics
	}

	if entity.Sys.Environment.Sys.ID != environmentID {
		return item, discoveryResponseIdentityError("Unsupported Locale response identity: the environment link does not echo the requested addressing context. This response cannot establish alias routing.")
	}

	item.LocaleID = types.StringValue(entity.Sys.ID)
	item.Name = types.StringValue(entity.Name)
	item.Code = types.StringValue(entity.Code)
	item.FallbackCode = types.StringPointerValue(entity.FallbackCode.ValueStringPointer())
	item.Default = types.BoolValue(entity.Default)
	item.Optional = types.BoolValue(entity.Optional)
	item.ContentManagementAPI = types.BoolValue(entity.ContentManagementApi)
	item.ContentDeliveryAPI = types.BoolValue(entity.ContentDeliveryApi)

	return item, nil
}

func readLocales(ctx context.Context, client *cm.Client, spaceID, environmentID string) ([]LocaleDataSourceItemModel, diag.Diagnostics) {
	const errorTitle = "Failed to read locales"

	items, diagnostics := readContentfulCollection(ctx, errorTitle,
		func(ctx context.Context, skip int64) (contentfulCollection[cm.Locale], diag.Diagnostics) {
			response, err := client.GetLocales(ctx, cm.GetLocalesParams{SpaceID: spaceID, EnvironmentID: environmentID, Skip: cm.NewOptInt64(skip), Limit: cm.NewOptInt64(defaultPageLimit)})
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, util.ErrorDetailFromContentfulManagementResponse(response, err))}
			}

			switch response := response.(type) {
			case *cm.LocaleCollection:
				return response, nil
			default:
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, contentfulListNonCollectionResponseDetail(response))}
			}
		},
		func(entity cm.Locale) (LocaleDataSourceItemModel, diag.Diagnostics) {
			return newLocaleDataSourceItem(entity, spaceID, environmentID)
		},
	)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	slices.SortFunc(items, func(a, b LocaleDataSourceItemModel) int {
		return cmp.Compare(a.LocaleID.ValueString(), b.LocaleID.ValueString())
	})

	return items, nil
}
