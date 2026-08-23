package provider

import (
	"context"
	"fmt"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
)

const defaultPageLimit int64 = 100

type contentfulCollection[Item any] interface {
	GetTotal() cm.OptInt
	GetItems() []Item
}

func paginateContentfulCollectionItemsAsListResults[Item, Response any](
	ctx context.Context,
	req list.ListRequest,
	errorTitle string,
	fetchPage func(context.Context, int64, int64) (Response, error),
	buildResult func(Item) list.ListResult,
) func(func(list.ListResult) bool) {
	return func(yield func(list.ListResult) bool) {
		var (
			emitted int64
			skip    int64
		)

		for {
			limit := defaultPageLimit

			if req.Limit > 0 {
				remaining := req.Limit - emitted
				if remaining <= 0 {
					return
				}

				limit = min(limit, remaining)
			}

			response, err := fetchPage(ctx, skip, limit)
			if err != nil {
				yield(list.ListResult{
					Diagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic(errorTitle, util.ErrorDetailFromContentfulManagementResponse(response, err)),
					},
				})

				return
			}

			collection, ok := any(response).(contentfulCollection[Item])
			if !ok {
				yield(list.ListResult{
					Diagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic(errorTitle, contentfulListNonCollectionResponseDetail(response)),
					},
				})

				return
			}

			items := collection.GetItems()
			for _, item := range items {
				if !yield(buildResult(item)) {
					return
				}

				emitted++
			}

			itemCount := int64(len(items))
			if itemCount == 0 {
				return
			}

			skip += itemCount
			if total, ok := collection.GetTotal().Get(); ok && skip >= int64(total) {
				return
			}
		}
	}
}

func contentfulListNonCollectionResponseDetail(response any) string {
	if detail, ok := contentfulListErrorResponseDetail(response); ok {
		return detail
	}

	return fmt.Sprintf("Unexpected response type %T while listing Contentful resources.", response)
}

func contentfulListErrorResponseDetail(response any) (string, bool) {
	switch response.(type) {
	case cm.ErrorStatusCodeResponse, cm.ErrorResponse:
		return util.ErrorDetailFromContentfulManagementResponse(response, nil), true
	default:
		return "", false
	}
}
