package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

	return fieldsItem, diags
}

func ItemsObjectToOptContentTypeRequestDataFieldsItemItems(ctx context.Context, path path.Path, itemsObject TypedObject[ContentTypeFieldItemsValue]) (cm.OptContentTypeRequestDataFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsItemItems := cm.OptContentTypeRequestDataFieldsItemItems{}

	itemsValue, itemsValueOk := itemsObject.GetValue()
	if itemsValueOk {
		items, itemsDiags := itemsValue.ToContentTypeRequestDataFieldsItemItems(ctx, path)
		diags.Append(itemsDiags...)

		fieldsItemItems.SetTo(items)
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

	return items, diags
}

func ValidationsListToContentTypeRequestDataFieldValidations(ctx context.Context, _ path.Path, validationsList TypedList[jsontypes.Normalized]) ([]jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	validationsStrings := []string{}
	if !validationsList.IsNull() && !validationsList.IsUnknown() {
		diags.Append(tfsdk.ValueAs(ctx, validationsList, &validationsStrings)...)
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

func (v ContentTypeFieldAllowedResourceItemValue) ToResourceLink(ctx context.Context, path path.Path) (cm.ResourceLink, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	resourceLink := cm.ResourceLink{}
	found := false

	if external, ok := v.External.GetValue(); ok {
		found = true

		diags.Append(external.SetResourceLink(ctx, path.AtName("external"), &resourceLink)...)
	}

	if contentfulEntry, ok := v.ContentfulEntry.GetValue(); ok {
		found = true

		diags.Append(contentfulEntry.SetResourceLink(ctx, path.AtName("contentful_entry"), &resourceLink)...)
	}

	if !found {
		diags.AddAttributeError(path, "Missing allowed resource type", "Exactly one external or Contentful entry resource type must be known and non-null.")
	}

	return resourceLink, diags
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

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) SetResourceLink(ctx context.Context, path path.Path, resourceLink *cm.ResourceLink) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if v.ContentTypes.IsNull() || v.ContentTypes.IsUnknown() {
		if v.ContentTypes.IsUnknown() {
			diags.AddAttributeError(path.AtName("content_types"), "Unexpected unknown content types", "Allowed content types must be known before they can be sent to Contentful.")
		} else {
			diags.AddAttributeError(path.AtName("content_types"), "Unexpected null content types", "Allowed content types are required.")
		}
	}

	contentTypes := make([]string, len(v.ContentTypes.Elements()))
	diags.Append(tfsdk.ValueAs(ctx, v.ContentTypes, &contentTypes)...)

	source, sourceDiags := KnownStringValue(v.Source, path.AtName("source"))
	diags.Append(sourceDiags...)

	resourceLink.Type = contentfulEntryAllowedResourceType
	resourceLink.Source = cm.NewOptString(source)
	resourceLink.ContentTypes = contentTypes

	return diags
}
