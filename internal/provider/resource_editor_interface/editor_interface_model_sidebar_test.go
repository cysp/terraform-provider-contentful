package resource_editor_interface_test

import (
	"context"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/resource_editor_interface"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSidebarValueToPutEditorInterfaceReqSidebarItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := path.Root("sidebar")

	model := resource_editor_interface.NewSidebarValueKnown()
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetId = types.StringValue("widget_id")
	model.Disabled = types.BoolNull()
	model.Settings = types.StringValue(`{"foo":"bar"}`)

	item, diags := model.ToPutEditorInterfaceReqSidebarItem(ctx, path)

	assert.EqualValues(t, "widget_namespace", item.WidgetNamespace)
	assert.EqualValues(t, "widget_id", item.WidgetId)
	assert.False(t, item.Disabled.Set)
	assert.True(t, item.Settings.Set)

	assert.Empty(t, diags)
}

func TestSidebarValueToPutEditorInterfaceReqSidebarItemInvalidSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := path.Root("sidebar")

	model := resource_editor_interface.NewSidebarValueKnown()
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetId = types.StringValue("widget_id")
	model.Disabled = types.BoolNull()
	model.Settings = types.StringValue(`invalid json`)

	sidebarItem, diags := model.ToPutEditorInterfaceReqSidebarItem(ctx, path)

	assert.NotNil(t, sidebarItem)
	assert.NotEmpty(t, diags)
	assert.Len(t, diags, 1)
}

func TestNewSidebarValueFromResponse(t *testing.T) {
	t.Parallel()

	path := path.Root("sidebar").AtListIndex(0)

	item := contentfulManagement.EditorInterfaceSidebarItem{
		WidgetNamespace: "widget_namespace",
		WidgetId:        "widget_id",
		Settings: contentfulManagement.NewOptEditorInterfaceSidebarItemSettings(map[string]jx.Raw{
			"foo": jx.Raw(`"bar"`),
		}),
	}

	value := resource_editor_interface.NewSidebarValueFromResponse(path, item)

	assert.EqualValues(t, "widget_namespace", value.WidgetNamespace.ValueString())
	assert.EqualValues(t, "widget_id", value.WidgetId.ValueString())
	assert.EqualValues(t, "{\"foo\":\"bar\"}", value.Settings.ValueString())
}

func TestNewSidebarValueFromResponseSettingsNull(t *testing.T) {
	t.Parallel()

	path := path.Root("sidebar").AtListIndex(0)

	item := contentfulManagement.EditorInterfaceSidebarItem{
		WidgetNamespace: "widget_namespace",
		WidgetId:        "widget_id",
	}

	value := resource_editor_interface.NewSidebarValueFromResponse(path, item)

	assert.EqualValues(t, "widget_namespace", value.WidgetNamespace.ValueString())
	assert.EqualValues(t, "widget_id", value.WidgetId.ValueString())
	assert.True(t, value.Settings.IsNull())
}
