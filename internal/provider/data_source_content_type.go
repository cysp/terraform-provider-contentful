package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource              = (*contentTypeDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*contentTypeDataSource)(nil)
)

//nolint:ireturn
func NewContentTypeDataSource() datasource.DataSource { return &contentTypeDataSource{} }

type contentTypeDataSource struct{ providerData ContentfulProviderData }

func (d *contentTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_type"
}

func (d *contentTypeDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ContentTypeDataSourceSchema(ctx)
}

func (d *contentTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *contentTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContentTypeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(requireDiscoveryID(path.Root("space_id"), data.SpaceID)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("environment_id"), data.EnvironmentID)...)
	resp.Diagnostics.Append(requireDiscoveryID(path.Root("content_type_id"), data.ContentTypeID)...)

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

	response, err := d.providerData.client.GetContentType(ctx, cm.GetContentTypeParams{
		SpaceID:       data.SpaceID.ValueString(),
		EnvironmentID: data.EnvironmentID.ValueString(),
		ContentTypeID: data.ContentTypeID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read content type", util.ErrorDetailFromContentfulManagementResponse(response, err))

		return
	}

	entity, ok := response.(*cm.ContentType)
	if !ok {
		resp.Diagnostics.AddError("Failed to read content type", contentfulListNonCollectionResponseDetail(response))

		return
	}

	if entity.Sys.ID != data.ContentTypeID.ValueString() {
		resp.Diagnostics.AddAttributeError(path.Root("content_type_id"), "Unexpected response identity", "The returned Content Type ID differs from the requested ID.")

		return
	}

	item, diagnostics := newContentTypeDataSourceItem(ctx, path.Empty(), *entity, data.SpaceID.ValueString(), data.EnvironmentID.ValueString())
	resp.Diagnostics.Append(diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ContentTypeDataSourceItemModel = item
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(data.SpaceID.ValueString(), data.EnvironmentID.ValueString(), data.ContentTypeID.ValueString())

	if ctx.Err() != nil {
		resp.Diagnostics.AddError("Failed to read content type", ctx.Err().Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
