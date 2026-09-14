package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource              = (*localesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*localesDataSource)(nil)
)

//nolint:ireturn
func NewLocalesDataSource() datasource.DataSource { return &localesDataSource{} }

type localesDataSource struct{ providerData ContentfulProviderData }

func (d *localesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locales"
}

func (d *localesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = LocalesDataSourceSchema(ctx)
}

func (d *localesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *localesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LocalesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(requireDiscoveryID(path.Root("space_id"), data.SpaceID)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("environment_id"), data.EnvironmentID)...)

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

	items, diagnostics := readLocales(ctx, d.providerData.client, data.SpaceID.ValueString(), data.EnvironmentID.ValueString())
	resp.Diagnostics.Append(diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Locales = items
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(data.SpaceID.ValueString(), data.EnvironmentID.ValueString())

	if ctx.Err() != nil {
		resp.Diagnostics.AddError("Failed to read locales", ctx.Err().Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
