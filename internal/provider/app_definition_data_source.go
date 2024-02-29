package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/datasource_app_definition"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*appDefinitionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*appDefinitionDataSource)(nil)
)

//nolint:ireturn
func NewAppDefinitionDataSource() datasource.DataSource {
	return &appDefinitionDataSource{}
}

type appDefinitionDataSource struct {
	providerData ContentfulProviderData
}

func (d *appDefinitionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_definition"
}

func (d *appDefinitionDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_app_definition.AppDefinitionDataSourceSchema(ctx)
}

func (d *appDefinitionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	util.ProviderDataFromDataSourceConfigureRequest(req, &d.providerData, resp)
}

func (d *appDefinitionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_app_definition.AppDefinitionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := d.providerData.client.GetAppDefinition(ctx, contentfulManagement.GetAppDefinitionParams{
		OrganizationID:  data.OrganizationId.ValueString(),
		AppDefinitionID: data.AppDefinitionId.ValueString(),
	})

	switch response := response.(type) {
	case *contentfulManagement.AppDefinition:
		data.Name = types.StringValue(response.Name)
	default:
		resp.Diagnostics.AddError("Failed to read app definition", util.ErrorDetailFromContentfulManagementResponse(response, err))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
