package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func editorInterfaceOptionalObjectList[T any, R any](
	ctx context.Context,
	valuePath path.Path,
	value TypedList[TypedObject[T]],
	convert knownObjectListElementConverter[T, R],
) ([]R, bool, diag.Diagnostics) {
	if value.IsUnknown() {
		diags := diag.Diagnostics{}
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown list",
			"The list value must be known before it can be sent to Contentful.",
		)

		return nil, false, diags
	}

	if value.IsNull() {
		return nil, false, nil
	}

	items, diags := convertKnownObjectListElements(ctx, valuePath, value.Elements(), convert)
	if diags.HasError() {
		return nil, false, diags
	}

	return items, true, diags
}

//nolint:ireturn // The concrete result type is supplied by each TypedObject caller.
func editorInterfaceRequiredObject[T any](value TypedObject[T], valuePath path.Path) (T, diag.Diagnostics) {
	if object, ok := value.GetValue(); ok {
		return object, nil
	}

	var zero T

	diags := diag.Diagnostics{}
	if value.IsUnknown() {
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown object",
			"The object value must be known before it can be sent to Contentful.",
		)
	} else {
		diags.AddAttributeError(
			valuePath,
			"Unexpected null object",
			"The required object value cannot be null.",
		)
	}

	return zero, diags
}

func editorInterfaceOptionalSettings(value jsontypes.Normalized, valuePath path.Path) ([]byte, diag.Diagnostics) {
	if value.IsUnknown() {
		diags := diag.Diagnostics{}
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown settings",
			"Settings must be known before they can be sent to Contentful.",
		)

		return nil, diags
	}

	if value.IsNull() || value.ValueString() == "" {
		return nil, nil
	}

	return []byte(value.ValueString()), nil
}

func toEditorInterfaceEditorLayoutGroupItem[T any](
	ctx context.Context,
	valuePath path.Path,
	groupID types.String,
	name types.String,
	items TypedList[TypedObject[T]],
	convert knownObjectListElementConverter[T, cm.EditorInterfaceEditorLayoutItem],
) (cm.EditorInterfaceEditorLayoutGroupItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	knownGroupID, groupIDDiags := requestRequiredString(groupID, valuePath.AtName("group_id"))
	diags.Append(groupIDDiags...)

	knownName, nameDiags := requestRequiredString(name, valuePath.AtName("name"))
	diags.Append(nameDiags...)

	itemsPath := valuePath.AtName("items")

	switch {
	case items.IsUnknown():
		diags.AddAttributeError(
			itemsPath,
			"Unexpected unknown editor layout items",
			"Editor layout items must be known before they can be sent to Contentful.",
		)
	case items.IsNull():
		diags.AddAttributeError(
			itemsPath,
			"Unexpected null editor layout items",
			"Editor layout items are required.",
		)
	}

	if diags.HasError() {
		return cm.EditorInterfaceEditorLayoutGroupItem{}, diags
	}

	knownItems, itemDiags := convertKnownObjectListElements(ctx, itemsPath, items.Elements(), convert)
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
