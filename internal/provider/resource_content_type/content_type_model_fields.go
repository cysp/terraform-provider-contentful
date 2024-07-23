//nolint:revive,stylecheck
package resource_content_type

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewFieldsListFromResponse(ctx context.Context, path path.Path, items []contentfulManagement.ContentTypeFieldsItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(items))

	for index, item := range items {
		path := path.AtListIndex(index)

		listElementValue, listElementValueDiags := NewFieldsValueFromResponse(ctx, path, item)
		diags.Append(listElementValueDiags...)

		listElementValues[index] = listElementValue
	}

	listValue, listValueDiags := types.ListValue(FieldsValue{}.Type(ctx), listElementValues)
	diags.Append(listValueDiags...)

	return listValue, diags
}

func NewFieldsValueFromResponse(ctx context.Context, path path.Path, item contentfulManagement.ContentTypeFieldsItem) (FieldsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := FieldsValue{
		Id:           types.StringValue(item.ID),
		Name:         types.StringValue(item.Name),
		FieldsType:   types.StringValue(item.Type),
		LinkType:     util.OptStringToStringValue(item.LinkType),
		Localized:    util.OptBoolToBoolValue(item.Localized),
		Disabled:     util.OptBoolToBoolValue(item.Disabled),
		Omitted:      util.OptBoolToBoolValue(item.Omitted),
		Required:     util.OptBoolToBoolValue(item.Required),
		Validations:  types.ListNull(types.StringType),
		DefaultValue: types.StringValue(item.DefaultValue.String()),
		state:        attr.ValueStateKnown,
	}

	itemsValue, itemsValueDiags := NewItemsValueFromResponse(ctx, path.AtName("items"), item.Items)
	diags.Append(itemsValueDiags...)

	itemsObjectValue, itemsObjectValueDiags := itemsValue.ToObjectValue(ctx)
	diags.Append(itemsObjectValueDiags...)

	value.Items = itemsObjectValue

	validationsList, validationsListDiags := NewValidationsListFromResponse(ctx, path.AtName("validations"), item.Validations)
	diags.Append(validationsListDiags...)

	value.Validations = validationsList

	return value, diags
}

func NewItemsValueFromResponse(ctx context.Context, path path.Path, item contentfulManagement.OptContentTypeFieldsItemItems) (ItemsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ItemsValue{
		state: attr.ValueStateNull,
	}

	if itemItems, ok := item.Get(); ok {
		value = ItemsValue{
			ItemsType:   util.OptStringToStringValue(itemItems.Type),
			LinkType:    util.OptStringToStringValue(itemItems.LinkType),
			Validations: types.ListNull(types.StringType),
			state:       attr.ValueStateKnown,
		}

		validationsList, validationsListDiags := NewValidationsListFromResponse(ctx, path.AtName("validations"), itemItems.Validations)
		diags.Append(validationsListDiags...)

		value.Validations = validationsList
	}

	return value, diags
}

func NewValidationsListFromResponse(ctx context.Context, path path.Path, validations []jx.Raw) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	validationElements := make([]attr.Value, len(validations))

	for i, validation := range validations {
		encoder := jx.Encoder{}
		encoder.Raw(validation)
		validationElements[i] = types.StringValue(encoder.String())
	}

	list, listDiags := types.ListValue(types.StringType, validationElements)
	diags.Append(listDiags...)

	return list, diags
}
