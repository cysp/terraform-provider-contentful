package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *AppDefinitionBaseModel) ToAppDefinitionData(ctx context.Context, path path.Path) (cm.AppDefinitionData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.AppDefinitionData{
		Name: model.Name.ValueString(),
	}

	if !model.Src.IsUnknown() {
		fields.Src = cm.NewOptPointerString(model.Src.ValueStringPointer())
	}

	if !model.BundleID.IsNull() && !model.BundleID.IsUnknown() {
		fields.Bundle.SetTo(cm.NewAppBundleLink(model.BundleID.ValueString()))
	}

	if model.Locations != nil {
		path := path.AtName("locations")

		locations := make([]cm.AppDefinitionDataLocationsItem, 0, len(model.Locations))
		for _, location := range model.Locations {
			path := path.AtListIndex(len(locations))

			locationsItem := location.ToAppDefinitionDataLocationsItem(ctx, path)

			locations = append(locations, locationsItem)
		}

		fields.Locations = locations
	}

	if model.Parameters != nil {
		path := path.AtName("parameters")

		var installationParameters, instanceParameters []cm.AppDefinitionParameter

		if model.Parameters.Installation != nil {
			path := path.AtName("installation")

			installationParameters = make([]cm.AppDefinitionParameter, 0, len(model.Parameters.Installation))

			for index, element := range model.Parameters.Installation {
				parameter, parameterDiags := element.ToAppDefinitionParameter(ctx, path.AtListIndex(index))
				diags.Append(parameterDiags...)

				installationParameters = append(installationParameters, parameter)
			}
		}

		if model.Parameters.Instance != nil {
			path := path.AtName("instance")

			instanceParameters = make([]cm.AppDefinitionParameter, 0, len(model.Parameters.Instance))

			for index, element := range model.Parameters.Instance {
				parameter, parameterDiags := element.ToAppDefinitionParameter(ctx, path.AtListIndex(index))
				diags.Append(parameterDiags...)

				instanceParameters = append(instanceParameters, parameter)
			}
		}

		fields.Parameters.SetTo(cm.AppDefinitionParameters{
			Installation: installationParameters,
			Instance:     instanceParameters,
		})
	}

	return fields, diags
}

func (model AppDefinitionLocationsItem) ToAppDefinitionDataLocationsItem(_ context.Context, _ path.Path) cm.AppDefinitionDataLocationsItem {
	item := cm.AppDefinitionDataLocationsItem{
		Location: model.Location.ValueString(),
	}

	if model.FieldTypes != nil {
		fieldTypes := make([]cm.AppDefinitionDataLocationsItemFieldTypesItem, 0, len(model.FieldTypes))

		for _, fieldType := range model.FieldTypes {
			fieldTypesItem := cm.AppDefinitionDataLocationsItemFieldTypesItem{
				Type:     fieldType.Type.ValueString(),
				LinkType: cm.NewOptPointerString(fieldType.LinkType.ValueStringPointer()),
			}

			if fieldType.Items != nil {
				fieldTypesItem.Items.SetTo(cm.AppDefinitionDataLocationsItemFieldTypesItemItems{
					Type:     fieldType.Items.Type.ValueString(),
					LinkType: cm.NewOptPointerString(fieldType.Items.LinkType.ValueStringPointer()),
				})
			}

			fieldTypes = append(fieldTypes, fieldTypesItem)
		}

		item.FieldTypes = fieldTypes
	}

	if model.NavigationItem != nil {
		item.NavigationItem.SetTo(cm.AppDefinitionDataLocationsItemNavigationItem{
			Name: model.NavigationItem.Name.ValueString(),
			Path: model.NavigationItem.Path.ValueString(),
		})
	}

	return item
}

func (model AppDefinitionParameter) ToAppDefinitionParameter(_ context.Context, parameterPath path.Path) (cm.AppDefinitionParameter, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	parameter := cm.AppDefinitionParameter{
		ID:          model.ID,
		Name:        model.Name,
		Description: cm.NewOptPointerString(model.Description),
		Type:        model.Type,
		Required:    cm.NewOptPointerBool(model.Required),
	}

	if model.Default.IsUnknown() {
		diags.AddAttributeError(parameterPath.AtName("default"), "Unexpected unknown parameter default", "The parameter default must be known before it can be sent to Contentful.")
	} else {
		value := model.Default.ValueStringPointer()
		if value != nil {
			parameter.Default = jx.Raw(*value)
		}
	}

	options, optionsSet, optionsDiags := toAppDefinitionParameterOptions(model, parameterPath.AtName("options"))
	diags.Append(optionsDiags...)

	if optionsSet {
		parameter.Options = options
	}

	if model.Labels != nil {
		parameter.Labels.SetTo(cm.AppDefinitionParameterLabels{
			Empty: cm.NewOptPointerString(model.Labels.Empty),
			True:  cm.NewOptPointerString(model.Labels.True),
			False: cm.NewOptPointerString(model.Labels.False),
		})
	}

	return parameter, diags
}

func toAppDefinitionParameterOptions(model AppDefinitionParameter, optionsPath path.Path) ([]jx.Raw, bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if model.Options.IsUnknown() {
		diags.AddAttributeError(optionsPath, "Unexpected unknown parameter options", "Parameter options must be known before they can be sent to Contentful.")

		return nil, false, diags
	}

	if model.Options.IsNull() {
		return nil, false, nil
	}

	options := make([]jx.Raw, 0, len(model.Options.Elements()))

	for index, option := range model.Options.Elements() {
		optionPath := optionsPath.AtListIndex(index)

		switch {
		case option.IsUnknown():
			diags.AddAttributeError(optionPath, "Unexpected unknown parameter option", "Parameter options must be known before they can be sent to Contentful.")
		case option.IsNull():
			diags.AddAttributeError(optionPath, "Unexpected null parameter option", "Parameter options cannot be null.")
		default:
			options = append(options, jx.Raw(option.ValueString()))
		}
	}

	return options, true, diags
}
