package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource              = (*spaceDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*spaceDataSource)(nil)
)

//nolint:ireturn
func NewSpaceDataSource() datasource.DataSource { return &spaceDataSource{} }

type spaceDataSource struct{ providerData ContentfulProviderData }

func (d *spaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (d *spaceDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SpaceDataSourceSchema(ctx)
}

func (d *spaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *spaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SpaceDataSourceModel
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

	response, err := d.providerData.client.GetSpace(ctx, cm.GetSpaceParams{SpaceID: data.SpaceID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read space", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	entity, ok := response.(*cm.Space)
	if !ok {
		resp.Diagnostics.AddError("Failed to read space", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	if entity.Sys.ID != data.SpaceID.ValueString() {
		resp.Diagnostics.AddError("Unexpected response identity", "The returned space ID differs from the requested ID.")

		return
	}

	data.SpaceDataSourceItemModel = newSpaceDataSourceItem(*entity)
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(data.SpaceID.ValueString())

	if ctx.Err() != nil {
		resp.Diagnostics.AddError("Failed to read space", ctx.Err().Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
