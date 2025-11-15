package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *AppDefinitionModel) ToAppDefinitionFields(ctx context.Context, path path.Path) (cm.AppDefinitionFields, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	fields := cm.AppDefinitionFields{
		Name: model.Name.ValueString(),
	}

	if !model.Src.IsUnknown() {
		fields.Src = cm.NewOptPointerString(model.Src.ValueStringPointer())
	}

	if !model.BundleID.IsNull() && !model.BundleID.IsUnknown() {
		fields.Bundle.SetTo(cm.AppBundleLink{
			Sys: cm.AppBundleLinkSys{
				Type:     cm.AppBundleLinkSysTypeLink,
				LinkType: cm.AppBundleLinkSysLinkTypeAppBundle,
				ID:       model.BundleID.ValueString(),
			},
		})
	}

	if model.Locations != nil {
		locationsPath := path.AtName("locations")

		locations := make([]cm.AppDefinitionFieldsLocationsItem, 0, len(model.Locations))
		for _, location := range model.Locations {
			locationPath := locationsPath.AtListIndex(len(locations))

			locationsItem := location.ToAppDefinitionFieldsLocationsItem(ctx, locationPath)

			locations = append(locations, locationsItem)
		}

		fields.Locations = locations
	}

	if model.Parameters != nil {
		parametersPath := path.AtName("parameters")

		var installationParameters, instanceParameters []cm.AppDefinitionParameter

		if model.Parameters.Installation != nil {
			installationPath := parametersPath.AtName("installation")

			installationParameters = make([]cm.AppDefinitionParameter, 0, len(model.Parameters.Installation))

			for index, element := range model.Parameters.Installation {
				parameter, parameterDiags := element.ToAppDefinitionParameter(ctx, installationPath.AtListIndex(index))
				diags.Append(parameterDiags...)

				installationParameters = append(installationParameters, parameter)
			}
		}

		if model.Parameters.Instance != nil {
			instancePath := parametersPath.AtName("instance")

			instanceParameters = make([]cm.AppDefinitionParameter, 0, len(model.Parameters.Instance))

			for index, element := range model.Parameters.Instance {
				parameter, parameterDiags := element.ToAppDefinitionParameter(ctx, instancePath.AtListIndex(index))
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

func (model AppDefinitionLocationsItem) ToAppDefinitionFieldsLocationsItem(_ context.Context, _ path.Path) cm.AppDefinitionFieldsLocationsItem {
	item := cm.AppDefinitionFieldsLocationsItem{
		Location: model.Location.ValueString(),
	}

	if model.FieldTypes != nil {
		fieldTypes := make([]cm.AppDefinitionFieldsLocationsItemFieldTypesItem, 0, len(model.FieldTypes))

		for _, fieldType := range model.FieldTypes {
			fieldTypesItem := cm.AppDefinitionFieldsLocationsItemFieldTypesItem{
				Type:     fieldType.Type.ValueString(),
				LinkType: cm.NewOptPointerString(fieldType.LinkType.ValueStringPointer()),
			}

			if fieldType.Items != nil {
				fieldTypesItem.Items.SetTo(cm.AppDefinitionFieldsLocationsItemFieldTypesItemItems{
					Type:     fieldType.Items.Type.ValueString(),
					LinkType: cm.NewOptPointerString(fieldType.Items.LinkType.ValueStringPointer()),
				})
			}

			fieldTypes = append(fieldTypes, fieldTypesItem)
		}

		item.FieldTypes = fieldTypes
	}

	if model.NavigationItem != nil {
		item.NavigationItem.SetTo(cm.AppDefinitionFieldsLocationsItemNavigationItem{
			Name: model.NavigationItem.Name.ValueString(),
			Path: model.NavigationItem.Path.ValueString(),
		})
	}

	return item
}

func (model AppDefinitionParameter) ToAppDefinitionParameter(_ context.Context, _ path.Path) (cm.AppDefinitionParameter, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	parameter := cm.AppDefinitionParameter{
		ID:          model.ID,
		Name:        model.Name,
		Description: cm.NewOptPointerString(model.Description),
		Type:        model.Type,
		Required:    cm.NewOptPointerBool(model.Required),
	}

	if !model.Default.IsUnknown() {
		value := model.Default.ValueStringPointer()
		if value != nil {
			parameter.Default = jx.Raw(*value)
		}
	}

	if !model.Options.IsUnknown() && !model.Options.IsNull() {
		options := make([]jx.Raw, 0, len(model.Options.Elements()))

		for _, option := range model.Options.Elements() {
			value := option.ValueStringPointer()
			if value != nil {
				options = append(options, jx.Raw(*value))
			}
		}

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
