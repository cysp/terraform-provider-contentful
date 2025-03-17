package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewFieldsListFromResponse(ctx context.Context, path path.Path, items []cm.ContentTypeFieldsItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(items))

	for index, item := range items {
		path := path.AtListIndex(index)

		listElementValues[index] = DiagsAppendResult3(diags, NewFieldsValueFromResponse, ctx, path, item)
	}

	listValue := DiagsAppendResult2(diags, types.ListValue, ContentTypeFieldValue{}.Type(ctx), listElementValues)

	return listValue, diags
}

func NewFieldsValueFromResponse(ctx context.Context, path path.Path, item cm.ContentTypeFieldsItem) (ContentTypeFieldValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	defaultValueValue := jsontypes.NewNormalizedNull()
	if item.DefaultValue != nil {
		defaultValueValue = jsontypes.NewNormalizedValue(item.DefaultValue.String())
	}

	value := ContentTypeFieldValue{
		ID:           types.StringValue(item.ID),
		Name:         types.StringValue(item.Name),
		FieldType:    types.StringValue(item.Type),
		LinkType:     util.OptStringToStringValue(item.LinkType),
		DefaultValue: defaultValueValue,
		Localized:    util.OptBoolToBoolValue(item.Localized),
		Disabled:     util.OptBoolToBoolValue(item.Disabled),
		Omitted:      util.OptBoolToBoolValue(item.Omitted),
		Required:     util.OptBoolToBoolValue(item.Required),
		Validations:  types.ListNull(jsontypes.NormalizedType{}),
		state:        attr.ValueStateKnown,
	}

	itemsValue, itemsValueDiags := NewItemsValueFromResponse(ctx, path.AtName("items"), item.Items)
	diags.Append(itemsValueDiags...)

	value.Items = itemsValue

	validationsList, validationsListDiags := NewValidationsListFromResponse(ctx, path.AtName("validations"), item.Validations)
	diags.Append(validationsListDiags...)

	value.Validations = validationsList

	return value, diags
}

func NewItemsValueFromResponse(ctx context.Context, path path.Path, item cm.OptContentTypeFieldsItemItems) (ContentTypeFieldItemsValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := ContentTypeFieldItemsValue{
		state: attr.ValueStateNull,
	}

	if itemItems, ok := item.Get(); ok {
		value = ContentTypeFieldItemsValue{
			ItemsType:   util.OptStringToStringValue(itemItems.Type),
			LinkType:    util.OptStringToStringValue(itemItems.LinkType),
			Validations: types.ListNull(jsontypes.NormalizedType{}),
			state:       attr.ValueStateKnown,
		}

		validationsList, validationsListDiags := NewValidationsListFromResponse(ctx, path.AtName("validations"), itemItems.Validations)
		diags.Append(validationsListDiags...)

		value.Validations = validationsList
	}

	return value, diags
}

func NewValidationsListFromResponse(_ context.Context, _ path.Path, validations []jx.Raw) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	validationElements := make([]attr.Value, len(validations))

	for i, validation := range validations {
		encoder := jx.Encoder{}
		encoder.Raw(validation)
		validationElements[i] = jsontypes.NewNormalizedValue(encoder.String())
	}

	list, listDiags := types.ListValue(jsontypes.NormalizedType{}, validationElements)
	diags.Append(listDiags...)

	return list, diags
}
