//nolint:dupl // Framework lifecycle wiring stays typed; family policy lives in shared projections and collection reading.
package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource              = (*environmentAliasDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*environmentAliasDataSource)(nil)
)

//nolint:ireturn
func NewEnvironmentAliasDataSource() datasource.DataSource { return &environmentAliasDataSource{} }

type environmentAliasDataSource struct{ providerData ContentfulProviderData }

func (d *environmentAliasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_alias"
}

func (d *environmentAliasDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = EnvironmentAliasDataSourceSchema(ctx)
}

func (d *environmentAliasDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *environmentAliasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnvironmentAliasDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(requireDiscoveryID(path.Root("space_id"), data.SpaceID)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("environment_alias_id"), data.EnvironmentAliasID)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiags := data.Timeouts.Read(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	response, err := d.providerData.client.GetEnvironmentAlias(ctx, cm.GetEnvironmentAliasParams{SpaceID: data.SpaceID.ValueString(), EnvironmentAliasID: data.EnvironmentAliasID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	entity, ok := response.(*cm.EnvironmentAlias)
	if !ok {
		resp.Diagnostics.AddError("Failed to read environment alias", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	if entity.Sys.ID != data.EnvironmentAliasID.ValueString() {
		resp.Diagnostics.AddError("Unexpected response identity", "The returned environment alias ID differs from the requested ID.")

		return
	}

	item, diagnostics := newEnvironmentAliasDataSourceItem(*entity, data.SpaceID.ValueString())
	resp.Diagnostics.Append(diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.EnvironmentAliasDataSourceItemModel = item
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(data.SpaceID.ValueString(), data.EnvironmentAliasID.ValueString())

	if ctx.Err() != nil {
		resp.Diagnostics.AddError("Failed to read environment alias", ctx.Err().Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
