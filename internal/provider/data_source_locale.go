package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource              = (*localeDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*localeDataSource)(nil)
)

//nolint:ireturn
func NewLocaleDataSource() datasource.DataSource { return &localeDataSource{} }

type localeDataSource struct{ providerData ContentfulProviderData }

func (d *localeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locale"
}

func (d *localeDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = LocaleDataSourceSchema(ctx)
}

func (d *localeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *localeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LocaleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(requireDiscoveryID(path.Root("space_id"), data.SpaceID)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("environment_id"), data.EnvironmentID)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("locale_id"), data.LocaleID)...)

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

	response, err := d.providerData.client.GetLocale(ctx, cm.GetLocaleParams{SpaceID: data.SpaceID.ValueString(), EnvironmentID: data.EnvironmentID.ValueString(), LocaleID: data.LocaleID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read locale", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	entity, ok := response.(*cm.Locale)
	if !ok {
		resp.Diagnostics.AddError("Failed to read locale", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	if entity.Sys.ID != data.LocaleID.ValueString() {
		resp.Diagnostics.AddError("Unexpected response identity", "The returned locale ID differs from the requested ID.")

		return
	}

	item, diagnostics := newLocaleDataSourceItem(*entity, data.SpaceID.ValueString(), data.EnvironmentID.ValueString())
	resp.Diagnostics.Append(diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.LocaleDataSourceItemModel = item
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(data.SpaceID.ValueString(), data.EnvironmentID.ValueString(), data.LocaleID.ValueString())

	if ctx.Err() != nil {
		resp.Diagnostics.AddError("Failed to read locale", ctx.Err().Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
