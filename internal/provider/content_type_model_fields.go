package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewFieldsListFromResponse(ctx context.Context, path path.Path, items []cm.ContentTypeFieldsItem) (TypedList[TypedObject[ContentTypeFieldValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]TypedObject[ContentTypeFieldValue], len(items))

	for index, item := range items {
		path := path.AtListIndex(index)

		listElementValue, listElementValueDiags := NewFieldsValueFromResponse(ctx, path, item)
		diags.Append(listElementValueDiags...)

		listElementValues[index] = listElementValue
	}

	list := NewTypedList(listElementValues)

	return list, diags
}

func NewFieldsValueFromResponse(ctx context.Context, path path.Path, item cm.ContentTypeFieldsItem) (TypedObject[ContentTypeFieldValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	defaultValueValue := jsontypes.NewNormalizedNull()
	if item.DefaultValue != nil {
		defaultValueValue = NewNormalizedJSONValue(item.DefaultValue)
	}

	value := ContentTypeFieldValue{
		ID:               types.StringValue(item.ID),
		Name:             types.StringValue(item.Name),
		FieldType:        types.StringValue(item.Type),
		LinkType:         util.OptStringToStringValue(item.LinkType),
		DefaultValue:     defaultValueValue,
		Localized:        util.OptBoolToBoolValue(item.Localized),
		Disabled:         util.OptBoolToBoolValue(item.Disabled),
		Omitted:          util.OptBoolToBoolValue(item.Omitted),
		Required:         util.OptBoolToBoolValue(item.Required),
		Validations:      NewTypedListNull[jsontypes.Normalized](),
		AllowedResources: NewTypedListNull[TypedObject[ContentTypeFieldAllowedResourceItemValue]](),
	}

	itemsValue, itemsValueDiags := NewItemsValueFromResponse(ctx, path.AtName("items"), item.Items)
	diags.Append(itemsValueDiags...)

	value.Items = itemsValue

	validationsList, validationsListDiags := NewValidationsListFromResponse(ctx, path.AtName("validations"), item.Validations)
	diags.Append(validationsListDiags...)

	value.Validations = validationsList

	if allowedResources, ok := item.GetAllowedResources().Get(); ok {
		value.AllowedResources = NewContentTypeFieldAllowedResourcesListFromResponse(allowedResources)
	}

	return NewTypedObject(value), diags
}

func NewItemsValueFromResponse(ctx context.Context, path path.Path, item cm.OptContentTypeFieldsItemItems) (TypedObject[ContentTypeFieldItemsValue], diag.Diagnostics) {
	itemItems, itemItemsOk := item.Get()
	if !itemItemsOk {
		return NewTypedObjectNull[ContentTypeFieldItemsValue](), nil
	}

	diags := diag.Diagnostics{}

	value := ContentTypeFieldItemsValue{
		ItemsType:   util.OptStringToStringValue(itemItems.Type),
		LinkType:    util.OptStringToStringValue(itemItems.LinkType),
		Validations: NewTypedListNull[jsontypes.Normalized](),
	}

	validationsList, validationsListDiags := NewValidationsListFromResponse(ctx, path.AtName("validations"), itemItems.Validations)
	diags.Append(validationsListDiags...)

	value.Validations = validationsList

	return NewTypedObject(value), diags
}

func NewValidationsListFromResponse(_ context.Context, _ path.Path, validations []jx.Raw) (TypedList[jsontypes.Normalized], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	validationElements := make([]jsontypes.Normalized, len(validations))

	for i, validation := range validations {
		encoder := jx.Encoder{}
		encoder.Raw(validation)
		validationElements[i] = NewNormalizedJSONValue(encoder.Bytes())
	}

	list := NewTypedList(validationElements)

	return list, diags
}

func NewContentTypeFieldAllowedResourcesListFromResponse(resourceLinks []cm.ResourceLink) TypedList[TypedObject[ContentTypeFieldAllowedResourceItemValue]] {
	allowedResourceElements := make([]TypedObject[ContentTypeFieldAllowedResourceItemValue], len(resourceLinks))

	for i, resourceLink := range resourceLinks {
		switch resourceLink.Type {
		case contentfulEntryAllowedResourceType:
			allowedResourceElements[i] = NewTypedObject(ContentTypeFieldAllowedResourceItemValue{
				ContentfulEntry: NewTypedObject(ContentTypeFieldAllowedResourceItemContentfulEntryValue{
					Source:       types.StringPointerValue(resourceLink.Source.ValueStringPointer()),
					ContentTypes: NewTypedListFromStringSlice(resourceLink.ContentTypes),
				}),
				External: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemExternalValue](),
			})

		default:
			allowedResourceElements[i] = NewTypedObject(ContentTypeFieldAllowedResourceItemValue{
				ContentfulEntry: NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue](),
				External: NewTypedObject(ContentTypeFieldAllowedResourceItemExternalValue{
					TypeID: types.StringValue(resourceLink.Type),
				}),
			})
		}
	}

	return NewTypedList(allowedResourceElements)
}
