package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const defaultPageLimit int64 = 100

type contentfulCollection[Item any] interface {
	GetTotal() cm.OptInt
	GetItems() []Item
}

func readContentfulCollection[Item, Model any](
	ctx context.Context,
	errorTitle string,
	fetchPage func(ctx context.Context, skip int64) (contentfulCollection[Item], diag.Diagnostics),
	project func(item Item) (Model, diag.Diagnostics),
) ([]Model, diag.Diagnostics) {
	items := make([]Model, 0)

	var skip int64

	for {
		err := ctx.Err()
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic(errorTitle, err.Error())}
		}

		collection, diagnostics := fetchPage(ctx, skip)
		if diagnostics.HasError() {
			return nil, diagnostics
		}

		page := collection.GetItems()
		for _, entity := range page {
			item, diagnostics := project(entity)
			if diagnostics.HasError() {
				return nil, diagnostics
			}

			items = append(items, item)
		}

		skip += int64(len(page))

		total, totalSet := collection.GetTotal().Get()
		if len(page) == 0 || (totalSet && skip >= int64(total)) {
			return items, nil
		}
	}
}
