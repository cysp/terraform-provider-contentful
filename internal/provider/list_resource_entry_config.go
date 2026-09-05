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
	spaceID, spaceIDDiags := requestRequiredString(c.SpaceID, path.Root("space_id"))
	environmentID, environmentIDDiags := requestRequiredString(c.EnvironmentID, path.Root("environment_id"))
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
		Description: "Lists Contentful Entries in an existing space and environment.",
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Description: "ID of the space from which to list Entries.",
				Required:    true,
			},
			"environment_id": listschema.StringAttribute{
				Description: "ID of the environment from which to list Entries.",
				Required:    true,
			},
			"content_type": listschema.StringAttribute{
				Description: "Content Type ID to use when filtering Entries.",
				Optional:    true,
			},
			"order": listschema.ListAttribute{
				Description: "Contentful Entry collection order expressions. Prefix an attribute with `-` for descending order. Empty expressions are ignored.",
				ElementType: types.StringNull().Type(ctx),
				CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
				Optional:    true,
			},
			"query": listschema.MapAttribute{
				Description: "Additional Contentful Entry collection query parameters, keyed by parameter name. `skip` and `limit` are ignored because pagination is controlled by the list operation. Prefer the dedicated `content_type` and `order` attributes for those filters.",
				ElementType: types.StringNull().Type(ctx),
				CustomType:  NewTypedMapNull[types.String]().CustomType(ctx),
				Optional:    true,
			},
		},
	}
}
