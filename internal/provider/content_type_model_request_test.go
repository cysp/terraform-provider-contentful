package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestContentTypeRequestRejectsUnresolvedRequiredValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mutate       func(*ContentTypeModel)
		expectedPath string
	}{
		"unknown name": {
			mutate:       func(model *ContentTypeModel) { model.Name = types.StringUnknown() },
			expectedPath: "name",
		},
		"null name": {
			mutate:       func(model *ContentTypeModel) { model.Name = types.StringNull() },
			expectedPath: "name",
		},
		"unknown description": {
			mutate:       func(model *ContentTypeModel) { model.Description = types.StringUnknown() },
			expectedPath: "description",
		},
		"null description": {
			mutate:       func(model *ContentTypeModel) { model.Description = types.StringNull() },
			expectedPath: "description",
		},
		"unknown display field": {
			mutate:       func(model *ContentTypeModel) { model.DisplayField = types.StringUnknown() },
			expectedPath: "display_field",
		},
		"null display field": {
			mutate:       func(model *ContentTypeModel) { model.DisplayField = types.StringNull() },
			expectedPath: "display_field",
		},
		"unknown fields": {
			mutate: func(model *ContentTypeModel) {
				model.Fields = NewTypedListUnknown[TypedObject[ContentTypeFieldValue]]()
			},
			expectedPath: "fields",
		},
		"null fields": {
			mutate: func(model *ContentTypeModel) {
				model.Fields = NewTypedListNull[TypedObject[ContentTypeFieldValue]]()
			},
			expectedPath: "fields",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := validContentTypeRequestModel()
			test.mutate(&model)

			actual, diags := model.ToContentTypeRequestData(t.Context())

			assert.Equal(t, cm.ContentTypeRequestData{}, actual)
			assert.Equal(t, []string{test.expectedPath}, diagnosticPaths(t, diags))
		})
	}
}

func TestContentTypeFieldsRejectUnresolvedElementsWithoutPartialOutput(t *testing.T) {
	t.Parallel()

	result, diags := FieldsListToContentTypeRequestDataFields(
		t.Context(),
		path.Root("fields"),
		NewTypedList([]TypedObject[ContentTypeFieldValue]{
			NewTypedObject(validContentTypeFieldValue()),
			NewTypedObjectNull[ContentTypeFieldValue](),
			NewTypedObjectUnknown[ContentTypeFieldValue](),
		}),
	)

	assert.Nil(t, result)
	assert.Equal(t, []string{"fields[1]", "fields[2]"}, diagnosticPaths(t, diags))
}

func TestContentTypeFieldRejectsUnresolvedRequiredValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mutate       func(*ContentTypeFieldValue)
		expectedPath string
	}{
		"unknown id": {
			mutate:       func(value *ContentTypeFieldValue) { value.ID = types.StringUnknown() },
			expectedPath: "fields[0].id",
		},
		"null id": {
			mutate:       func(value *ContentTypeFieldValue) { value.ID = types.StringNull() },
			expectedPath: "fields[0].id",
		},
		"unknown name": {
			mutate:       func(value *ContentTypeFieldValue) { value.Name = types.StringUnknown() },
			expectedPath: "fields[0].name",
		},
		"null name": {
			mutate:       func(value *ContentTypeFieldValue) { value.Name = types.StringNull() },
			expectedPath: "fields[0].name",
		},
		"unknown type": {
			mutate:       func(value *ContentTypeFieldValue) { value.FieldType = types.StringUnknown() },
			expectedPath: "fields[0].type",
		},
		"null type": {
			mutate:       func(value *ContentTypeFieldValue) { value.FieldType = types.StringNull() },
			expectedPath: "fields[0].type",
		},
		"unknown localized": {
			mutate:       func(value *ContentTypeFieldValue) { value.Localized = types.BoolUnknown() },
			expectedPath: "fields[0].localized",
		},
		"null localized": {
			mutate:       func(value *ContentTypeFieldValue) { value.Localized = types.BoolNull() },
			expectedPath: "fields[0].localized",
		},
		"unknown required": {
			mutate:       func(value *ContentTypeFieldValue) { value.Required = types.BoolUnknown() },
			expectedPath: "fields[0].required",
		},
		"null required": {
			mutate:       func(value *ContentTypeFieldValue) { value.Required = types.BoolNull() },
			expectedPath: "fields[0].required",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value := validContentTypeFieldValue()
			test.mutate(&value)

			actual, diags := ToContentTypeRequestDataFieldsItem(t.Context(), path.Root("fields").AtListIndex(0), value)

			assert.Equal(t, cm.ContentTypeRequestDataFieldsItem{}, actual)
			assert.Equal(t, []string{test.expectedPath}, diagnosticPaths(t, diags))
		})
	}
}

