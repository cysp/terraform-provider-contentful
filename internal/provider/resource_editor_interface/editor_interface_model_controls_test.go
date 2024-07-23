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

func TestControlsValueToPutEditorInterfaceReqControlsItem(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := path.Root("controls")

	model := resource_editor_interface.NewControlsValueKnown()
	model.FieldId = types.StringValue("field_id")
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetId = types.StringValue("widget_id")
	model.Settings = types.StringValue(`{"foo":"bar"}`)

	item, diags := model.ToPutEditorInterfaceReqControlsItem(ctx, path)

	assert.EqualValues(t, "field_id", item.FieldId)
	assert.EqualValues(t, contentfulManagement.NewOptString("widget_namespace"), item.WidgetNamespace)
	assert.EqualValues(t, contentfulManagement.NewOptString("widget_id"), item.WidgetId)
	assert.True(t, item.Settings.Set)

	assert.Empty(t, diags)
}

func TestControlsValueToPutEditorInterfaceReqControlsItemInvalidSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := path.Root("controls")

	model := resource_editor_interface.NewControlsValueKnown()
	model.FieldId = types.StringValue("field_id")
	model.WidgetNamespace = types.StringValue("widget_namespace")
	model.WidgetId = types.StringValue("widget_id")
	model.Settings = types.StringValue(`invalid json`)

	controlsItem, diags := model.ToPutEditorInterfaceReqControlsItem(ctx, path)

	assert.NotNil(t, controlsItem)
	assert.NotEmpty(t, diags)
	assert.Len(t, diags, 1)
}

func TestNewControlsValueFromResponse(t *testing.T) {
	t.Parallel()

	path := path.Root("controls").AtListIndex(0)

	item := contentfulManagement.EditorInterfaceControlsItem{
		FieldId:         "field_id",
		WidgetNamespace: contentfulManagement.NewOptString("widget_namespace"),
		WidgetId:        contentfulManagement.NewOptString("widget_id"),
		Settings: contentfulManagement.NewOptEditorInterfaceControlsItemSettings(map[string]jx.Raw{
			"foo": jx.Raw(`"bar"`),
		}),
	}

	value := resource_editor_interface.NewControlsValueFromResponse(path, item)

	assert.EqualValues(t, "field_id", value.FieldId.ValueString())
	assert.EqualValues(t, "widget_namespace", value.WidgetNamespace.ValueString())
	assert.EqualValues(t, "widget_id", value.WidgetId.ValueString())
	assert.EqualValues(t, "{\"foo\":\"bar\"}", value.Settings.ValueString())
}

func TestNewControlsValueFromResponseSettingsNull(t *testing.T) {
	t.Parallel()

	path := path.Root("controls").AtListIndex(0)

	item := contentfulManagement.EditorInterfaceControlsItem{
		FieldId:         "field_id",
		WidgetNamespace: contentfulManagement.NewOptString("widget_namespace"),
		WidgetId:        contentfulManagement.NewOptString("widget_id"),
	}

	value := resource_editor_interface.NewControlsValueFromResponse(path, item)

	assert.EqualValues(t, "field_id", value.FieldId.ValueString())
	assert.EqualValues(t, "widget_namespace", value.WidgetNamespace.ValueString())
	assert.EqualValues(t, "widget_id", value.WidgetId.ValueString())
	assert.True(t, value.Settings.IsNull())
}
