package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToEnvironmentLinks(ctx context.Context, path path.Path, value types.List) ([]cm.EnvironmentLink, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if value.IsUnknown() {
		return nil, diags
	}

	environmentIDs := make([]string, len(value.Elements()))
	diags.Append(value.ElementsAs(ctx, &environmentIDs, false)...)

	environments := make([]cm.EnvironmentLink, len(environmentIDs))

	for index, environmentString := range environmentIDs {
		path := path.AtListIndex(index)

		environmentsItem, environmentsItemDiags := ToEnvironmentLink(ctx, path, environmentString)
		diags.Append(environmentsItemDiags...)

		environments[index] = environmentsItem
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

func NewEnvironmentIDsListValueFromEnvironmentLinks(_ context.Context, _ path.Path, environmentLinks []cm.EnvironmentLink) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]attr.Value, len(environmentLinks))

	for index, item := range environmentLinks {
		listElementValues[index] = types.StringValue(item.Sys.ID)
	}

	list, listDiags := types.ListValue(types.StringType, listElementValues)
	diags.Append(listDiags...)

	return list, diags
}
