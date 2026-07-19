package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
)

func newListResultFromResponse[T any](
	ctx context.Context,
	req list.ListRequest,
	displayName string,
	identity any,
	convert func() (T, diag.Diagnostics),
) list.ListResult {
	result := req.NewListResult(ctx)
	result.DisplayName = displayName
	result.Diagnostics.Append(result.Identity.Set(ctx, identity)...)

	if result.Diagnostics.HasError() || !req.IncludeResource {
		return result
	}

	model, modelDiags := convert()
	result.Diagnostics.Append(modelDiags...)

	if result.Diagnostics.HasError() {
		return result
	}

	result.Diagnostics.Append(result.Resource.Set(ctx, model)...)

	return result
}
