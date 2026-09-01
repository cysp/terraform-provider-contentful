package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type importAttribute struct {
	path  path.Path
	value types.String
}

func resourceIdentitySchema(attributeNames []string) identityschema.Schema {
	attributes := make(map[string]identityschema.Attribute, len(attributeNames))

	for _, attributeName := range attributeNames {
		attributes[attributeName] = identityschema.StringAttribute{RequiredForImport: true}
	}

	return identityschema.Schema{Attributes: attributes}
}

func resourceIdentityPaths(attributeNames []string) []path.Path {
	attributePaths := make([]path.Path, 0, len(attributeNames))

	for _, attributeName := range attributeNames {
		attributePaths = append(attributePaths, path.Root(attributeName))
	}

	return attributePaths
}

func ImportStatePassthroughMultipartID(ctx context.Context, identityAttributeNames []string, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	attrPaths := resourceIdentityPaths(identityAttributeNames)

	if req.ID != "" {
		attrValues := strings.Split(req.ID, "/")
		if len(attrPaths) != len(attrValues) {
			resp.Diagnostics.AddError(
				"Resource Import Passthrough Multipart ID Mismatch",
				"",
			)

			return
		}

		attributes := make([]importAttribute, 0, len(attrPaths))
		for i, attrPath := range attrPaths {
			if attrValues[i] == "" {
				resp.Diagnostics.AddAttributeError(attrPath, "Resource Import Passthrough Multipart ID Mismatch", "Import identity components must not be empty.")

				return
			}

			attributes = append(attributes, importAttribute{path: attrPath, value: types.StringValue(attrValues[i])})
		}

		resp.Diagnostics.Append(setImportAttributes(ctx, &resp.State, resp.Identity, attributes)...)

		return
	}

	if req.Identity == nil {
		resp.Diagnostics.AddError("Resource Import Passthrough Multipart ID Mismatch", "No import identity was provided.")

		return
	}

	attributes := make([]importAttribute, 0, len(attrPaths))
	for _, attrPath := range attrPaths {
		var identityValue types.String

		getDiags := req.Identity.GetAttribute(ctx, attrPath, &identityValue)
		resp.Diagnostics.Append(getDiags...)

		if getDiags.HasError() {
			return
		}

		if identityValue.IsUnknown() || identityValue.IsNull() || identityValue.ValueString() == "" {
			resp.Diagnostics.AddAttributeError(attrPath, "Resource Import Passthrough Multipart ID Mismatch", "Import identity components must be known, non-null, and non-empty.")

			return
		}

		attributes = append(attributes, importAttribute{path: attrPath, value: identityValue})
	}

	resp.Diagnostics.Append(setImportAttributes(ctx, &resp.State, resp.Identity, attributes)...)
}

func setImportAttributes(
	ctx context.Context,
	state *tfsdk.State,
	identity *tfsdk.ResourceIdentity,
	attributes []importAttribute,
) diag.Diagnostics {
	stagedState := *state

	var stagedIdentity tfsdk.ResourceIdentity
	if identity != nil {
		stagedIdentity = *identity
	}

	diags := diag.Diagnostics{}

	for _, attribute := range attributes {
		if identity != nil {
			diags.Append(stagedIdentity.SetAttribute(ctx, attribute.path, attribute.value)...)
		}

		diags.Append(stagedState.SetAttribute(ctx, attribute.path, attribute.value)...)
	}

	if diags.HasError() {
		return diags
	}

	*state = stagedState

	if identity != nil {
		*identity = stagedIdentity
	}

	return diags
}
