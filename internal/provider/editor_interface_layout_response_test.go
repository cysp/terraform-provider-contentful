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
	assert.False(t, actual.IsNull())
}
