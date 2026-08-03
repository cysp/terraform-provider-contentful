package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const contentfulEntryAllowedResourceType = "Contentful:Entry"

func (m *ContentTypeModel) ToContentTypeRequestData(ctx context.Context) (cm.ContentTypeRequestData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	name, nameDiags := contentTypeRequiredString(m.Name, path.Root("name"))
	diags.Append(nameDiags...)

	description := cm.NewOptNilStringNull()

	if m.Description.IsUnknown() {
		diags.AddAttributeError(
			path.Root("description"),
			"Unexpected unknown string",
			"The string value must be known before it can be sent to Contentful.",
		)
	} else if !m.Description.IsNull() {
		description = cm.NewOptNilString(m.Description.ValueString())
	}

	displayField, displayFieldDiags := contentTypeRequiredString(m.DisplayField, path.Root("display_field"))
	diags.Append(displayFieldDiags...)

	fields, fieldsDiags := FieldsListToContentTypeRequestDataFields(ctx, path.Root("fields"), m.Fields)
	diags.Append(fieldsDiags...)

	metadata, metadataDiags := ToOptContentTypeMetadata(ctx, path.Root("metadata"), m.Metadata)
	diags.Append(metadataDiags...)

	if diags.HasError() {
		return cm.ContentTypeRequestData{}, diags
	}

	return cm.ContentTypeRequestData{
		Name:         name,
		Description:  description,
		DisplayField: displayField,
		Fields:       fields,
		Metadata:     metadata,
	}, diags
}

func FieldsListToContentTypeRequestDataFields(
	ctx context.Context,
	valuePath path.Path,
	fieldsList TypedList[TypedObject[ContentTypeFieldValue]],
) ([]cm.ContentTypeRequestDataFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case fieldsList.IsUnknown():
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown content type fields",
			"Content type fields must be known before they can be sent to Contentful.",
		)
	case fieldsList.IsNull():
		diags.AddAttributeError(
			valuePath,
			"Unexpected null content type fields",
			"Content type fields are required.",
		)
	}

	if diags.HasError() {
		return nil, diags
	}

	return convertKnownObjectListElements(ctx, valuePath, fieldsList.Elements(), ToContentTypeRequestDataFieldsItem)
}

func ToContentTypeRequestDataFieldsItem(
	ctx context.Context,
	valuePath path.Path,
	v ContentTypeFieldValue,
) (cm.ContentTypeRequestDataFieldsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fieldID, idDiags := contentTypeRequiredString(v.ID, valuePath.AtName("id"))
	diags.Append(idDiags...)

	name, nameDiags := contentTypeRequiredString(v.Name, valuePath.AtName("name"))
	diags.Append(nameDiags...)

	fieldType, fieldTypeDiags := contentTypeRequiredString(v.FieldType, valuePath.AtName("type"))
	diags.Append(fieldTypeDiags...)

	localized, localizedDiags := contentTypeRequiredBool(v.Localized, valuePath.AtName("localized"))
	diags.Append(localizedDiags...)

	required, requiredDiags := contentTypeRequiredBool(v.Required, valuePath.AtName("required"))
	diags.Append(requiredDiags...)

	linkType, linkTypeDiags := contentTypeOptionalString(v.LinkType, valuePath.AtName("link_type"))
	diags.Append(linkTypeDiags...)

	disabled, disabledDiags := contentTypeOptionalBool(v.Disabled, valuePath.AtName("disabled"))
	diags.Append(disabledDiags...)

	omitted, omittedDiags := contentTypeOptionalBool(v.Omitted, valuePath.AtName("omitted"))
	diags.Append(omittedDiags...)

	items, itemsDiags := ItemsObjectToOptContentTypeRequestDataFieldsItemItems(valuePath.AtName("items"), v.Items)
	diags.Append(itemsDiags...)

	validations, validationsDiags := ValidationsListToContentTypeRequestDataFieldValidations(valuePath.AtName("validations"), v.Validations)
	diags.Append(validationsDiags...)

	defaultValue := jx.Raw(nil)

	if v.DefaultValue.IsUnknown() {
		diags.AddAttributeError(
			valuePath.AtName("default_value"),
			"Unexpected unknown default value",
			"The content type field default value must be known before it can be sent to Contentful.",
		)
	} else if !v.DefaultValue.IsNull() {
		defaultValue = jx.Raw(v.DefaultValue.ValueString())
	}

	allowedResources := cm.OptNilResourceLinkArray{}

	if v.AllowedResources.IsUnknown() {
		diags.AddAttributeError(
			valuePath.AtName("allowed_resources"),
			"Unexpected unknown allowed resources",
			"Allowed resources must be known before they can be sent to Contentful.",
		)
	} else if !v.AllowedResources.IsNull() {
		resourceLinks, resourceLinkDiags := AllowedResourceListToContentTypeRequestDataFieldAllowedResources(
			ctx,
			valuePath.AtName("allowed_resources"),
			v.AllowedResources,
		)
		diags.Append(resourceLinkDiags...)

		if !resourceLinkDiags.HasError() {
			allowedResources.SetTo(resourceLinks)
		}
	}

	if diags.HasError() {
		return cm.ContentTypeRequestDataFieldsItem{}, diags
	}

	return cm.ContentTypeRequestDataFieldsItem{
		ID:               fieldID,
		Name:             name,
		Type:             fieldType,
		LinkType:         linkType,
		Items:            items,
		Localized:        cm.NewOptBool(localized),
		Omitted:          omitted,
		Required:         cm.NewOptBool(required),
		Disabled:         disabled,
		DefaultValue:     defaultValue,
		Validations:      validations,
		AllowedResources: allowedResources,
	}, diags
}

