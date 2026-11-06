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
	project func(index int, item Item) (Model, diag.Diagnostics),
) ([]Model, diag.Diagnostics) {
	items := make([]Model, 0)

	var diagnostics diag.Diagnostics

	var skip int64

	for {
		err := ctx.Err()
		if err != nil {
			diagnostics.AddError(errorTitle, err.Error())

			return nil, diagnostics
		}

		collection, pageDiagnostics := fetchPage(ctx, skip)
		diagnostics.Append(pageDiagnostics...)

		if diagnostics.HasError() {
			return nil, diagnostics
		}

		page := collection.GetItems()
		for _, entity := range page {
			item, itemDiagnostics := project(len(items), entity)
			diagnostics.Append(itemDiagnostics...)

			if diagnostics.HasError() {
				return nil, diagnostics
			}

			items = append(items, item)
		}

		skip += int64(len(page))

		total, totalSet := collection.GetTotal().Get()
		if len(page) == 0 || (totalSet && skip >= int64(total)) {
			return items, diagnostics
		}
	}
}
