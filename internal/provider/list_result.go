package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
)

// newListResultFromResponse stages identity and resource encodings before
// publishing them. Warnings accompany a published result; conversion or
// publication errors leave identity and resource unpublished.
func newListResultFromResponse[T any](
	ctx context.Context,
	req list.ListRequest,
	displayName string,
	identity any,
	projectResource func() (T, diag.Diagnostics),
) list.ListResult {
	result := req.NewListResult(ctx)
	result.DisplayName = displayName

	var resourceModel T

	if req.IncludeResource {
		var resourceModelDiags diag.Diagnostics

		resourceModel, resourceModelDiags = projectResource()
		result.Diagnostics.Append(resourceModelDiags...)

		if result.Diagnostics.HasError() {
			return result
		}
	}

	stagedResult := req.NewListResult(ctx)
	stagedResult.DisplayName = displayName
	encodingDiags := stagedResult.Identity.Set(ctx, identity)

	if !encodingDiags.HasError() && req.IncludeResource {
		encodingDiags.Append(stagedResult.Resource.Set(ctx, resourceModel)...)
	}

	if encodingDiags.HasError() {
		result.Diagnostics.Append(encodingDiags...)

		return result
	}

	stagedResult.Diagnostics.Append(result.Diagnostics...)
	stagedResult.Diagnostics.Append(encodingDiags...)

	return stagedResult
}
