package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToEnvironmentLinks(ctx context.Context, path path.Path, value TypedList[types.String]) ([]cm.EnvironmentLink, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		return nil, diags
	}

	environmentIDValues := value.Elements()

	environments := make([]cm.EnvironmentLink, 0, len(environmentIDValues))

	for index, environmentIDValue := range environmentIDValues {
		envPath := path.AtListIndex(index)

		if environmentIDValue.IsUnknown() {
			diags.AddAttributeError(envPath, "Unexpectedly unknown value", "")
		}

		environmentsItem, environmentsItemDiags := ToEnvironmentLink(ctx, envPath, environmentIDValue.ValueString())
		diags.Append(environmentsItemDiags...)

		environments = append(environments, environmentsItem)
	}

	return environments, diags
}

func ToEnvironmentLink(_ context.Context, _ path.Path, environmentID string) (cm.EnvironmentLink, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	item := cm.EnvironmentLink{
		Sys: cm.EnvironmentLinkSys{
			Type:     cm.EnvironmentLinkSysTypeLink,
			LinkType: cm.EnvironmentLinkSysLinkTypeEnvironment,
			ID:       environmentID,
		},
	}

	return item, diags
}

func NewEnvironmentIDsListValueFromEnvironmentLinks(_ context.Context, _ path.Path, environmentLinks []cm.EnvironmentLink) (TypedList[types.String], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]types.String, len(environmentLinks))

	for index, item := range environmentLinks {
		listElementValues[index] = types.StringValue(item.Sys.ID)
	}

	list := NewTypedList(listElementValues)

	return list, diags
}
