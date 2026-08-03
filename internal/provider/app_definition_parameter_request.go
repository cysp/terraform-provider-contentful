package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *AppDefinitionParameters) ToAppDefinitionParameters(ctx context.Context, valuePath path.Path) (cm.AppDefinitionParameters, diag.Diagnostics) {
	installation, installationDiags := appDefinitionParameterListToRequest(ctx, valuePath.AtName("installation"), model.Installation)
	instance, instanceDiags := appDefinitionParameterListToRequest(ctx, valuePath.AtName("instance"), model.Instance)

	diags := diag.Diagnostics{}
	diags.Append(installationDiags...)
	diags.Append(instanceDiags...)

	if diags.HasError() {
		return cm.AppDefinitionParameters{}, diags
	}

	return cm.AppDefinitionParameters{
		Installation: installation,
		Instance:     instance,
	}, diags
}

func appDefinitionParameterListToRequest(ctx context.Context, valuePath path.Path, models []AppDefinitionParameter) ([]cm.AppDefinitionParameter, diag.Diagnostics) {
	if models == nil {
		return nil, nil
	}

	parameters := make([]cm.AppDefinitionParameter, 0, len(models))
	diags := diag.Diagnostics{}

	for index, model := range models {
		parameter, parameterDiags := model.ToAppDefinitionParameter(ctx, valuePath.AtListIndex(index))
		diags.Append(parameterDiags...)

		if !parameterDiags.HasError() {
			parameters = append(parameters, parameter)
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return parameters, diags
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

	switch {
	case model.Default.IsUnknown():
		diags.AddAttributeError(
			parameterPath.AtName("default"),
			"Unexpected unknown parameter default",
			"The parameter default must be known before it can be sent to Contentful.",
		)
	case model.Default.IsNull():
	default:
		parameter.Default = jx.Raw(model.Default.ValueString())
	}

	options, optionsSet, optionsDiags := appDefinitionParameterOptionsToRequest(
		parameterPath.AtName("options"),
		model.Options,
	)
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

	if diags.HasError() {
		return cm.AppDefinitionParameter{}, diags
	}

	return parameter, diags
}

func appDefinitionParameterOptionsToRequest(valuePath path.Path, value TypedList[jsontypes.Normalized]) ([]jx.Raw, bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown parameter options",
			"Parameter options must be known before they can be sent to Contentful.",
		)

		return nil, false, diags
	case value.IsNull():
		return nil, false, diags
	}

	options := make([]jx.Raw, 0, len(value.Elements()))

	for index, option := range value.Elements() {
		optionPath := valuePath.AtListIndex(index)

		switch {
		case option.IsUnknown():
			diags.AddAttributeError(
				optionPath,
				"Unexpected unknown parameter option",
				"The parameter option must be known before it can be sent to Contentful.",
			)
		case option.IsNull():
			diags.AddAttributeError(
				optionPath,
				"Unexpected null parameter option",
				"The parameter option cannot be null when it is sent to Contentful.",
			)
		default:
			options = append(options, jx.Raw(option.ValueString()))
		}
	}

	if diags.HasError() {
		return nil, false, diags
	}

	return options, true, diags
}
