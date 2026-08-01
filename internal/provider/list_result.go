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

	var model T
	if req.IncludeResource {
		var modelDiags diag.Diagnostics
		model, modelDiags = convert()
		result.Diagnostics.Append(modelDiags...)

		if result.Diagnostics.HasError() {
			return result
		}
	}

	staged := req.NewListResult(ctx)
	staged.DisplayName = displayName
	staged.Diagnostics.Append(staged.Identity.Set(ctx, identity)...)

	if !staged.Diagnostics.HasError() && req.IncludeResource {
		staged.Diagnostics.Append(staged.Resource.Set(ctx, model)...)
	}

	if staged.Diagnostics.HasError() {
		result.Diagnostics.Append(staged.Diagnostics...)

		return result
	}

	return staged
}
