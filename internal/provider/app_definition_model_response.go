package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewAppDefinitionResourceModelFromResponse(ctx context.Context, response cm.AppDefinition) (AppDefinitionModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	organizationID := response.Sys.Organization.Sys.ID
	appDefinitionID := response.Sys.ID

	model := AppDefinitionModel{
		IDIdentityModel: IDIdentityModel{
			ID: types.StringValue(organizationID + "/" + appDefinitionID),
		},
		AppDefinitionIdentityModel: AppDefinitionIdentityModel{
			OrganizationID:  types.StringValue(organizationID),
			AppDefinitionID: types.StringValue(appDefinitionID),
		},
	}

	model.Name = types.StringValue(response.Name)

	model.Src = types.StringPointerValue(response.Src.ValueStringPointer())

	if bundle, ok := response.Bundle.Get(); ok {
		model.BundleID = types.StringValue(bundle.Sys.ID)
	}

	model.Locations = NewAppDefinitionLocationItemSliceFromAppDefinitionLocations(response.Locations)

	if parameters, ok := response.Parameters.Get(); ok {
		var installationParameters, instanceParameters []AppDefinitionParameter

		if parameters.Installation != nil {
			installationParameters = make([]AppDefinitionParameter, 0, len(parameters.Installation))

			for _, element := range parameters.Installation {
				parameter, parameterDiags := NewAppDefinitionParameterFromResponse(ctx, element)
				diags.Append(parameterDiags...)

				installationParameters = append(installationParameters, parameter)
			}
		}

		if parameters.Instance != nil {
			instanceParameters = make([]AppDefinitionParameter, 0, len(parameters.Instance))

			for _, element := range parameters.Instance {
				parameter, parameterDiags := NewAppDefinitionParameterFromResponse(ctx, element)
				diags.Append(parameterDiags...)

				instanceParameters = append(instanceParameters, parameter)
			}
		}

		model.Parameters = &AppDefinitionParameters{
			Installation: installationParameters,
			Instance:     instanceParameters,
		}
	}

	return model, diags
}

func NewAppDefinitionLocationItemSliceFromAppDefinitionLocations(locations []cm.AppDefinitionLocationsItem) []AppDefinitionLocationsItem {
	if locations == nil {
		return nil
	}

	items := make([]AppDefinitionLocationsItem, len(locations))

	for i, location := range locations {
		fieldTypes := NewAppDefinitionLocationFieldTypesItemSliceFromFieldTypes(location.FieldTypes)

		items[i] = AppDefinitionLocationsItem{
			Location:   types.StringValue(location.Location),
			FieldTypes: fieldTypes,
		}

		if navigationItem, ok := location.NavigationItem.Get(); ok {
			items[i].NavigationItem = &AppDefinitionLocationNavigationItem{
				Name: types.StringValue(navigationItem.Name),
				Path: types.StringValue(navigationItem.Path),
			}
		}
	}

	return items
}

func NewAppDefinitionLocationFieldTypesItemSliceFromFieldTypes(fieldTypes []cm.AppDefinitionLocationsItemFieldTypesItem) []AppDefinitionLocationFieldTypesItem {
	if fieldTypes == nil {
		return nil
	}

	fieldTypedItems := make([]AppDefinitionLocationFieldTypesItem, len(fieldTypes))

	for i, fieldType := range fieldTypes {
		item := AppDefinitionLocationFieldTypesItem{
			Type:     types.StringValue(fieldType.Type),
			LinkType: types.StringPointerValue(fieldType.LinkType.ValueStringPointer()),
		}

		if fieldTypesItems, ok := fieldType.Items.Get(); ok {
			items := NewAppDefinitionLocationFieldTypeItemsItemSliceFromItems(fieldTypesItems)
			item.Items = &items
		}

		fieldTypedItems[i] = item
	}

	return fieldTypedItems
}

func NewAppDefinitionLocationFieldTypeItemsItemSliceFromItems(items cm.AppDefinitionLocationsItemFieldTypesItemItems) AppDefinitionLocationFieldTypeItemsItem {
	item := AppDefinitionLocationFieldTypeItemsItem{
		Type:     types.StringValue(items.Type),
		LinkType: types.StringPointerValue(items.LinkType.ValueStringPointer()),
	}

	return item
}

func NewAppDefinitionParameterFromResponse(_ context.Context, parameter cm.AppDefinitionParameter) (AppDefinitionParameter, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := AppDefinitionParameter{
		Name:        parameter.Name,
		ID:          parameter.ID,
		Description: parameter.Description.ValueStringPointer(),
		Type:        parameter.Type,
		Required:    parameter.Required.ValueBoolPointer(),
	}

	if parameter.Options != nil {
		optionsElements := make([]jsontypes.Normalized, 0, len(parameter.Options))

		for _, option := range parameter.Options {
			if option != nil {
				optionsElements = append(optionsElements, NewNormalizedJSONTypesNormalizedValue(option))
			}
		}

		model.Options = NewTypedList(optionsElements)
	}

	if parameter.Default != nil {
		model.Default = NewNormalizedJSONTypesNormalizedValue(parameter.Default)
	}

	if labels, ok := parameter.Labels.Get(); ok {
		model.Labels = &AppDefinitionParameterLabels{
			Empty: labels.Empty.ValueStringPointer(),
			True:  labels.True.ValueStringPointer(),
			False: labels.False.ValueStringPointer(),
		}
	}

	return model, diags
}
