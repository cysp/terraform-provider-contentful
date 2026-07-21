package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
			if resp.Identity != nil {
				resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, attrPath, attrValues[i])...)
			}

			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, attrPath, attrValues[i])...)
		}

		return
	}

	if req.Identity == nil {
		resp.Diagnostics.AddError("Resource Import Passthrough Multipart ID Mismatch", "No import identity was provided.")

		return
	}

	type identityComponent struct {
		path  path.Path
		value types.String
	}

	components := make([]identityComponent, 0, len(attrPaths))
	for _, attrPath := range attrPaths {
		var identityComponentValue types.String

		getDiags := req.Identity.GetAttribute(ctx, attrPath, &identityComponentValue)
		resp.Diagnostics.Append(getDiags...)

		if getDiags.HasError() {
			return
		}

		if identityComponentValue.IsUnknown() || identityComponentValue.IsNull() {
			resp.Diagnostics.AddAttributeError(attrPath, "Resource Import Passthrough Multipart ID Mismatch", "")

			return
		}

		components = append(components, identityComponent{path: attrPath, value: identityComponentValue})
	}

	for _, component := range components {
		if resp.Identity != nil {
			resp.Diagnostics.Append(resp.Identity.SetAttribute(ctx, component.path, component.value)...)
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, component.path, component.value)...)
	}
}
