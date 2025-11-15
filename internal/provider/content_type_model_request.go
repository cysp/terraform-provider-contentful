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

	return request, diags
}

func FieldsListToContentTypeRequestDataFields(ctx context.Context, path path.Path, fieldsList TypedList[TypedObject[ContentTypeFieldValue]]) ([]cm.ContentTypeRequestDataFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsValues := fieldsList.Elements()

	fieldsItems := make([]cm.ContentTypeRequestDataFieldsItem, len(fieldsValues))

	for index, fieldsValue := range fieldsValues {
		path := path.AtListIndex(index)

		fieldsItem, fieldsItemDiags := ToContentTypeRequestDataFieldsItem(ctx, path, fieldsValue.Value())
		diags.Append(fieldsItemDiags...)

		fieldsItems[index] = fieldsItem
	}

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

	if !v.AllowedResources.IsUnknown() && !v.AllowedResources.IsNull() {
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

func AllowedResourceListToContentTypeRequestDataFieldAllowedResources(ctx context.Context, path path.Path, allowedResourcesList TypedList[TypedObject[ContentTypeFieldAllowedResourceItemValue]]) ([]cm.ResourceLink, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	allowedResources := allowedResourcesList.Elements()
	resourceLinks := make([]cm.ResourceLink, len(allowedResources))

	for index, resource := range allowedResources {
		resourceLink, resourceDiags := resource.Value().ToResourceLink(ctx, path.AtListIndex(index))
		diags.Append(resourceDiags...)

		resourceLinks[index] = resourceLink
	}

	return resourceLinks, diags
}

func (v ContentTypeFieldAllowedResourceItemValue) ToResourceLink(ctx context.Context, path path.Path) (cm.ResourceLink, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	resourceLink := cm.ResourceLink{}

	if !v.External.IsNull() && !v.External.IsUnknown() {
		diags.Append(v.External.Value().SetResourceLink(ctx, path, &resourceLink)...)
	}

	if !v.ContentfulEntry.IsNull() && !v.ContentfulEntry.IsUnknown() {
		diags.Append(v.ContentfulEntry.Value().SetResourceLink(ctx, path, &resourceLink)...)
	}

	return resourceLink, diags
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) SetResourceLink(_ context.Context, _ path.Path, resourceLink *cm.ResourceLink) diag.Diagnostics {
	diags := diag.Diagnostics{}

	resourceLink.SetExternalResourceLink(cm.ExternalResourceLink{
		Type: v.TypeID.ValueString(),
	})

	return diags
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) SetResourceLink(ctx context.Context, _ path.Path, resourceLink *cm.ResourceLink) diag.Diagnostics {
	diags := diag.Diagnostics{}

	contentTypes := make([]string, len(v.ContentTypes.Elements()))
	diags.Append(tfsdk.ValueAs(ctx, v.ContentTypes, &contentTypes)...)

	resourceLink.SetContentfulEntryResourceLink(cm.ContentfulEntryResourceLink{
		Type:         cm.ContentfulEntryResourceLinkTypeContentfulEntry,
		Source:       v.Source.ValueString(),
		ContentTypes: contentTypes,
	})

	return diags
}
