package util

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ImportStatePassthroughMultipartID(ctx context.Context, attrPaths []path.Path, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
}
