package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditorLayoutGroupItemPreservesFieldItemAsKnownNull(t *testing.T) {
	t.Parallel()

	actual, diags := NewEditorInterfaceEditorLayoutItemGroupValueFromResponse(
		t.Context(),
		path.Root("editor_layout"),
		cm.NewEditorInterfaceEditorLayoutFieldItemEditorInterfaceEditorLayoutItem(cm.EditorInterfaceEditorLayoutFieldItem{FieldId: "title"}),
	)

	assert.False(t, diags.HasError())
	assert.Len(t, diags.Warnings(), 1)
	assert.Equal(t, []string{"editor_layout"}, editorLayoutWarningPaths(t, diags))
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

func TestEditorLayoutNestedGroupItemFieldPreservesGroupItemAsKnownNull(t *testing.T) {
	t.Parallel()

	actual, diags := NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueFromResponse(
		t.Context(),
		path.Root("items"),
		cm.NewEditorInterfaceEditorLayoutGroupItemEditorInterfaceEditorLayoutItem(cm.EditorInterfaceEditorLayoutGroupItem{}),
	)

	assert.False(t, diags.HasError())
	assert.Len(t, diags.Warnings(), 1)
	assert.Equal(t, []string{"items"}, editorLayoutWarningPaths(t, diags))
	assert.True(t, actual.IsNull())
}

func TestEditorLayoutPreservesUnknownItemTypesAndListPositions(t *testing.T) {
	t.Parallel()

	itemPath := path.Root("editor_layout").AtListIndex(0)
	item := cm.EditorInterfaceEditorLayoutItem{
		Type: cm.EditorInterfaceEditorLayoutItemType("unknown"),
	}

	groupItem, groupItemDiags := NewEditorInterfaceEditorLayoutItemValueFromResponse(t.Context(), itemPath, item)
	assert.False(t, groupItemDiags.HasError())
	assert.Len(t, groupItemDiags.Warnings(), 1)
	assert.Equal(t, []string{itemPath.String()}, editorLayoutWarningPaths(t, groupItemDiags))
	assert.False(t, groupItem.IsNull())
	assert.True(t, groupItem.Value().Field.IsNull())
	assert.True(t, groupItem.Value().Group.IsNull())

	topLevelGroup, topLevelGroupDiags := NewEditorInterfaceEditorLayoutItemGroupValueFromResponse(t.Context(), itemPath, item)
	assert.False(t, topLevelGroupDiags.HasError())
	assert.Len(t, topLevelGroupDiags.Warnings(), 1)
	assert.Equal(t, []string{itemPath.String()}, editorLayoutWarningPaths(t, topLevelGroupDiags))
	assert.True(t, topLevelGroup.IsNull())

	nestedField, nestedFieldDiags := NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueFromResponse(t.Context(), itemPath, item)
	assert.False(t, nestedFieldDiags.HasError())
	assert.Len(t, nestedFieldDiags.Warnings(), 1)
	assert.Equal(t, []string{itemPath.String()}, editorLayoutWarningPaths(t, nestedFieldDiags))
	assert.True(t, nestedField.IsNull())

	layout, layoutDiags := NewEditorInterfaceEditorLayoutListValueFromResponse(t.Context(), path.Root("editor_layout"), []cm.EditorInterfaceEditorLayoutItem{item})
	assert.False(t, layoutDiags.HasError())
	assert.Len(t, layoutDiags.Warnings(), 1)
	assert.Equal(t, []string{itemPath.String()}, editorLayoutWarningPaths(t, layoutDiags))
	assert.False(t, layout.IsNull())
	require.Len(t, layout.Elements(), 1)
	assert.True(t, layout.Elements()[0].Value().Group.IsNull())
}

func editorLayoutWarningPaths(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	paths := make([]string, 0, len(diags.Warnings()))
	for _, diagnostic := range diags.Warnings() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok)

		paths = append(paths, withPath.Path().String())
	}

	return paths
}
