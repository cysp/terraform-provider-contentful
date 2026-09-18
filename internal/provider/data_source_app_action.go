package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type AppActionDataSourceModel struct {
	AppActionBaseModel

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

var (
	_ datasource.DataSource              = (*appActionDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*appActionDataSource)(nil)
)

//nolint:ireturn
func NewAppActionDataSource() datasource.DataSource { return &appActionDataSource{} }

type appActionDataSource struct{ providerData ContentfulProviderData }

func (d *appActionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_action"
}

func (d *appActionDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AppActionDataSourceSchema(ctx)
}

func (d *appActionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *appActionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AppActionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(appActionScope(data.AppActionBaseModel)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("app_action_id"), data.AppActionID)...)
	timeout, diags := data.Timeouts.Read(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	response, err := d.providerData.client.GetAppAction(ctx, cm.GetAppActionParams{OrganizationID: data.OrganizationID.ValueString(), AppDefinitionID: data.AppDefinitionID.ValueString(), AppActionID: data.AppActionID.ValueString()})

	result, ok := response.(*cm.AppAction)
	if err != nil || !ok || result == nil {
		resp.Diagnostics.AddError("Failed to read app action", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	base, identityDiags := appActionResponse(*result, data.AppActionBaseModel)
	resp.Diagnostics.Append(identityDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.AppActionBaseModel = base

	if ctx.Err() != nil {
		resp.Diagnostics.AddError("Failed to read app action", ctx.Err().Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