func TestContentTypeFieldRejectsUnresolvedOptionalValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mutate       func(*ContentTypeFieldValue)
		expectedPath string
	}{
		"link type": {
			mutate:       func(value *ContentTypeFieldValue) { value.LinkType = types.StringUnknown() },
			expectedPath: "fields[0].link_type",
		},
		"disabled": {
			mutate:       func(value *ContentTypeFieldValue) { value.Disabled = types.BoolUnknown() },
			expectedPath: "fields[0].disabled",
		},
		"omitted": {
			mutate:       func(value *ContentTypeFieldValue) { value.Omitted = types.BoolUnknown() },
			expectedPath: "fields[0].omitted",
		},
		"items": {
			mutate: func(value *ContentTypeFieldValue) {
				value.Items = NewTypedObjectUnknown[ContentTypeFieldItemsValue]()
			},
			expectedPath: "fields[0].items",
		},
		"default value": {
			mutate:       func(value *ContentTypeFieldValue) { value.DefaultValue = jsontypes.NewNormalizedUnknown() },
			expectedPath: "fields[0].default_value",
		},
		"validations": {
			mutate: func(value *ContentTypeFieldValue) {
				value.Validations = NewTypedListUnknown[jsontypes.Normalized]()
			},
			expectedPath: "fields[0].validations",
		},
		"allowed resources": {
			mutate: func(value *ContentTypeFieldValue) {
				value.AllowedResources = NewTypedListUnknown[TypedObject[ContentTypeFieldAllowedResourceItemValue]]()
			},
			expectedPath: "fields[0].allowed_resources",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value := validContentTypeFieldValue()
			test.mutate(&value)

			actual, diags := ToContentTypeRequestDataFieldsItem(t.Context(), path.Root("fields").AtListIndex(0), value)

			assert.Equal(t, cm.ContentTypeRequestDataFieldsItem{}, actual)
			assert.Equal(t, []string{test.expectedPath}, diagnosticPaths(t, diags))
		})
	}
}

func TestContentTypeFieldItemsRejectUnresolvedValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value        ContentTypeFieldItemsValue
		expectedPath string
	}{
		"unknown type": {
			value: ContentTypeFieldItemsValue{
				ItemsType:   types.StringUnknown(),
				LinkType:    types.StringNull(),
				Validations: NewTypedList([]jsontypes.Normalized{}),
			},
			expectedPath: "fields[0].items.type",
		},
		"null type": {
			value: ContentTypeFieldItemsValue{
				ItemsType:   types.StringNull(),
				LinkType:    types.StringNull(),
				Validations: NewTypedList([]jsontypes.Normalized{}),
			},
			expectedPath: "fields[0].items.type",
		},
		"unknown link type": {
			value: ContentTypeFieldItemsValue{
				ItemsType:   types.StringValue("Link"),
				LinkType:    types.StringUnknown(),
				Validations: NewTypedList([]jsontypes.Normalized{}),
			},
			expectedPath: "fields[0].items.link_type",
		},
		"unknown validations": {
			value: ContentTypeFieldItemsValue{
				ItemsType:   types.StringValue("Link"),
				LinkType:    types.StringNull(),
				Validations: NewTypedListUnknown[jsontypes.Normalized](),
			},
			expectedPath: "fields[0].items.validations",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := test.value.ToContentTypeRequestDataFieldsItemItems(
				t.Context(),
				path.Root("fields").AtListIndex(0).AtName("items"),
			)

			assert.Equal(t, cm.ContentTypeRequestDataFieldsItemItems{}, actual)
			assert.Equal(t, []string{test.expectedPath}, diagnosticPaths(t, diags))
		})
	}
}

func TestContentTypeValidationsFailClosedWithExactPaths(t *testing.T) {
	t.Parallel()

	actual, diags := ValidationsListToContentTypeRequestDataFieldValidations(
		t.Context(),
		path.Root("fields").AtListIndex(0).AtName("validations"),
		NewTypedList([]jsontypes.Normalized{
			jsontypes.NewNormalizedValue(`{"size":{"min":1}}`),
			jsontypes.NewNormalizedNull(),
			jsontypes.NewNormalizedUnknown(),
		}),
	)

	assert.Nil(t, actual)
	assert.Equal(t, []string{"fields[0].validations[1]", "fields[0].validations[2]"}, diagnosticPaths(t, diags))
}

func validContentTypeRequestModel() ContentTypeModel {
	return ContentTypeModel{
		Name:         types.StringValue("Test"),
		Description:  types.StringValue("A test content type"),
		DisplayField: types.StringValue("title"),
		Fields: NewTypedList([]TypedObject[ContentTypeFieldValue]{
			NewTypedObject(validContentTypeFieldValue()),
		}),
		Metadata: NewTypedObjectNull[ContentTypeMetadataValue](),
	}
}

func validContentTypeFieldValue() ContentTypeFieldValue {
	return ContentTypeFieldValue{
		ID:               types.StringValue("title"),
		Name:             types.StringValue("Title"),
		FieldType:        types.StringValue("Symbol"),
		LinkType:         types.StringNull(),
		Disabled:         types.BoolValue(false),
		Omitted:          types.BoolValue(false),
		Required:         types.BoolValue(true),
		DefaultValue:     jsontypes.NewNormalizedNull(),
		Items:            NewTypedObjectNull[ContentTypeFieldItemsValue](),
		Localized:        types.BoolValue(false),
		Validations:      NewTypedList([]jsontypes.Normalized{}),
		AllowedResources: NewTypedListNull[TypedObject[ContentTypeFieldAllowedResourceItemValue]](),
	}
}
