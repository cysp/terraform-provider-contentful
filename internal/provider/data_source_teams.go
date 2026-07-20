package provider

import (
	"cmp"
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = (*teamsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*teamsDataSource)(nil)
)

//nolint:ireturn
func NewTeamsDataSource() datasource.DataSource {
	return &teamsDataSource{}
}

type teamsDataSource struct {
	providerData ContentfulProviderData
}

func (d *teamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (d *teamsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = TeamsDataSourceSchema(ctx)
}

func (d *teamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	resp.Diagnostics.Append(SetProviderDataFromDataSourceConfigureRequest(req, &d.providerData)...)
}

func (d *teamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TeamsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

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

	teams, diagnostics := readTeams(ctx, d.providerData.client, data.OrganizationID.ValueString())
	resp.Diagnostics.Append(diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(data.OrganizationID.ValueString())
	data.Teams = teams

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func readTeams(ctx context.Context, client *cm.Client, organizationID string) ([]TeamsDataSourceTeamModel, diag.Diagnostics) {
	var (
		skip  int64
		teams = make([]TeamsDataSourceTeamModel, 0)
	)

	for {
		params := cm.GetTeamsParams{
			OrganizationID: organizationID,
			Skip:           cm.NewOptInt64(skip),
			Limit:          cm.NewOptInt64(defaultPageLimit),
		}

		response, err := client.GetTeams(ctx, params)

		tflog.Info(ctx, "teams.read", map[string]any{
			"params":   params,
			"response": response,
			"err":      err,
		})

		if err != nil {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to read teams", util.ErrorDetailFromContentfulManagementResponse(response, err)),
			}
		}

		collection, ok := response.(*cm.TeamCollection)
		if !ok {
			return nil, diag.Diagnostics{
				diag.NewErrorDiagnostic("Failed to read teams", contentfulListNonCollectionResponseDetail(response)),
			}
		}

		for _, team := range collection.Items {
			teams = append(teams, TeamsDataSourceTeamModel{
				TeamID:      types.StringValue(team.Sys.ID),
				Name:        types.StringValue(team.Name),
				Description: types.StringPointerValue(team.Description.ValueStringPointer()),
			})
		}

		skip += int64(len(collection.Items))
		total, totalSet := collection.Total.Get()

		if len(collection.Items) == 0 || (totalSet && skip >= int64(total)) {
			break
		}
	}

	// Contentful does not define collection order; canonicalize Terraform's ordered list.
	slices.SortFunc(teams, func(a, b TeamsDataSourceTeamModel) int {
		return cmp.Compare(a.TeamID.ValueString(), b.TeamID.ValueString())
	})

	return teams, nil
}
