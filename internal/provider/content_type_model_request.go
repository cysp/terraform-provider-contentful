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

func (m *ContentTypeModel) ToContentTypeRequestFields(ctx context.Context) (cm.ContentTypeRequestFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	request := cm.ContentTypeRequestFields{
		Name:         m.Name.ValueString(),
		Description:  cm.NewOptNilPointerString(m.Description.ValueStringPointer()),
		DisplayField: m.DisplayField.ValueString(),
	}

	fields, fieldsDiags := FieldsListToContentTypeRequestFieldsFields(ctx, path.Root("fields"), m.Fields)
	diags.Append(fieldsDiags...)

	request.Fields = fields

	metadata, metadataDiags := ToOptContentTypeMetadata(ctx, path.Root("metadata"), m.Metadata)
	diags.Append(metadataDiags...)

	request.Metadata = metadata

	return request, diags
}

func FieldsListToContentTypeRequestFieldsFields(ctx context.Context, path path.Path, fieldsList TypedList[TypedObject[ContentTypeFieldValue]]) ([]cm.ContentTypeRequestFieldsFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsValues := fieldsList.Elements()

	fieldsItems := make([]cm.ContentTypeRequestFieldsFieldsItem, len(fieldsValues))

	for index, fieldsValue := range fieldsValues {
		fieldPath := path.AtListIndex(index)

		fieldsItem, fieldsItemDiags := ToContentTypeRequestFieldsFieldsItem(ctx, fieldPath, fieldsValue.Value())
		diags.Append(fieldsItemDiags...)

		fieldsItems[index] = fieldsItem
	}

	return fieldsItems, diags
}

func ToContentTypeRequestFieldsFieldsItem(ctx context.Context, path path.Path, v ContentTypeFieldValue) (cm.ContentTypeRequestFieldsFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsItemItems, fieldsItemItemsDiags := ItemsObjectToOptContentTypeRequestFieldsFieldsItemItems(ctx, path.AtName("items"), v.Items)
	diags.Append(fieldsItemItemsDiags...)

	fieldsItemValidations, fieldsItemValidationsDiags := ValidationsListToContentTypeRequestFieldsFieldValidations(ctx, path.AtName("validations"), v.Validations)
	diags.Append(fieldsItemValidationsDiags...)

	fieldsItem := cm.ContentTypeRequestFieldsFieldsItem{
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
		fieldsItemAllowedResources, fieldsItemAllowedResourcesDiags := AllowedResourceListToContentTypeRequestFieldsFieldAllowedResources(ctx, path.AtName("allowed_resources"), v.AllowedResources)
		diags.Append(fieldsItemAllowedResourcesDiags...)

		fieldsItem.AllowedResources.SetTo(fieldsItemAllowedResources)
	}

	return fieldsItem, diags
}

func ItemsObjectToOptContentTypeRequestFieldsFieldsItemItems(ctx context.Context, path path.Path, itemsObject TypedObject[ContentTypeFieldItemsValue]) (cm.OptContentTypeRequestFieldsFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsItemItems := cm.OptContentTypeRequestFieldsFieldsItemItems{}

	itemsValue, itemsValueOk := itemsObject.GetValue()
	if itemsValueOk {
		items, itemsDiags := itemsValue.ToContentTypeRequestFieldsFieldsItemItems(ctx, path)
		diags.Append(itemsDiags...)

		fieldsItemItems.SetTo(items)
	}

	return fieldsItemItems, diags
}

func (v ContentTypeFieldItemsValue) ToContentTypeRequestFieldsFieldsItemItems(ctx context.Context, path path.Path) (cm.ContentTypeRequestFieldsFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	itemsValidations, itemsValidationsDiags := ValidationsListToContentTypeRequestFieldsFieldValidations(ctx, path.AtName("validations"), v.Validations)
	diags.Append(itemsValidationsDiags...)

	items := cm.ContentTypeRequestFieldsFieldsItemItems{
		Type:        util.StringValueToOptString(v.ItemsType),
		LinkType:    util.StringValueToOptString(v.LinkType),
		Validations: itemsValidations,
	}

	return items, diags
}

func ValidationsListToContentTypeRequestFieldsFieldValidations(ctx context.Context, _ path.Path, validationsList TypedList[jsontypes.Normalized]) ([]jx.Raw, diag.Diagnostics) {
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

func AllowedResourceListToContentTypeRequestFieldsFieldAllowedResources(ctx context.Context, path path.Path, allowedResourcesList TypedList[TypedObject[ContentTypeFieldAllowedResourceItemValue]]) ([]cm.ResourceLink, diag.Diagnostics) {
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
