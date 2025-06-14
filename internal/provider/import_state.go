package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ImportStatePassthroughMultipartID(ctx context.Context, attrPaths []path.Path, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "" {
		attrValues := strings.Split(req.ID, "/")
		if len(attrPaths) != len(attrValues) {
			resp.Diagnostics.AddError(
				"Resource Import Passthrough Multipart ID Mismatch",
				"",
			)

			return
		}

		for i, attrPath := range attrPaths {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, attrPath, attrValues[i])...)
		}

		return
	}

	for _, attrPath := range attrPaths {
		var identityComponentValue attr.Value

		resp.Diagnostics.Append(req.Identity.GetAttribute(ctx, attrPath, &identityComponentValue)...)

		if identityComponentValue.IsUnknown() || identityComponentValue.IsNull() {
			resp.Diagnostics.AddAttributeError(attrPath, "Resource Import Passthrough Multipart ID Mismatch", "")
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, attrPath, identityComponentValue)...)
	}
}
