package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestControlsValueToEditorInterfaceFieldsControlsItem(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("controls")

	model := provider.NewControlsValueKnown()
	model.FieldId = types.StringValue("field_id")
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetId = types.StringValue("widget_id")
	model.Settings = types.StringValue(`{"foo":"bar"}`)

	item, diags := model.ToEditorInterfaceFieldsControlsItem(ctx, path)

	assert.EqualValues(t, "field_id", item.FieldId)
	assert.EqualValues(t, cm.NewOptString("widget_namespace"), item.WidgetNamespace)
	assert.EqualValues(t, cm.NewOptString("widget_id"), item.WidgetId)
	assert.NotEmpty(t, item.Settings)

	assert.Empty(t, diags)
}

func TestControlsValueToEditorInterfaceFieldsControlsItemInvalidSettings(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("controls")

	model := provider.NewControlsValueKnown()
	model.FieldId = types.StringValue("field_id")
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetId = types.StringValue("widget_id")
	model.Settings = types.StringValue(`invalid json`)

	controlsItem, diags := model.ToEditorInterfaceFieldsControlsItem(ctx, path)

	assert.NotNil(t, controlsItem)
	assert.Empty(t, diags)
}

func TestNewControlsValueFromResponse(t *testing.T) {
	t.Parallel()

	path := path.Root("controls").AtListIndex(0)

	item := cm.EditorInterfaceControlsItem{
		FieldId:         "field_id",
		WidgetNamespace: cm.NewOptString("widget_namespace"),
		WidgetId:        cm.NewOptString("widget_id"),
		Settings:        []byte(`{"foo":"bar"}`),
	}

	value, diags := provider.NewControlsValueFromResponse(path, item)

	assert.EqualValues(t, "field_id", value.FieldId.ValueString())
	assert.EqualValues(t, "widget_namespace", value.WidgetNamespace.ValueString())
	assert.EqualValues(t, "widget_id", value.WidgetId.ValueString())
	assert.JSONEq(t, `{"foo":"bar"}`, value.Settings.ValueString())

	assert.Empty(t, diags)
}

func TestNewControlsValueFromResponseSettingsNull(t *testing.T) {
	t.Parallel()

	path := path.Root("controls").AtListIndex(0)

	item := cm.EditorInterfaceControlsItem{
		FieldId:         "field_id",
		WidgetNamespace: cm.NewOptString("widget_namespace"),
		WidgetId:        cm.NewOptString("widget_id"),
	}

	value, diags := provider.NewControlsValueFromResponse(path, item)

	assert.EqualValues(t, "field_id", value.FieldId.ValueString())
	assert.EqualValues(t, "widget_namespace", value.WidgetNamespace.ValueString())
	assert.EqualValues(t, "widget_id", value.WidgetId.ValueString())
	assert.True(t, value.Settings.IsNull())

	assert.Empty(t, diags)
}
