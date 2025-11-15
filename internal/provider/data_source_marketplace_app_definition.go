package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = (*marketplaceAppDefinitionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*marketplaceAppDefinitionDataSource)(nil)
)

//nolint:ireturn
func NewMarketplaceAppDefinitionDataSource() datasource.DataSource {
	return &marketplaceAppDefinitionDataSource{}
}

type marketplaceAppDefinitionDataSource struct {
	providerData ContentfulProviderData
}

func (d *marketplaceAppDefinitionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_marketplace_app_definition"
}

func (d *marketplaceAppDefinitionDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = MarketplaceAppDefinitionDataSourceSchema(ctx)
}

func (d *marketplaceAppDefinitionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *marketplaceAppDefinitionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AppDefinitionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := cm.GetMarketplaceAppDefinitionsParams{
		SysIDIn: []string{data.AppDefinitionID.ValueString()},
	}

	response, err := d.providerData.client.GetMarketplaceAppDefinitions(ctx, params)

	tflog.Info(ctx, "marketplace_app_definition.read", map[string]any{
		"params":   params,
		"response": response,
		"err":      err,
	})

	switch response := response.(type) {
	case *cm.GetMarketplaceAppDefinitionsOK:
		if len(response.Items) == 0 {
			resp.Diagnostics.AddError("No app definitions found", "No app definitions were found for the given ID")

			return
		}

		responseModel, responseModelDiags := NewAppDefinitionResourceModelFromResponse(ctx, response.Items[0])
		resp.Diagnostics.Append(responseModelDiags...)

		data = responseModel
	default:
		resp.Diagnostics.AddError("Failed to read app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
