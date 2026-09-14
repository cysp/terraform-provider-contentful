package provider

import (
	"context"
	"fmt"
	"iter"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
)

func paginateContentfulCollectionItemsAsListResults[Item any](
	ctx context.Context,
	req list.ListRequest,
	fetchPage func(ctx context.Context, skip, limit int64) (contentfulCollection[Item], diag.Diagnostics),
	buildResult func(item Item) list.ListResult,
) iter.Seq[list.ListResult] {
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

			collection, diagnostics := fetchPage(ctx, skip, limit)
			if diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: diagnostics})

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
