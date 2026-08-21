package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *AppDefinitionBaseModel) ToAppDefinitionData(config AppDefinitionBaseModel, path path.Path) (cm.AppDefinitionData, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	diags.Append(rejectUnknownConfigurationOwnedRequestValue(model.Src, config.Src, path.AtName("src"))...)
	diags.Append(rejectUnknownConfigurationOwnedRequestValue(model.BundleID, config.BundleID, path.AtName("bundle_id"))...)

	name, nameDiags := appRequestRequiredString(model.Name, path.AtName("name"))
	diags.Append(nameDiags...)

	fields := cm.AppDefinitionData{
		Name: name,
	}

	// An unknown Optional+Computed plan is omitted only when configuration is
	// null, meaning the value is response-owned. Every known plan value remains
	// the request source, including defaults and values preserved from state.
	if !model.Src.IsUnknown() && !model.Src.IsNull() {
		fields.Src = cm.NewOptPointerString(model.Src.ValueStringPointer())
	}

	if !model.BundleID.IsNull() && !model.BundleID.IsUnknown() {
		fields.Bundle.SetTo(cm.NewAppBundleLink(model.BundleID.ValueString()))
	}

	if model.Locations != nil {
		path := path.AtName("locations")

		locations := make([]cm.AppDefinitionDataLocationsItem, 0, len(model.Locations))
		for index, location := range model.Locations {
			requestLocation, locationDiags := location.ToAppDefinitionDataLocationsItem(path.AtListIndex(index))
			diags.Append(locationDiags...)

			if !locationDiags.HasError() {
				locations = append(locations, requestLocation)
			}
		}

		fields.Locations = locations
	}

	if model.Parameters != nil {
		parameters, parameterDiags := model.Parameters.ToAppDefinitionParameters(path.AtName("parameters"))
		diags.Append(parameterDiags...)

		if !parameterDiags.HasError() {
			fields.Parameters.SetTo(parameters)
		}
	}

	if diags.HasError() {
		return cm.AppDefinitionData{}, diags
	}

	return fields, diags
}

func (model AppDefinitionLocationsItem) ToAppDefinitionDataLocationsItem(path path.Path) (cm.AppDefinitionDataLocationsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	location, locationDiags := appRequestRequiredString(model.Location, path.AtName("location"))
	diags.Append(locationDiags...)

	item := cm.AppDefinitionDataLocationsItem{
		Location: location,
	}

	if model.FieldTypes != nil {
		fieldTypes := make([]cm.AppDefinitionDataLocationsItemFieldTypesItem, 0, len(model.FieldTypes))

		for index, fieldType := range model.FieldTypes {
			fieldType, fieldTypeDiags := appDefinitionLocationFieldTypeToRequest(
				fieldType,
				path.AtName("field_types").AtListIndex(index),
			)
			diags.Append(fieldTypeDiags...)

			if fieldTypeDiags.HasError() {
				continue
			}

			fieldTypesItem := cm.AppDefinitionDataLocationsItemFieldTypesItem{
				Type:     fieldType.Type,
				LinkType: cm.NewOptPointerString(fieldType.LinkType),
			}

			if fieldType.Items != nil {
				fieldTypesItem.Items.SetTo(cm.AppDefinitionDataLocationsItemFieldTypesItemItems{
					Type:     fieldType.Items.Type,
					LinkType: cm.NewOptPointerString(fieldType.Items.LinkType),
				})
			}

			fieldTypes = append(fieldTypes, fieldTypesItem)
		}

		item.FieldTypes = fieldTypes
	}

	if model.NavigationItem != nil {
		name, nameDiags := appRequestRequiredString(model.NavigationItem.Name, path.AtName("navigation_item").AtName("name"))
		pathValue, pathDiags := appRequestRequiredString(model.NavigationItem.Path, path.AtName("navigation_item").AtName("path"))

		diags.Append(nameDiags...)
		diags.Append(pathDiags...)

		item.NavigationItem.SetTo(cm.AppDefinitionDataLocationsItemNavigationItem{
			Name: name,
			Path: pathValue,
		})
	}

	if diags.HasError() {
		return cm.AppDefinitionDataLocationsItem{}, diags
	}

	return item, diags
}

type appDefinitionLocationFieldTypeRequest struct {
	Type     string
	LinkType *string
	Items    *appDefinitionLocationFieldTypeRequest
}

func appDefinitionLocationFieldTypeToRequest(
	model AppDefinitionLocationFieldTypesItem,
	path path.Path,
) (appDefinitionLocationFieldTypeRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	typeValue, typeDiags := appRequestRequiredString(model.Type, path.AtName("type"))
	linkType, linkTypeDiags := appRequestOptionalString(model.LinkType, path.AtName("link_type"))

	diags.Append(typeDiags...)
	diags.Append(linkTypeDiags...)

	fieldType := appDefinitionLocationFieldTypeRequest{
		Type:     typeValue,
		LinkType: linkType,
	}

	if model.Items != nil {
		itemsType, itemsTypeDiags := appRequestRequiredString(model.Items.Type, path.AtName("items").AtName("type"))
		itemsLinkType, itemsLinkTypeDiags := appRequestOptionalString(model.Items.LinkType, path.AtName("items").AtName("link_type"))

		diags.Append(itemsTypeDiags...)
		diags.Append(itemsLinkTypeDiags...)

		fieldType.Items = &appDefinitionLocationFieldTypeRequest{
			Type:     itemsType,
			LinkType: itemsLinkType,
		}
	}

	if diags.HasError() {
		return appDefinitionLocationFieldTypeRequest{}, diags
	}

	return fieldType, diags
}
