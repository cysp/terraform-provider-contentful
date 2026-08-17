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
	publicationDiags := staged.Identity.Set(ctx, identity)

	if !publicationDiags.HasError() && req.IncludeResource {
		publicationDiags.Append(staged.Resource.Set(ctx, model)...)
	}

	if publicationDiags.HasError() {
		result.Diagnostics.Append(publicationDiags...)

		return result
	}

	staged.Diagnostics.Append(result.Diagnostics...)
	staged.Diagnostics.Append(publicationDiags...)

	return staged
}
