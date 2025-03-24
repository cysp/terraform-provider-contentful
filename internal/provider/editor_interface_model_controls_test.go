package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestEditorInterfaceControlValueToEditorInterfaceFieldsControlsItem(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("controls")

	model := provider.NewEditorInterfaceControlValueKnown()
	model.FieldID = types.StringValue("field_id")
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetID = types.StringValue("widget_id")
	model.Settings = jsontypes.NewNormalizedValue(`{"foo":"bar"}`)

	item, diags := model.ToEditorInterfaceFieldsControlsItem(ctx, path)

	assert.Equal(t, "field_id", item.FieldId)
	assert.Equal(t, cm.NewOptString("widget_namespace"), item.WidgetNamespace)
	assert.Equal(t, cm.NewOptString("widget_id"), item.WidgetId)
	assert.NotEmpty(t, item.Settings)

	assert.Empty(t, diags)
}

func TestEditorInterfaceControlValueToEditorInterfaceFieldsControlsItemInvalidSettings(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("controls")

	model := provider.NewEditorInterfaceControlValueKnown()
	model.FieldID = types.StringValue("field_id")
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetID = types.StringValue("widget_id")
	model.Settings = jsontypes.NewNormalizedValue(`invalid json`)

	controlsItem, diags := model.ToEditorInterfaceFieldsControlsItem(ctx, path)

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

	value, diags := provider.NewEditorInterfaceControlValueFromResponse(path, item)

	assert.Equal(t, "field_id", value.FieldID.ValueString())
	assert.Equal(t, "widget_namespace", value.WidgetNamespace.ValueString())
	assert.Equal(t, "widget_id", value.WidgetID.ValueString())
	assert.JSONEq(t, `{"foo":"bar"}`, value.Settings.ValueString())

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

	value, diags := provider.NewEditorInterfaceControlValueFromResponse(path, item)

	assert.Equal(t, "field_id", value.FieldID.ValueString())
	assert.Equal(t, "widget_namespace", value.WidgetNamespace.ValueString())
	assert.Equal(t, "widget_id", value.WidgetID.ValueString())
	assert.True(t, value.Settings.IsNull())

	assert.Empty(t, diags)
}
