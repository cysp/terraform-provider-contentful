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

func (m *ContentTypeResourceModel) ToContentTypeRequestFields(ctx context.Context) (cm.ContentTypeRequestFields, diag.Diagnostics) {
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

func FieldsListToContentTypeRequestFieldsFields(ctx context.Context, path path.Path, fieldsList TypedList[ContentTypeFieldValue]) ([]cm.ContentTypeRequestFieldsFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsValues := fieldsList.Elements()

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

func ItemsObjectToOptContentTypeRequestFieldsFieldsItemItems(ctx context.Context, path path.Path, itemsObject ContentTypeFieldItemsValue) (cm.OptContentTypeRequestFieldsFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldsItemItems := cm.OptContentTypeRequestFieldsFieldsItemItems{}

	if !itemsObject.IsNull() && !itemsObject.IsUnknown() {
		items, itemsDiags := itemsObject.ToContentTypeRequestFieldsFieldsItemItems(ctx, path)
		diags.Append(itemsDiags...)

		fieldsItemItems.SetTo(items)
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
