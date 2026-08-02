package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func toEditorInterfaceEditorLayoutGroupItem[T any](
	ctx context.Context,
	valuePath path.Path,
	groupID types.String,
	name types.String,
	items TypedList[TypedObject[T]],
	convert func(context.Context, path.Path, T) (cm.EditorInterfaceEditorLayoutItem, diag.Diagnostics),
) (cm.EditorInterfaceEditorLayoutGroupItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	knownGroupID, groupIDDiags := KnownStringValue(groupID, valuePath.AtName("group_id"))
	diags.Append(groupIDDiags...)

	knownName, nameDiags := KnownStringValue(name, valuePath.AtName("name"))
	diags.Append(nameDiags...)

	switch {
	case items.IsUnknown():
		diags.AddAttributeError(valuePath.AtName("items"), "Unexpected unknown editor layout items", "Editor layout items must be known before they can be sent to Contentful.")
	case items.IsNull():
		diags.AddAttributeError(valuePath.AtName("items"), "Unexpected null editor layout items", "Editor layout items are required.")
	}

	if diags.HasError() {
		return cm.EditorInterfaceEditorLayoutGroupItem{}, diags
	}

	knownItems, itemDiags := ConvertKnownObjectListElements(ctx, valuePath.AtName("items"), items.Elements(), convert)
	diags.Append(itemDiags...)

	if diags.HasError() {
		return cm.EditorInterfaceEditorLayoutGroupItem{}, diags
	}

	return cm.EditorInterfaceEditorLayoutGroupItem{
		GroupId: knownGroupID,
		Name:    knownName,
		Items:   knownItems,
	}, diags
}
