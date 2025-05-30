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

func NewFieldsListFromResponse(ctx context.Context, path path.Path, items []cm.ContentTypeFieldsItem) (TypedList[ContentTypeFieldValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]ContentTypeFieldValue, len(items))

	for index, item := range items {
		path := path.AtListIndex(index)

		listElementValue, listElementValueDiags := NewFieldsValueFromResponse(ctx, path, item)
		diags.Append(listElementValueDiags...)

		listElementValues[index] = listElementValue
	}

	list, listDiags := NewTypedList(ctx, listElementValues)
	diags.Append(listDiags...)

	return list, diags
}

func NewFieldsValueFromResponse(ctx context.Context, path path.Path, item cm.ContentTypeFieldsItem) (ContentTypeFieldValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	defaultValueValue := jsontypes.NewNormalizedNull()
	if item.DefaultValue != nil {
		defaultValueValue = jsontypes.NewNormalizedValue(item.DefaultValue.String())
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
		Validations:      NewTypedListNull[jsontypes.Normalized](ctx),
		AllowedResources: NewTypedListNull[ContentTypeFieldAllowedResourceItemValue](ctx),
		state:            attr.ValueStateKnown,
	}

	itemsValue, itemsValueDiags := NewItemsValueFromResponse(ctx, path.AtName("items"), item.Items)
	diags.Append(itemsValueDiags...)

	value.Items = itemsValue

	validationsList, validationsListDiags := NewValidationsListFromResponse(ctx, path.AtName("validations"), item.Validations)
	diags.Append(validationsListDiags...)

	value.Validations = validationsList

	if allowedResources, ok := item.GetAllowedResources().Get(); ok {
		allowedResourcesList, allowedResourcesListDiags := NewContentTypeFieldAllowedResourcesListFromResponse(ctx, path.AtName("allowed_resources"), allowedResources)
		diags.Append(allowedResourcesListDiags...)

		value.AllowedResources = allowedResourcesList
	}

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
			Validations: NewTypedListNull[jsontypes.Normalized](ctx),
			state:       attr.ValueStateKnown,
		}

		validationsList, validationsListDiags := NewValidationsListFromResponse(ctx, path.AtName("validations"), itemItems.Validations)
		diags.Append(validationsListDiags...)

		value.Validations = validationsList
	}

	return value, diags
}

func NewValidationsListFromResponse(ctx context.Context, _ path.Path, validations []jx.Raw) (TypedList[jsontypes.Normalized], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	validationElements := make([]jsontypes.Normalized, len(validations))

	for i, validation := range validations {
		encoder := jx.Encoder{}
		encoder.Raw(validation)
		validationElements[i] = jsontypes.NewNormalizedValue(encoder.String())
	}

	list, listDiags := NewTypedList(ctx, validationElements)
	diags.Append(listDiags...)

	return list, diags
}

func NewContentTypeFieldAllowedResourcesListFromResponse(ctx context.Context, path path.Path, resourceLinks []cm.ResourceLink) (TypedList[ContentTypeFieldAllowedResourceItemValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	allowedResourceElements := make([]ContentTypeFieldAllowedResourceItemValue, len(resourceLinks))

	for i, resourceLink := range resourceLinks {
		path := path.AtListIndex(i)

		switch resourceLink.Type {
		case cm.ContentfulEntryResourceLinkResourceLink:
			contentfulEntryResourceLink, contentfulEntryResourceLinkOk := resourceLink.GetContentfulEntryResourceLink()
			if !contentfulEntryResourceLinkOk {
				diags.AddAttributeError(path, "Invalid data", "Expected contentful entry resource link")

				break
			}

			contentTypesList, contentTypesListDiags := NewTypedListFromStringSlice(ctx, contentfulEntryResourceLink.ContentTypes)
			diags.Append(contentTypesListDiags...)

			contentfulEntryResourceItem, contentfulEntryResourceItemDiags := NewContentTypeFieldAllowedResourceItemContentfulEntryValueKnownFromAttributes(ctx, map[string]attr.Value{
				"source":        types.StringValue(contentfulEntryResourceLink.Source),
				"content_types": contentTypesList,
			})
			diags.Append(contentfulEntryResourceItemDiags...)

			allowedResourceItem, allowedResourceItemDiags := NewContentTypeFieldAllowedResourceItemValueKnownFromAttributes(ctx, map[string]attr.Value{
				"contentful_entry": contentfulEntryResourceItem,
				"external":         NewContentTypeFieldAllowedResourceItemExternalValueNull(),
			})
			diags.Append(allowedResourceItemDiags...)

			allowedResourceElements[i] = allowedResourceItem

		case cm.ExternalResourceLinkResourceLink:
			externalResourceLink, externalResourceLinkOk := resourceLink.GetExternalResourceLink()
			if !externalResourceLinkOk {
				diags.AddAttributeError(path, "Invalid data", "Expected external resource link")

				break
			}

			externalResourceItem, externalResourceItemDiags := NewContentTypeFieldAllowedResourceItemExternalValueKnownFromAttributes(ctx, map[string]attr.Value{
				"type": types.StringValue(externalResourceLink.Type),
			})
			diags.Append(externalResourceItemDiags...)

			allowedResourceItem, allowedResourceItemDiags := NewContentTypeFieldAllowedResourceItemValueKnownFromAttributes(ctx, map[string]attr.Value{
				"external":         externalResourceItem,
				"contentful_entry": NewContentTypeFieldAllowedResourceItemContentfulEntryValueNull(),
			})
			diags.Append(allowedResourceItemDiags...)

			allowedResourceElements[i] = allowedResourceItem

		default:
			diags.AddAttributeError(path, "Invalid data", "Unknown resource link type: "+string(resourceLink.Type))
		}
	}

	list, listDiags := NewTypedList(ctx, allowedResourceElements)
	diags.Append(listDiags...)

	return list, diags
}
