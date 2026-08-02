package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

const contentfulEntryAllowedResourceType = "Contentful:Entry"

func (m *ContentTypeModel) ToContentTypeRequestData(ctx context.Context) (cm.ContentTypeRequestData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	request := cm.ContentTypeRequestData{
		Name:         m.Name.ValueString(),
		Description:  cm.NewOptNilPointerString(m.Description.ValueStringPointer()),
		DisplayField: m.DisplayField.ValueString(),
	}

	fields, fieldsDiags := FieldsListToContentTypeRequestDataFields(ctx, path.Root("fields"), m.Fields)
	diags.Append(fieldsDiags...)

	request.Fields = fields

	metadata, metadataDiags := ToOptContentTypeMetadata(ctx, path.Root("metadata"), m.Metadata)
	diags.Append(metadataDiags...)

	request.Metadata = metadata

	if diags.HasError() {
		return cm.ContentTypeRequestData{}, diags
	}

	return request, diags
}

func FieldsListToContentTypeRequestDataFields(ctx context.Context, path path.Path, fieldsList TypedList[TypedObject[ContentTypeFieldValue]]) ([]cm.ContentTypeRequestDataFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if fieldsList.IsNull() || fieldsList.IsUnknown() {
		if fieldsList.IsUnknown() {
			diags.AddAttributeError(path, "Unexpected unknown content type fields", "Content type fields must be known before they can be sent to Contentful.")
		} else {
			diags.AddAttributeError(path, "Unexpected null content type fields", "Content type fields are required.")
		}

		return nil, diags
	}

	fieldsItems, fieldsDiags := ConvertKnownObjectListElements(ctx, path, fieldsList.Elements(), ToContentTypeRequestDataFieldsItem)
	diags.Append(fieldsDiags...)

	return fieldsItems, diags
}

func ToContentTypeRequestDataFieldsItem(ctx context.Context, path path.Path, v ContentTypeFieldValue) (cm.ContentTypeRequestDataFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsItemItems, fieldsItemItemsDiags := ItemsObjectToOptContentTypeRequestDataFieldsItemItems(ctx, path.AtName("items"), v.Items)
	diags.Append(fieldsItemItemsDiags...)

	fieldsItemValidations, fieldsItemValidationsDiags := ValidationsListToContentTypeRequestDataFieldValidations(ctx, path.AtName("validations"), v.Validations)
	diags.Append(fieldsItemValidationsDiags...)

	fieldsItem := cm.ContentTypeRequestDataFieldsItem{
		ID:          v.ID.ValueString(),
		Name:        v.Name.ValueString(),
		Type:        v.FieldType.ValueString(),
		LinkType:    util.StringValueToOptString(v.LinkType),
		Items:       fieldsItemItems,
		Validations: fieldsItemValidations,
		Disabled:    util.BoolValueToOptBool(v.Disabled),
		Omitted:     util.BoolValueToOptBool(v.Omitted),
		Required:    util.BoolValueToOptBool(v.Required),
		Localized:   util.BoolValueToOptBool(v.Localized),
	}

	modelDefaultValueValue := v.DefaultValue.ValueString()
	if modelDefaultValueValue != "" {
		fieldsItem.DefaultValue = []byte(modelDefaultValueValue)
	}

	if v.AllowedResources.IsUnknown() {
		diags.AddAttributeError(path.AtName("allowed_resources"), "Unexpected unknown allowed resources", "Allowed resources must be known before they can be sent to Contentful.")
	} else if !v.AllowedResources.IsNull() {
		fieldsItemAllowedResources, fieldsItemAllowedResourcesDiags := AllowedResourceListToContentTypeRequestDataFieldAllowedResources(ctx, path.AtName("allowed_resources"), v.AllowedResources)
		diags.Append(fieldsItemAllowedResourcesDiags...)

		fieldsItem.AllowedResources.SetTo(fieldsItemAllowedResources)
	}

	if diags.HasError() {
		return cm.ContentTypeRequestDataFieldsItem{}, diags
	}

	return fieldsItem, diags
}

func ItemsObjectToOptContentTypeRequestDataFieldsItemItems(ctx context.Context, path path.Path, itemsObject TypedObject[ContentTypeFieldItemsValue]) (cm.OptContentTypeRequestDataFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsItemItems := cm.OptContentTypeRequestDataFieldsItemItems{}

	itemsValue, itemsValueOk := itemsObject.GetValue()
	if itemsValueOk {
		items, itemsDiags := itemsValue.ToContentTypeRequestDataFieldsItemItems(ctx, path)
		diags.Append(itemsDiags...)

		if !itemsDiags.HasError() {
			fieldsItemItems.SetTo(items)
		}
	}

	if diags.HasError() {
		return cm.OptContentTypeRequestDataFieldsItemItems{}, diags
	}

	return fieldsItemItems, diags
}

