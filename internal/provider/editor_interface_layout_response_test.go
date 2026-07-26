package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestEditorLayoutGroupItemRejectsFieldItem(t *testing.T) {
	t.Parallel()

	actual, diags := NewEditorInterfaceEditorLayoutItemGroupValueFromResponse(
		t.Context(),
		path.Root("editor_layout"),
		cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(cm.EditorInterfaceEditorLayoutFieldItem{FieldId: "title"}),
	)

	assert.True(t, diags.HasError())
	assert.True(t, actual.IsNull())
}

func TestEditorLayoutNestedGroupItemFieldFromFieldItem(t *testing.T) {
	t.Parallel()

	actual, diags := NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueFromResponse(
		t.Context(),
		path.Root("items"),
		cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(cm.EditorInterfaceEditorLayoutFieldItem{FieldId: "title"}),
	)

	assert.Empty(t, diags)
	assert.Equal(t, NewTypedObject(EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue{
		FieldID: types.StringValue("title"),
	}), actual)
}

func TestEditorLayoutNestedGroupItemFieldRejectsGroupItem(t *testing.T) {
	t.Parallel()

	actual, diags := NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueFromResponse(
		t.Context(),
		path.Root("items"),
		cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(cm.EditorInterfaceEditorLayoutGroupItem{}),
	)

	assert.True(t, diags.HasError())
	assert.True(t, actual.IsNull())
}

func TestEditorLayoutRejectsUnknownItemTypes(t *testing.T) {
	t.Parallel()

	itemPath := path.Root("editor_layout").AtListIndex(0)
	item := cm.EditorInterfaceEditorLayoutItem{
		Type: cm.EditorInterfaceEditorLayoutItemType("unknown"),
	}

	groupItem, groupItemDiags := NewEditorInterfaceEditorLayoutItemValueFromResponse(t.Context(), itemPath, item)
	assert.True(t, groupItemDiags.HasError())
	assert.True(t, groupItem.IsNull())
	assert.Equal(t, []string{itemPath.String()}, diagnosticPaths(t, groupItemDiags))

	topLevelGroup, topLevelGroupDiags := NewEditorInterfaceEditorLayoutItemGroupValueFromResponse(t.Context(), itemPath, item)
	assert.True(t, topLevelGroupDiags.HasError())
	assert.True(t, topLevelGroup.IsNull())
	assert.Equal(t, []string{itemPath.String()}, diagnosticPaths(t, topLevelGroupDiags))

	nestedField, nestedFieldDiags := NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueFromResponse(t.Context(), itemPath, item)
	assert.True(t, nestedFieldDiags.HasError())
	assert.True(t, nestedField.IsNull())
	assert.Equal(t, []string{itemPath.String()}, diagnosticPaths(t, nestedFieldDiags))

	layout, layoutDiags := NewEditorInterfaceEditorLayoutListValueFromResponse(t.Context(), path.Root("editor_layout"), []cm.EditorInterfaceEditorLayoutItem{item})
	assert.True(t, layoutDiags.HasError())
	assert.True(t, layout.IsNull())
	assert.Equal(t, []string{itemPath.String()}, diagnosticPaths(t, layoutDiags))
}