func ItemsObjectToOptContentTypeRequestDataFieldsItemItems(
	valuePath path.Path,
	itemsObject TypedObject[ContentTypeFieldItemsValue],
) (cm.OptContentTypeRequestDataFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if itemsObject.IsUnknown() {
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown content type field items",
			"Content type field items must be known before they can be sent to Contentful.",
		)

		return cm.OptContentTypeRequestDataFieldsItemItems{}, diags
	}

	itemsValue, ok := itemsObject.GetValue()
	if !ok {
		return cm.OptContentTypeRequestDataFieldsItemItems{}, diags
	}

	items, itemsDiags := itemsValue.ToContentTypeRequestDataFieldsItemItems(valuePath)
	diags.Append(itemsDiags...)

	if diags.HasError() {
		return cm.OptContentTypeRequestDataFieldsItemItems{}, diags
	}

	return cm.NewOptContentTypeRequestDataFieldsItemItems(items), diags
}

func (v ContentTypeFieldItemsValue) ToContentTypeRequestDataFieldsItemItems(
	valuePath path.Path,
) (cm.ContentTypeRequestDataFieldsItemItems, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	itemsType, itemsTypeDiags := contentTypeRequiredString(v.ItemsType, valuePath.AtName("type"))
	diags.Append(itemsTypeDiags...)

	linkType, linkTypeDiags := contentTypeOptionalString(v.LinkType, valuePath.AtName("link_type"))
	diags.Append(linkTypeDiags...)

	validations, validationsDiags := ValidationsListToContentTypeRequestDataFieldValidations(valuePath.AtName("validations"), v.Validations)
	diags.Append(validationsDiags...)

	if diags.HasError() {
		return cm.ContentTypeRequestDataFieldsItemItems{}, diags
	}

	return cm.ContentTypeRequestDataFieldsItemItems{
		Type:        cm.NewOptString(itemsType),
		LinkType:    linkType,
		Validations: validations,
	}, diags
}

func ValidationsListToContentTypeRequestDataFieldValidations(
	valuePath path.Path,
	validationsList TypedList[jsontypes.Normalized],
) ([]jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if validationsList.IsUnknown() {
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown content type field validations",
			"Content type field validations must be known before they can be sent to Contentful.",
		)

		return nil, diags
	}

	if validationsList.IsNull() {
		return []jx.Raw{}, diags
	}

	validations := make([]jx.Raw, 0, len(validationsList.Elements()))
	for index, validation := range validationsList.Elements() {
		validationPath := valuePath.AtListIndex(index)

		switch {
		case validation.IsUnknown():
			diags.AddAttributeError(
				validationPath,
				"Unexpected unknown content type field validation",
				"The validation must be known before it can be sent to Contentful.",
			)
		case validation.IsNull():
			diags.AddAttributeError(
				validationPath,
				"Unexpected null content type field validation",
				"Null validations cannot be sent to Contentful.",
			)
		default:
			validations = append(validations, jx.Raw(validation.ValueString()))
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return validations, diags
}

func AllowedResourceListToContentTypeRequestDataFieldAllowedResources(
	ctx context.Context,
	valuePath path.Path,
	allowedResourcesList TypedList[TypedObject[ContentTypeFieldAllowedResourceItemValue]],
) ([]cm.ResourceLink, diag.Diagnostics) {
	if allowedResourcesList.IsUnknown() {
		return nil, diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
			valuePath,
			"Unexpected unknown allowed resources",
			"Allowed resources must be known before they can be sent to Contentful.",
		)}
	}

	if allowedResourcesList.IsNull() {
		return nil, nil
	}

	return convertKnownObjectListElements(
		ctx,
		valuePath,
		allowedResourcesList.Elements(),
		func(_ context.Context, resourcePath path.Path, value ContentTypeFieldAllowedResourceItemValue) (cm.ResourceLink, diag.Diagnostics) {
			return value.ToResourceLink(resourcePath)
		},
	)
}

