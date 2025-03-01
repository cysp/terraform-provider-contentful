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
		Description:  m.Description.ValueString(),
		DisplayField: m.DisplayField.ValueString(),
	}

	fields, fieldsDiags := FieldsListToContentTypeRequestFieldsFields(ctx, path.Root("fields"), m.Fields)
	diags.Append(fieldsDiags...)

	request.Fields = fields

	return request, diags
}

func FieldsListToContentTypeRequestFieldsFields(ctx context.Context, path path.Path, fieldsList types.List) ([]cm.ContentTypeRequestFieldsFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsValues := make([]FieldsValue, len(fieldsList.Elements()))
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

func (model *FieldsValue) ToContentTypeRequestFieldsFieldsItem(ctx context.Context, path path.Path) (cm.ContentTypeRequestFieldsFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsItemItems, fieldsItemItemsDiags := ItemsObjectToOptContentTypeRequestFieldsFieldsItemItems(ctx, path.AtName("items"), model.Items)
	diags.Append(fieldsItemItemsDiags...)

	fieldsItemValidations, fieldsItemValidationsDiags := ValidationsListToContentTypeRequestFieldsFieldValidations(ctx, path.AtName("validations"), model.Validations)
	diags.Append(fieldsItemValidationsDiags...)

	fieldsItem := cm.ContentTypeRequestFieldsFieldsItem{
		ID:          model.Id.ValueString(),
		Name:        model.Name.ValueString(),
		Type:        model.FieldsType.ValueString(),
		LinkType:    util.StringValueToOptString(model.LinkType),
		Items:       fieldsItemItems,
		Validations: fieldsItemValidations,
		Disabled:    util.BoolValueToOptBool(model.Disabled),
		Omitted:     util.BoolValueToOptBool(model.Omitted),
		Required:    util.BoolValueToOptBool(model.Required),
		Localized:   util.BoolValueToOptBool(model.Localized),
	}

	modelDefaultValueValue := model.DefaultValue.ValueString()
	if modelDefaultValueValue != "" {
		fieldsItem.DefaultValue = []byte(modelDefaultValueValue)
	}

	return fieldsItem, diags
}

func ItemsObjectToOptContentTypeRequestFieldsFieldsItemItems(ctx context.Context, path path.Path, itemsObject types.Object) (cm.OptContentTypeRequestFieldsFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsItemItems := cm.OptContentTypeRequestFieldsFieldsItemItems{}

	if !itemsObject.IsNull() && !itemsObject.IsUnknown() {
		modelItemsObjectValue, modelItemsObjectValueDiags := ItemsType{}.ValueFromObject(ctx, itemsObject)
		diags.Append(modelItemsObjectValueDiags...)

		if modelItemsValue, ok := modelItemsObjectValue.(ItemsValue); ok {
			items, itemsDiags := modelItemsValue.ToContentTypeRequestFieldsFieldsItemItems(ctx, path)
			diags.Append(itemsDiags...)

			fieldsItemItems.SetTo(items)
		} else {
			diags.AddAttributeError(path, "Failed to convert to ItemsValue", "Failed to convert to ItemsValue")
		}
	}

	return fieldsItemItems, diags
}

func (model *ItemsValue) ToContentTypeRequestFieldsFieldsItemItems(ctx context.Context, path path.Path) (cm.ContentTypeRequestFieldsFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	itemsValidations, itemsValidationsDiags := ValidationsListToContentTypeRequestFieldsFieldValidations(ctx, path.AtName("validations"), model.Validations)
	diags.Append(itemsValidationsDiags...)

	items := cm.ContentTypeRequestFieldsFieldsItemItems{
		Type:        util.StringValueToOptString(model.ItemsType),
		LinkType:    util.StringValueToOptString(model.LinkType),
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

func (m *ContentTypeModel) ReadFromResponse(ctx context.Context, contentType *cm.ContentType) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// SpaceId, EnvironmentId and ContentTypeId are all already known

	m.Name = types.StringValue(contentType.Name)
	m.Description = types.StringValue(contentType.Description.Or(""))
	m.DisplayField = types.StringValue(contentType.DisplayField.Or(""))

	fieldsList, fieldsListDiags := NewFieldsListFromResponse(ctx, path.Root("fields"), contentType.Fields)
	diags.Append(fieldsListDiags...)

	m.Fields = fieldsList

	return diags
}
