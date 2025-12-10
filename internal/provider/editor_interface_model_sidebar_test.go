package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSidebarValueToEditorInterfaceDataSidebarItem(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("sidebar")

	model := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceSidebarValue](ctx, map[string]attr.Value{
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"disabled":         types.BoolNull(),
		"settings":         NewNormalizedJSONTypesNormalizedValue([]byte(`{"foo":"bar"}`)),
	}))

	item, diags := model.Value().ToEditorInterfaceDataSidebarItem(ctx, path)

	assert.Equal(t, "widget_namespace", item.WidgetNamespace)
	assert.Equal(t, "widget_id", item.WidgetId)
	assert.False(t, item.Disabled.Set)
	assert.NotEmpty(t, item.Settings)

	assert.Empty(t, diags)
}

func TestSidebarValueToEditorInterfaceDataSidebarItemInvalidSettings(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("sidebar")

	model := DiagsNoErrorsMust(NewTypedObjectFromAttributes[EditorInterfaceSidebarValue](ctx, map[string]attr.Value{
		"widget_namespace": types.StringValue("widget_namespace"),
		"widget_id":        types.StringValue("widget_id"),
		"disabled":         types.BoolNull(),
		"settings":         NewNormalizedJSONTypesNormalizedValue([]byte(`invalid json`)),
	}))

	sidebarItem, diags := model.Value().ToEditorInterfaceDataSidebarItem(ctx, path)

	assert.NotNil(t, sidebarItem)
	assert.Empty(t, diags)
}

func TestNewEditorInterfaceSidebarValueFromResponse(t *testing.T) {
	t.Parallel()

	path := path.Root("sidebar").AtListIndex(0)

	item := cm.EditorInterfaceSidebarItem{
		WidgetNamespace: "widget_namespace",
		WidgetId:        "widget_id",
		Settings:        []byte(`{"foo":"bar"}`),
	}

	value, valueDiags := NewEditorInterfaceSidebarValueFromResponse(path, item)

	assert.Equal(t, "widget_namespace", value.Value().WidgetNamespace.ValueString())
	assert.Equal(t, "widget_id", value.Value().WidgetID.ValueString())
	assert.JSONEq(t, `{"foo":"bar"}`, value.Value().Settings.ValueString())

	assert.Empty(t, valueDiags)
}

func TestNewEditorInterfaceSidebarValueFromResponseSettingsNull(t *testing.T) {
	t.Parallel()

	path := path.Root("sidebar").AtListIndex(0)

	item := cm.EditorInterfaceSidebarItem{
		WidgetNamespace: "widget_namespace",
		WidgetId:        "widget_id",
	}

	value, valueDiags := NewEditorInterfaceSidebarValueFromResponse(path, item)

	assert.Equal(t, "widget_namespace", value.Value().WidgetNamespace.ValueString())
	assert.Equal(t, "widget_id", value.Value().WidgetID.ValueString())
	assert.True(t, value.Value().Settings.IsNull())

	assert.Empty(t, valueDiags)
}
