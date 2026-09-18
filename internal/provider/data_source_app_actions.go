package provider

import (
	"cmp"
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppActionsDataSourceModel struct {
	IDIdentityModel

	OrganizationID  types.String      `tfsdk:"organization_id"`
	AppDefinitionID types.String      `tfsdk:"app_definition_id"`
	AppActions      []AppActionFields `tfsdk:"app_actions"`
	Timeouts        timeouts.Value    `tfsdk:"timeouts"`
}

var (
	_ datasource.DataSource              = (*appActionsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*appActionsDataSource)(nil)
)

//nolint:ireturn
func NewAppActionsDataSource() datasource.DataSource { return &appActionsDataSource{} }

type appActionsDataSource struct{ providerData ContentfulProviderData }

func (d *appActionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_actions"
}

func (d *appActionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = AppActionsDataSourceSchema(ctx)
}

func (d *appActionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *appActionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AppActionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	target := AppActionBaseModel{OrganizationID: data.OrganizationID, AppDefinitionID: data.AppDefinitionID}
	resp.Diagnostics.Append(appActionScope(target)...)

	timeout, diags := data.Timeouts.Read(ctx, defaultResourceOperationTimeout)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	const title = "Failed to read app actions"

	items, diags := readContentfulCollection(ctx, title,
		func(ctx context.Context, skip int64) (contentfulCollection[cm.AppAction], diag.Diagnostics) {
			response, err := d.providerData.client.GetAppActions(ctx, cm.GetAppActionsParams{OrganizationID: target.OrganizationID.ValueString(), AppDefinitionID: target.AppDefinitionID.ValueString(), Skip: cm.NewOptInt64(skip), Limit: cm.NewOptInt64(defaultPageLimit)})
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(title, util.ErrorDetailFromContentfulManagementResponse(response, err))}
			}

			switch response := response.(type) {
			case *cm.AppActionCollection:
				return response, nil
			default:
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic(title, contentfulListNonCollectionResponseDetail(response))}
			}
		},
		func(entity cm.AppAction) (AppActionFields, diag.Diagnostics) {
			item, diags := appActionResponse(entity, target)

			return item.AppActionFields, diags
		})
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	slices.SortFunc(items, func(a, b AppActionFields) int {
		return cmp.Compare(a.AppActionID.ValueString(), b.AppActionID.ValueString())
	})

	data.AppActions = items
	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(data.OrganizationID.ValueString(), data.AppDefinitionID.ValueString())

	if ctx.Err() != nil {
		resp.Diagnostics.AddError(title, ctx.Err().Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
