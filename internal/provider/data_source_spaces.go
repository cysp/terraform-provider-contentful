package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*spacesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*spacesDataSource)(nil)
)

//nolint:ireturn
func NewSpacesDataSource() datasource.DataSource { return &spacesDataSource{} }

type spacesDataSource struct{ providerData ContentfulProviderData }

func (d *spacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spaces"
}

func (d *spacesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = SpacesDataSourceSchema(ctx)
}

func (d *spacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *spacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SpacesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.OrganizationID.IsNull() {
		resp.Diagnostics.Append(requireDiscoveryID(path.Root("organization_id"), data.OrganizationID)...)
	}

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

	items, diagnostics := readSpaces(ctx, d.providerData.client, data.OrganizationID)
	resp.Diagnostics.Append(diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Spaces = items

	data.ID = types.StringValue("spaces")
	if !data.OrganizationID.IsNull() {
		data.ID = types.StringValue("organizations/" + data.OrganizationID.ValueString() + "/spaces")
	}

	if ctx.Err() != nil {
		resp.Diagnostics.AddError("Failed to read spaces", ctx.Err().Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
