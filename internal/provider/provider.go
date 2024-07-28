package provider

import (
	"context"
	"net/http"
	"os"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/provider_contentful"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = (*ContentfulProvider)(nil)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ContentfulProvider{
			version: version,
		}
	}
}

type ContentfulProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *ContentfulProvider) Schema(ctx context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = provider_contentful.ContentfulProviderSchema(ctx)
}

func (p *ContentfulProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data provider_contentful.ContentfulModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var contentfulURL string
	if !data.Url.IsNull() {
		contentfulURL = data.Url.ValueString()
	} else if contentfulURLFromEnv, found := os.LookupEnv("CONTENTFUL_URL"); found {
		contentfulURL = contentfulURLFromEnv
	} else {
		contentfulURL = contentfulManagement.DefaultServerURL
	}

	if contentfulURL == "" {
		resp.Diagnostics.AddAttributeError(path.Root("url"), "Failed to configure client", "No API URL provided")
	}

	var accessToken string
	if !data.AccessToken.IsNull() {
		accessToken = data.AccessToken.ValueString()
	} else {
		if accessTokenFromEnv, found := os.LookupEnv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN"); found {
			accessToken = accessTokenFromEnv
		}
	}

	if accessToken == "" {
		resp.Diagnostics.AddAttributeError(path.Root("access_token"), "Failed to configure client", "No access token provided")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	contentfulManagementClient, err := contentfulManagement.NewClient(
		contentfulURL,
		contentfulManagement.NewAccessTokenSecuritySource(accessToken),
		contentfulManagement.WithClient(util.NewClientWithUserAgent(http.DefaultClient, "terraform-provider-contentful/"+p.version)),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Contentful client: %s", err.Error())
	}

	resp.DataSourceData = ContentfulProviderData{client: contentfulManagementClient}
	resp.ResourceData = ContentfulProviderData{client: contentfulManagementClient}
}

func (p *ContentfulProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "contentful"
	resp.Version = p.version
}

func (p *ContentfulProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *ContentfulProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAppInstallationResource,
		NewEditorInterfaceResource,
	}
}
