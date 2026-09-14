//nolint:dupl // Framework lifecycle wiring stays typed; family policy lives in shared projections and collection reading.
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource              = (*environmentsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*environmentsDataSource)(nil)
)

//nolint:ireturn
func NewEnvironmentsDataSource() datasource.DataSource { return &environmentsDataSource{} }

type environmentsDataSource struct{ providerData ContentfulProviderData }

func (d *environmentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environments"
}

func (d *environmentsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = EnvironmentsDataSourceSchema(ctx)
}

func (d *environmentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *environmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnvironmentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(requireDiscoveryID(path.Root("space_id"), data.SpaceID)...)

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

	items, diagnostics := readEnvironments(ctx, d.providerData.client, data.SpaceID.ValueString())
	resp.Diagnostics.Append(diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Environments = items
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(data.SpaceID.ValueString())

	if ctx.Err() != nil {
		resp.Diagnostics.AddError("Failed to read environments", ctx.Err().Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
