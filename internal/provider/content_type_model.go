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

func (m *ContentTypeModel) ToPutContentTypeReq(ctx context.Context) (cm.PutContentTypeReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	request := cm.PutContentTypeReq{
		Name:         m.Name.ValueString(),
		Description:  m.Description.ValueString(),
		DisplayField: m.DisplayField.ValueString(),
	}

	fields, fieldsDiags := FieldsListToPutContentTypeReqFields(ctx, path.Root("fields"), m.Fields)
	diags.Append(fieldsDiags...)

	request.Fields = fields

	return request, diags
}

func FieldsListToPutContentTypeReqFields(ctx context.Context, path path.Path, fieldsList types.List) ([]cm.PutContentTypeReqFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsValues := make([]FieldsValue, len(fieldsList.Elements()))
	diags.Append(fieldsList.ElementsAs(ctx, &fieldsValues, false)...)

	fieldsItems := make([]cm.PutContentTypeReqFieldsItem, len(fieldsValues))

	for index, fieldsValue := range fieldsValues {
		path := path.AtListIndex(index)
		fieldsItem, fieldsItemDiags := fieldsValue.ToPutContentTypeReqFieldsItem(ctx, path)
		diags.Append(fieldsItemDiags...)

		fieldsItems[index] = fieldsItem
	}

	return fieldsItems, diags
}

func (model *FieldsValue) ToPutContentTypeReqFieldsItem(ctx context.Context, path path.Path) (cm.PutContentTypeReqFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsItemItems, fieldsItemItemsDiags := ItemsObjectToOptPutContentTypeReqFieldsItemItems(ctx, path.AtName("items"), model.Items)
	diags.Append(fieldsItemItemsDiags...)

	fieldsItemValidations, fieldsItemValidationsDiags := ValidationsListToPutContentTypeReqValidations(ctx, path.AtName("validations"), model.Validations)
	diags.Append(fieldsItemValidationsDiags...)

	fieldsItem := cm.PutContentTypeReqFieldsItem{
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

func ItemsObjectToOptPutContentTypeReqFieldsItemItems(ctx context.Context, path path.Path, itemsObject types.Object) (cm.OptPutContentTypeReqFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsItemItems := cm.OptPutContentTypeReqFieldsItemItems{}

	if !itemsObject.IsNull() && !itemsObject.IsUnknown() {
		modelItemsObjectValue, modelItemsObjectValueDiags := ItemsType{}.ValueFromObject(ctx, itemsObject)
		diags.Append(modelItemsObjectValueDiags...)

		if modelItemsValue, ok := modelItemsObjectValue.(ItemsValue); ok {
			items, itemsDiags := modelItemsValue.ToPutContentTypeReqFieldsItemItems(ctx, path)
			diags.Append(itemsDiags...)

			fieldsItemItems.SetTo(items)
		} else {
			diags.AddAttributeError(path, "Failed to convert to ItemsValue", "Failed to convert to ItemsValue")
		}
	}

	return fieldsItemItems, diags
}

func (model *ItemsValue) ToPutContentTypeReqFieldsItemItems(ctx context.Context, path path.Path) (cm.PutContentTypeReqFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	itemsValidations, itemsValidationsDiags := ValidationsListToPutContentTypeReqValidations(ctx, path.AtName("validations"), model.Validations)
	diags.Append(itemsValidationsDiags...)

	items := cm.PutContentTypeReqFieldsItemItems{
		Type:        util.StringValueToOptString(model.ItemsType),
		LinkType:    util.StringValueToOptString(model.LinkType),
		Validations: itemsValidations,
	}

	return items, diags
}

func ValidationsListToPutContentTypeReqValidations(ctx context.Context, _ path.Path, validationsList types.List) ([]jx.Raw, diag.Diagnostics) {
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
