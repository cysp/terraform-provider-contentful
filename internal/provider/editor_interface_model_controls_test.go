package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestEditorInterfaceControlValueToEditorInterfaceDataControlsItem(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("controls")

	model := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceControlValue](ctx, map[string]attr.Value{
		"field_id":         types.StringValue("field_id"),
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         jsontypes.NewNormalizedValue(`{"foo":"bar"}`),
	}))

	item, diags := model.Value().ToEditorInterfaceDataControlsItem(ctx, path)

	assert.Equal(t, "field_id", item.FieldId)
	assert.Equal(t, cm.NewOptString("widget_namespace"), item.WidgetNamespace)
	assert.Equal(t, cm.NewOptString("widget_id"), item.WidgetId)
	assert.NotEmpty(t, item.Settings)

	assert.Empty(t, diags)
}

func TestEditorInterfaceControlValueToEditorInterfaceDataControlsItemInvalidSettings(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("controls")

	model := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceControlValue](ctx, map[string]attr.Value{
		"field_id":         types.StringValue("field_id"),
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"settings":         jsontypes.NewNormalizedValue(`invalid json`),
	}))

	controlsItem, diags := model.Value().ToEditorInterfaceDataControlsItem(ctx, path)

	assert.NotNil(t, controlsItem)
	assert.Empty(t, diags)
}

func TestNewEditorInterfaceControlValueFromResponse(t *testing.T) {
	t.Parallel()

	path := path.Root("controls").AtListIndex(0)

	item := cm.EditorInterfaceControlsItem{
		FieldId:         "field_id",
		WidgetNamespace: cm.NewOptString("widget_namespace"),
		WidgetId:        cm.NewOptString("widget_id"),
		Settings:        []byte(`{"foo":"bar"}`),
	}

	value, diags := NewEditorInterfaceControlValueFromResponse(path, item)

	assert.Equal(t, "field_id", value.Value().FieldID.ValueString())
	assert.Equal(t, "widget_namespace", value.Value().WidgetNamespace.ValueString())
	assert.Equal(t, "widget_id", value.Value().WidgetID.ValueString())
	assert.JSONEq(t, `{"foo":"bar"}`, value.Value().Settings.ValueString())

	assert.Empty(t, diags)
}

func TestNewEditorInterfaceControlValueFromResponseSettingsNull(t *testing.T) {
	t.Parallel()

	path := path.Root("controls").AtListIndex(0)

	item := cm.EditorInterfaceControlsItem{
		FieldId:         "field_id",
		WidgetNamespace: cm.NewOptString("widget_namespace"),
		WidgetId:        cm.NewOptString("widget_id"),
	}

	value, diags := NewEditorInterfaceControlValueFromResponse(path, item)

	assert.Equal(t, "field_id", value.Value().FieldID.ValueString())
	assert.Equal(t, "widget_namespace", value.Value().WidgetNamespace.ValueString())
	assert.Equal(t, "widget_id", value.Value().WidgetID.ValueString())
	assert.True(t, value.Value().Settings.IsNull())

	assert.Empty(t, diags)
}