func (v ContentTypeFieldAllowedResourceItemValue) ToResourceLink(
	valuePath path.Path,
) (cm.ResourceLink, diag.Diagnostics) {
	externalPath := valuePath.AtName("external")
	contentfulEntryPath := valuePath.AtName("contentful_entry")

	return convertExactlyOneKnownAlternative(
		valuePath,
		knownUnionAlternative[cm.ResourceLink]{
			Name:  "external",
			Path:  externalPath,
			Value: v.External,
			Convert: func() (cm.ResourceLink, diag.Diagnostics) {
				external, _ := v.External.GetValue()

				return external.ToResourceLink(externalPath)
			},
		},
		knownUnionAlternative[cm.ResourceLink]{
			Name:  "contentful_entry",
			Path:  contentfulEntryPath,
			Value: v.ContentfulEntry,
			Convert: func() (cm.ResourceLink, diag.Diagnostics) {
				contentfulEntry, _ := v.ContentfulEntry.GetValue()

				return contentfulEntry.ToResourceLink(contentfulEntryPath)
			},
		},
	)
}

func (v ContentTypeFieldAllowedResourceItemExternalValue) ToResourceLink(
	valuePath path.Path,
) (cm.ResourceLink, diag.Diagnostics) {
	typeID, diags := contentTypeRequiredString(v.TypeID, valuePath.AtName("type"))
	if diags.HasError() {
		return cm.ResourceLink{}, diags
	}

	return cm.ResourceLink{Type: typeID}, diags
}

func (v ContentTypeFieldAllowedResourceItemContentfulEntryValue) ToResourceLink(
	valuePath path.Path,
) (cm.ResourceLink, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	source, sourceDiags := contentTypeRequiredString(v.Source, valuePath.AtName("source"))
	diags.Append(sourceDiags...)

	contentTypesPath := valuePath.AtName("content_types")
	contentTypes := []string(nil)

	switch {
	case v.ContentTypes.IsUnknown():
		diags.AddAttributeError(
			contentTypesPath,
			"Unexpected unknown content types",
			"Allowed content types must be known before they can be sent to Contentful.",
		)
	case v.ContentTypes.IsNull():
		diags.AddAttributeError(
			contentTypesPath,
			"Unexpected null content types",
			"Allowed content types are required.",
		)
	default:
		var contentTypeDiags diag.Diagnostics

		contentTypes, contentTypeDiags = knownStringListElements(contentTypesPath, v.ContentTypes.Elements())
		diags.Append(contentTypeDiags...)
	}

	if diags.HasError() {
		return cm.ResourceLink{}, diags
	}

	return cm.ResourceLink{
		Type:         contentfulEntryAllowedResourceType,
		Source:       cm.NewOptString(source),
		ContentTypes: contentTypes,
	}, diags
}

func contentTypeRequiredString(value types.String, valuePath path.Path) (string, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown string",
			"The string value must be known before it can be sent to Contentful.",
		)
	case value.IsNull():
		diags.AddAttributeError(
			valuePath,
			"Unexpected null string",
			"The required string value cannot be null.",
		)
	default:
		return value.ValueString(), diags
	}

	return "", diags
}

func contentTypeRequiredBool(value types.Bool, valuePath path.Path) (bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown boolean",
			"The boolean value must be known before it can be sent to Contentful.",
		)
	case value.IsNull():
		diags.AddAttributeError(
			valuePath,
			"Unexpected null boolean",
			"The required boolean value cannot be null.",
		)
	default:
		return value.ValueBool(), diags
	}

	return false, diags
}

func contentTypeOptionalString(value types.String, valuePath path.Path) (cm.OptString, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown string",
			"The string value must be known before it can be sent to Contentful.",
		)

		return cm.OptString{}, diags
	}

	if value.IsNull() {
		return cm.OptString{}, diags
	}

	return cm.NewOptString(value.ValueString()), diags
}

func contentTypeOptionalBool(value types.Bool, valuePath path.Path) (cm.OptBool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown boolean",
			"The boolean value must be known before it can be sent to Contentful.",
		)

		return cm.OptBool{}, diags
	}

	if value.IsNull() {
		return cm.OptBool{}, diags
	}

	return cm.NewOptBool(value.ValueBool()), diags
}
