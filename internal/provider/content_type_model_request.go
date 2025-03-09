package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

	return request, diags
}

func FieldsListToContentTypeRequestFieldsFields(ctx context.Context, path path.Path, fieldsList types.List) ([]cm.ContentTypeRequestFieldsFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsValues := make([]ContentTypeFieldValue, len(fieldsList.Elements()))
	diags.Append(fieldsList.ElementsAs(ctx, &fieldsValues, false)...)

	fieldsItems := make([]cm.ContentTypeRequestFieldsFieldsItem, len(fieldsValues))

	for index, fieldsValue := range fieldsValues {
		path := path.AtListIndex(index)
		fieldsItem, fieldsItemDiags := fieldsValue.ToContentTypeRequestFieldsFieldsItem(ctx, path)
		diags.Append(fieldsItemDiags...)

		fieldsItems[index] = fieldsItem
	}

	return fieldsItems, diags
}

func (v *ContentTypeFieldValue) ToContentTypeRequestFieldsFieldsItem(ctx context.Context, path path.Path) (cm.ContentTypeRequestFieldsFieldsItem, diag.Diagnostics) {
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

	return fieldsItem, diags
}

func ItemsObjectToOptContentTypeRequestFieldsFieldsItemItems(ctx context.Context, path path.Path, itemsObject types.Object) (cm.OptContentTypeRequestFieldsFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsItemItems := cm.OptContentTypeRequestFieldsFieldsItemItems{}

	if !itemsObject.IsNull() && !itemsObject.IsUnknown() {
		modelItemsObjectValue, modelItemsObjectValueDiags := ContentTypeFieldItemsType{}.ValueFromObject(ctx, itemsObject)
		diags.Append(modelItemsObjectValueDiags...)

		if modelItemsValue, ok := modelItemsObjectValue.(ContentTypeFieldItemsValue); ok {
			items, itemsDiags := modelItemsValue.ToContentTypeRequestFieldsFieldsItemItems(ctx, path)
			diags.Append(itemsDiags...)

			fieldsItemItems.SetTo(items)
		} else {
			diags.AddAttributeError(path, "Failed to convert to ContentTypeFieldItemsValue", "Failed to convert to ContentTypeFieldItemsValue")
		}
	}

	return fieldsItemItems, diags
}

func (v *ContentTypeFieldItemsValue) ToContentTypeRequestFieldsFieldsItemItems(ctx context.Context, path path.Path) (cm.ContentTypeRequestFieldsFieldsItemItems, diag.Diagnostics) {
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

func ValidationsListToContentTypeRequestFieldsFieldValidations(ctx context.Context, _ path.Path, validationsList types.List) ([]jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	validationsStrings := []string{}
	if !validationsList.IsNull() && !validationsList.IsUnknown() {
		diags.Append(validationsList.ElementsAs(ctx, &validationsStrings, false)...)
	}

	validations := make([]jx.Raw, len(validationsStrings))
	for index, validationsString := range validationsStrings {
		validations[index] = jx.Raw(validationsString)
	}

	return validations, diags
}
