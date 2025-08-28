package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTypedListFromStringSlice(slice []string) TypedList[types.String] {
	listElementValues := make([]types.String, len(slice))
	for index, item := range slice {
		listElementValues[index] = types.StringValue(item)
	}

	return NewTypedList(listElementValues)
}
