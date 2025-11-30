package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ list.ListResource              = (*entryListResource)(nil)
	_ list.ListResourceWithConfigure = (*entryListResource)(nil)
)

//nolint:ireturn
func NewEntryListResource() list.ListResource {
	return &entryListResource{}
}

type entryListResource struct {
	providerData ContentfulProviderData
}

func (r *entryListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entry"
}

func (r *entryListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = EntryListResourceSchema(ctx)
}

func (r *entryListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromResourceConfigureRequest(req, &r.providerData)...)
}

func (r *entryListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	var config EntryListConfigModel

	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	params := cm.GetEntriesParams{
		SpaceID:       config.SpaceID.ValueString(),
		EnvironmentID: config.EnvironmentID.ValueString(),
	}

	// Apply query parameters
	if !config.ContentType.IsNull() && !config.ContentType.IsUnknown() {
		params.ContentType = cm.NewOptString(config.ContentType.ValueString())
	}

	if !config.Select.IsNull() && !config.Select.IsUnknown() {
		params.Select = cm.NewOptString(config.Select.ValueString())
	}

	if !config.Order.IsNull() && !config.Order.IsUnknown() {
		params.Order = cm.NewOptString(config.Order.ValueString())
	}

	if !config.Query.IsNull() && !config.Query.IsUnknown() {
		params.Query = cm.NewOptString(config.Query.ValueString())
	}

	// Apply limit from request or config
	if req.Limit > 0 {
		params.Limit = cm.NewOptInt(int(req.Limit))
	} else if !config.Limit.IsNull() && !config.Limit.IsUnknown() {
		params.Limit = cm.NewOptInt(int(config.Limit.ValueInt64()))
	}

	if !config.Skip.IsNull() && !config.Skip.IsUnknown() {
		params.Skip = cm.NewOptInt(int(config.Skip.ValueInt64()))
	}

	result, err := r.providerData.client.GetEntries(ctx, params)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Failed to list entries", err.Error()),
		})
		return
	}

	entryCollection, ok := result.(*cm.EntryCollection)
	if !ok {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid response", "Expected EntryCollection response"),
		})
		return
	}

	stream.Results = func(push func(list.ListResult) bool) {
		for _, entry := range entryCollection.Items {
			listResult := req.NewListResult(ctx)

			// Set identity data
			identityModel := EntryIdentityModel{
				SpaceID:       types.StringValue(params.SpaceID),
				EnvironmentID: types.StringValue(params.EnvironmentID),
				EntryID:       types.StringValue(entry.Sys.ID),
			}

			diags := listResult.Identity.Set(ctx, &identityModel)
			if diags.HasError() {
				listResult.Diagnostics.Append(diags...)
				push(listResult)
				return
			}

			// Set resource data if requested
			if req.IncludeResource {
				responseModel, responseDiags := convertEntryToModel(ctx, entry, params.SpaceID, params.EnvironmentID)
				if responseDiags.HasError() {
					listResult.Diagnostics.Append(responseDiags...)
					push(listResult)
					return
				}

				diags = listResult.Resource.Set(ctx, &responseModel)
				if diags.HasError() {
					listResult.Diagnostics.Append(diags...)
					push(listResult)
					return
				}
			}

			// Set display name
			listResult.DisplayName = entry.Sys.ID

			if !push(listResult) {
				return
			}
		}
	}
}

func convertEntryToModel(ctx context.Context, entry cm.Entry, spaceID, environmentID string) (EntryModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Use the existing conversion function
	model, convertDiags := NewEntryResourceModelFromResponse(ctx, entry)
	diags.Append(convertDiags...)

	return model, diags
}
