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

func TestSidebarValueToEditorInterfaceFieldsSidebarItem(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("sidebar")

	model := provider.NewEditorInterfaceSidebarValueKnown()
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetID = types.StringValue("widget_id")
	model.Disabled = types.BoolNull()
	model.Settings = jsontypes.NewNormalizedValue(`{"foo":"bar"}`)

	item, diags := model.ToEditorInterfaceFieldsSidebarItem(ctx, path)

	assert.EqualValues(t, "widget_namespace", item.WidgetNamespace)
	assert.EqualValues(t, "widget_id", item.WidgetId)
	assert.False(t, item.Disabled.Set)
	assert.NotEmpty(t, item.Settings)

	assert.Empty(t, diags)
}

func TestSidebarValueToEditorInterfaceFieldsSidebarItemInvalidSettings(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := path.Root("sidebar")

	model := provider.NewEditorInterfaceSidebarValueKnown()
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetID = types.StringValue("widget_id")
	model.Disabled = types.BoolNull()
	model.Settings = jsontypes.NewNormalizedValue(`invalid json`)

	sidebarItem, diags := model.ToEditorInterfaceFieldsSidebarItem(ctx, path)

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

	value, valueDiags := provider.NewEditorInterfaceSidebarValueFromResponse(path, item)

	assert.EqualValues(t, "widget_namespace", value.WidgetNamespace.ValueString())
	assert.EqualValues(t, "widget_id", value.WidgetID.ValueString())
	assert.JSONEq(t, `{"foo":"bar"}`, value.Settings.ValueString())

	assert.Empty(t, valueDiags)
}

func TestNewEditorInterfaceSidebarValueFromResponseSettingsNull(t *testing.T) {
	t.Parallel()

	path := path.Root("sidebar").AtListIndex(0)

	item := cm.EditorInterfaceSidebarItem{
		WidgetNamespace: "widget_namespace",
		WidgetId:        "widget_id",
	}

	value, valueDiags := provider.NewEditorInterfaceSidebarValueFromResponse(path, item)

	assert.EqualValues(t, "widget_namespace", value.WidgetNamespace.ValueString())
	assert.EqualValues(t, "widget_id", value.WidgetID.ValueString())
	assert.True(t, value.Settings.IsNull())

	assert.Empty(t, valueDiags)
}
