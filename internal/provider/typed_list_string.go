package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTypedListFromStringSlice(ctx context.Context, slice []string) (TypedList[types.String], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	listElementValues := make([]types.String, len(slice))
	for index, item := range slice {
		listElementValues[index] = types.StringValue(item)
	}

	list, listDiags := NewTypedList(ctx, listElementValues)
	diags.Append(listDiags...)

	return list, diags
}
