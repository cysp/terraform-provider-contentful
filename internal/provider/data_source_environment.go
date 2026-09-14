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
	_ datasource.DataSource              = (*environmentDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*environmentDataSource)(nil)
)

//nolint:ireturn
func NewEnvironmentDataSource() datasource.DataSource { return &environmentDataSource{} }

type environmentDataSource struct{ providerData ContentfulProviderData }

func (d *environmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *environmentDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = EnvironmentDataSourceSchema(ctx)
}

func (d *environmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *environmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EnvironmentDataSourceModel
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

	response, err := d.providerData.client.GetEnvironment(ctx, cm.GetEnvironmentParams{SpaceID: data.SpaceID.ValueString(), EnvironmentID: data.EnvironmentID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read environment", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	entity, ok := response.(*cm.Environment)
	if !ok {
		resp.Diagnostics.AddError("Failed to read environment", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	if entity.Sys.ID != data.EnvironmentID.ValueString() {
		resp.Diagnostics.AddError("Unexpected response identity", "The returned environment ID differs from the requested ID.")

		return
	}

	item, diagnostics := newEnvironmentDataSourceItem(*entity, data.SpaceID.ValueString())
	resp.Diagnostics.Append(diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.EnvironmentDataSourceItemModel = item
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(data.SpaceID.ValueString(), data.EnvironmentID.ValueString())

	if ctx.Err() != nil {
		resp.Diagnostics.AddError("Failed to read environment", ctx.Err().Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
