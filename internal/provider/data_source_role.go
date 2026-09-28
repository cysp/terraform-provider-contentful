package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource              = (*roleDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*roleDataSource)(nil)
)

//nolint:ireturn
func NewRoleDataSource() datasource.DataSource { return &roleDataSource{} }

type roleDataSource struct{ providerData ContentfulProviderData }

func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *roleDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = RoleDataSourceSchema(ctx)
}

func (d *roleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(requireDiscoveryID(path.Root("space_id"), data.SpaceID)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("role_id"), data.RoleID)...)

	if resp.Diagnostics.HasError() {
		return
	}

	timeout, timeoutDiagnostics := data.Timeouts.Read(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(timeoutDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	response, err := d.providerData.client.GetRole(ctx, cm.GetRoleParams{SpaceID: data.SpaceID.ValueString(), RoleID: data.RoleID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read role", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	entity, ok := response.(*cm.Role)
	if !ok {
		resp.Diagnostics.AddError("Failed to read role", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	if entity.Sys.ID != data.RoleID.ValueString() {
		resp.Diagnostics.AddError("Unexpected response identity", "The returned role ID differs from the requested ID.")

		return
	}

	resp.Diagnostics.Append(validateDiscoverySpace(entity.Sys.Space.Sys.ID, data.SpaceID.ValueString())...)

	if resp.Diagnostics.HasError() {
		return
	}

	item, diagnostics := newRoleDataSourceItem(ctx, *entity, path.Empty())
	resp.Diagnostics.Append(diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.RoleDataSourceItemModel = item
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(data.SpaceID.ValueString(), data.RoleID.ValueString())

	if ctx.Err() != nil {
		resp.Diagnostics.AddError("Failed to read role", ctx.Err().Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
