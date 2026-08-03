package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type entryListResourceConfig struct {
	SpaceID       types.String            `tfsdk:"space_id"`
	EnvironmentID types.String            `tfsdk:"environment_id"`
	ContentType   types.String            `tfsdk:"content_type"`
	Order         TypedList[types.String] `tfsdk:"order"`
	Query         TypedMap[types.String]  `tfsdk:"query"`
}

type entryListResourceRequest struct {
	params cm.GetEntriesParams
	query  map[string]string
}

func (c entryListResourceConfig) request() (entryListResourceRequest, diag.Diagnostics) {
	spaceID, spaceIDDiags := knownListResourceString(path.Root("space_id"), c.SpaceID)
	environmentID, environmentIDDiags := knownListResourceString(path.Root("environment_id"), c.EnvironmentID)
	contentType, contentTypePresent, contentTypeDiags := knownOptionalListResourceString(path.Root("content_type"), c.ContentType)
	order, orderDiags := knownOptionalListResourceStringList(path.Root("order"), c.Order)
	query, queryDiags := knownOptionalListResourceStringMap(path.Root("query"), c.Query)

	if order != nil {
		filteredOrder := make([]string, 0, len(order))
		for _, value := range order {
			if value != "" {
				filteredOrder = append(filteredOrder, value)
			}
		}

		order = filteredOrder
	}

	diags := diag.Diagnostics{}
	diags.Append(spaceIDDiags...)
	diags.Append(environmentIDDiags...)
	diags.Append(contentTypeDiags...)
	diags.Append(orderDiags...)
	diags.Append(queryDiags...)

	if diags.HasError() {
		return entryListResourceRequest{}, diags
	}

	request := entryListResourceRequest{
		params: cm.GetEntriesParams{
			SpaceID:       spaceID,
			EnvironmentID: environmentID,
			Order:         order,
		},
		query: query,
	}

	if contentTypePresent && contentType != "" {
		request.params.ContentType.SetTo(contentType)
	}

	return request, diags
}

func EntryListResourceConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Description: "List entries from a Contentful space and environment",
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Description: "The ID of the space for which to list entries.",
				Required:    true,
			},
			"environment_id": listschema.StringAttribute{
				Description: "The ID of the environment for which to list entries.",
				Required:    true,
			},
			"content_type": listschema.StringAttribute{
				Description: "Query entries for a specific content type.",
				Optional:    true,
			},
			"order": listschema.ListAttribute{
				Description: "Order entries by one or more attributes.",
				ElementType: types.StringNull().Type(ctx),
				CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
				Optional:    true,
			},
			"query": listschema.MapAttribute{
				Description: "Query parameters to filter the entries listed.",
				ElementType: types.StringNull().Type(ctx),
				CustomType:  NewTypedMapNull[types.String]().CustomType(ctx),
				Optional:    true,
			},
		},
	}
}
