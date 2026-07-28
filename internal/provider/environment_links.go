package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToEnvironmentLinks(_ context.Context, valuePath path.Path, value TypedList[types.String]) ([]cm.EnvironmentLink, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if value.IsNull() || value.IsUnknown() {
		return nil, diags
	}

	environmentIDs, environmentIDDiags := knownStringListElements(valuePath, value.Elements())
	diags.Append(environmentIDDiags...)

	if diags.HasError() {
		return nil, diags
	}

	environments := make([]cm.EnvironmentLink, 0, len(environmentIDs))

	for _, environmentID := range environmentIDs {
		environments = append(environments, cm.NewEnvironmentLink(environmentID))
	}

	return environments, diags
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