func (v ContentTypeFieldItemsValue) ToContentTypeRequestDataFieldsItemItems(ctx context.Context, path path.Path) (cm.ContentTypeRequestDataFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	itemsValidations, itemsValidationsDiags := ValidationsListToContentTypeRequestDataFieldValidations(ctx, path.AtName("validations"), v.Validations)
	diags.Append(itemsValidationsDiags...)

	items := cm.ContentTypeRequestDataFieldsItemItems{
		Type:        util.StringValueToOptString(v.ItemsType),
		LinkType:    util.StringValueToOptString(v.LinkType),
		Validations: itemsValidations,
	}

	if diags.HasError() {
		return cm.ContentTypeRequestDataFieldsItemItems{}, diags
	}

	return items, diags
}

func ValidationsListToContentTypeRequestDataFieldValidations(_ context.Context, valuePath path.Path, validationsList TypedList[jsontypes.Normalized]) ([]jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	validationsStrings := []string{}

	if !validationsList.IsNull() && !validationsList.IsUnknown() {
		var valueDiags diag.Diagnostics

		validationsStrings, valueDiags = KnownStringValues(validationsList.Elements(), valuePath)
		diags.Append(valueDiags...)
	}

	if diags.HasError() {
		return nil, diags
	}

	validations := make([]jx.Raw, len(validationsStrings))
	for index, validationsString := range validationsStrings {
		validations[index] = jx.Raw(validationsString)
	}

	return validations, diags
}

func AllowedResourceListToContentTypeRequestDataFieldAllowedResources(ctx context.Context, valuePath path.Path, allowedResourcesList TypedList[TypedObject[ContentTypeFieldAllowedResourceItemValue]]) ([]cm.ResourceLink, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	resourceLinks, resourceDiags := ConvertKnownObjectListElements(
		ctx,
		valuePath,
		allowedResourcesList.Elements(),
		func(ctx context.Context, resourcePath path.Path, value ContentTypeFieldAllowedResourceItemValue) (cm.ResourceLink, diag.Diagnostics) {
			return value.ToResourceLink(ctx, resourcePath)
		},
	)
	diags.Append(resourceDiags...)

	return resourceLinks, diags
}

func (v ContentTypeFieldAllowedResourceItemValue) ToResourceLink(ctx context.Context, valuePath path.Path) (cm.ResourceLink, diag.Diagnostics) {
	externalPath := valuePath.AtName("external")
	contentfulEntryPath := valuePath.AtName("contentful_entry")

	return ConvertExactlyOneKnownAlternative(
		valuePath,
		KnownUnionAlternative[cm.ResourceLink]{
			Name: "external", Path: externalPath, Value: v.External,
			Convert: func() (cm.ResourceLink, diag.Diagnostics) {
				external, _ := v.External.GetValue()
				resourceLink := cm.ResourceLink{}
				diags := external.SetResourceLink(ctx, externalPath, &resourceLink)

				return resourceLink, diags
			},
		},
		KnownUnionAlternative[cm.ResourceLink]{
			Name: "contentful_entry", Path: contentfulEntryPath, Value: v.ContentfulEntry,
			Convert: func() (cm.ResourceLink, diag.Diagnostics) {
				contentfulEntry, _ := v.ContentfulEntry.GetValue()
				resourceLink := cm.ResourceLink{}
				diags := contentfulEntry.SetResourceLink(ctx, contentfulEntryPath, &resourceLink)

				return resourceLink, diags
			},
		},
	)
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) SetResourceLink(_ context.Context, path path.Path, resourceLink *cm.ResourceLink) diag.Diagnostics {
	diags := diag.Diagnostics{}

	typeID, typeIDDiags := KnownStringValue(v.TypeID, path.AtName("type"))
	diags.Append(typeIDDiags...)

	resourceLink.Type = typeID
	resourceLink.Source = cm.OptString{}
	resourceLink.ContentTypes = nil

	return diags
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) SetResourceLink(_ context.Context, path path.Path, resourceLink *cm.ResourceLink) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if v.ContentTypes.IsNull() || v.ContentTypes.IsUnknown() {
		if v.ContentTypes.IsUnknown() {
			diags.AddAttributeError(path.AtName("content_types"), "Unexpected unknown content types", "Allowed content types must be known before they can be sent to Contentful.")
		} else {
			diags.AddAttributeError(path.AtName("content_types"), "Unexpected null content types", "Allowed content types are required.")
		}

		return diags
	}

	contentTypes, contentTypesDiags := KnownStringValues(v.ContentTypes.Elements(), path.AtName("content_types"))
	diags.Append(contentTypesDiags...)

	source, sourceDiags := KnownStringValue(v.Source, path.AtName("source"))
	diags.Append(sourceDiags...)

	if diags.HasError() {
		return diags
	}

	resourceLink.Type = contentfulEntryAllowedResourceType
	resourceLink.Source = cm.NewOptString(source)
	resourceLink.ContentTypes = contentTypes

	return diags
}
