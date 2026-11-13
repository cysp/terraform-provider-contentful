package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/cysp/terraform-provider-contentful/internal/provider/testdata"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

func TestContentTypeFieldRequestProperties(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	rapid.Check(t, func(t *rapid.T) {
		field := testdata.ContentTypeFieldValue("field").Draw(t, "field")

		request, diags := ToContentTypeRequestDataFieldsItem(ctx, path.Root("fields").AtListIndex(0), field)
		if diags.HasError() {
			t.Fatalf("ToContentTypeRequestDataFieldsItem failed: %v", diags.Errors())
		}

		// Check the request before a response converter can undo a symmetric
		// fault. These expectations never use the mock or response conversion.
		assert.Equal(t, field.ID.ValueString(), request.ID)
		assert.Equal(t, field.Name.ValueString(), request.Name)
		assert.Equal(t, field.FieldType.ValueString(), request.Type)
		assert.Equal(t, cm.NewOptBool(field.Localized.ValueBool()), request.Localized)
		assert.Equal(t, cm.NewOptBool(field.Required.ValueBool()), request.Required)
		assert.Equal(t, cm.NewOptBool(field.Disabled.ValueBool()), request.Disabled)
		assert.Equal(t, cm.NewOptBool(field.Omitted.ValueBool()), request.Omitted)
		assert.Equal(t, !field.LinkType.IsNull(), request.LinkType.IsSet())
		assert.Equal(t, field.LinkType.ValueString(), request.LinkType.Value)
		assert.Equal(t, !field.Items.IsNull(), request.Items.IsSet())

		if items, present := field.Items.GetValue(); present {
			assert.Equal(t, cm.NewOptString(items.ItemsType.ValueString()), request.Items.Value.Type)
			assert.Equal(t, !items.LinkType.IsNull(), request.Items.Value.LinkType.IsSet())
			assert.Equal(t, items.LinkType.ValueString(), request.Items.Value.LinkType.Value)

			expectedValidations := make([]jx.Raw, len(items.Validations.Elements()))
			for index, validation := range items.Validations.Elements() {
				expectedValidations[index] = jx.Raw(validation.ValueString())
			}

			assert.Equal(t, expectedValidations, request.Items.Value.Validations)
		}

		expectedDefault := jx.Raw(nil)
		if !field.DefaultValue.IsNull() {
			expectedDefault = jx.Raw(field.DefaultValue.ValueString())
		}

		assert.Equal(t, expectedDefault, request.DefaultValue)

		expectedValidations := make([]jx.Raw, len(field.Validations.Elements()))
		for index, validation := range field.Validations.Elements() {
			expectedValidations[index] = jx.Raw(validation.ValueString())
		}

		assert.Equal(t, expectedValidations, request.Validations)

		expectedAllowedResources := cm.OptNilResourceLinkArray{}

		if !field.AllowedResources.IsNull() {
			expectedResources := make([]cm.ResourceLink, len(field.AllowedResources.Elements()))
			for index, resource := range field.AllowedResources.Elements() {
				if entry, present := resource.Value().ContentfulEntry.GetValue(); present {
					contentTypes := make([]string, len(entry.ContentTypes.Elements()))
					for contentTypeIndex, contentType := range entry.ContentTypes.Elements() {
						contentTypes[contentTypeIndex] = contentType.ValueString()
					}

					expectedResources[index] = cm.ResourceLink{
						Type: "Contentful:Entry", Source: cm.NewOptString(entry.Source.ValueString()), ContentTypes: contentTypes,
					}
				} else {
					expectedResources[index] = cm.ResourceLink{Type: resource.Value().External.Value().TypeID.ValueString()}
				}
			}

			expectedAllowedResources = cm.NewOptNilResourceLinkArray(expectedResources)
		}

		assert.Equal(t, expectedAllowedResources, request.AllowedResources)
	})
}

func TestContentTypeModelRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	rapid.Check(t, func(t *rapid.T) {
		spaceID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "spaceID")
		environmentID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "environmentID")
		contentTypeID := testdata.AlphanumericStringOfN(1, 10).Draw(t, "contentTypeID")

		model := testdata.ContentTypeModel(spaceID, environmentID, contentTypeID).Draw(t, "model")

		request, diags := model.ToContentTypeRequestData(ctx, ContentTypeModel{})
		if diags.HasError() {
			t.Fatalf("ToContentTypeRequestData failed: %v", diags.Errors())
		}

		contentType := cmt.NewContentTypeFromRequestFields(spaceID, environmentID, contentTypeID, request)

		result, diags := NewContentTypeResourceModelFromResponse(ctx, contentType)
		if diags.HasError() {
			t.Fatalf("NewContentTypeResourceModelFromResponse failed: %v", diags.Errors())
		}

		assert.Equal(t, model, result)
	})
}
