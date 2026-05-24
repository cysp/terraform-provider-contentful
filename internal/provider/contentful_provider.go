package provider

import (
	"context"
	"math"
	"net/http"
	"net/url"
	"os"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Factory(version string, options ...Option) func() provider.Provider {
	return func() provider.Provider {
		return New(version, options...)
	}
}

func New(version string, options ...Option) *ContentfulProvider {
	provider := ContentfulProvider{
		version: version,
	}

	for _, option := range options {
		option(&provider)
	}

	return &provider
}

type ContentfulProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	contentfulURL string
	httpClient    *http.Client
	accessToken   string
}

type ContentfulProviderModel struct {
	URL         types.String `tfsdk:"url"`
	AccessToken types.String `tfsdk:"access_token"`
}

var _ provider.Provider = (*ContentfulProvider)(nil)

type Option func(*ContentfulProvider)

func WithContentfulURL(url string) Option {
	return func(p *ContentfulProvider) {
		p.contentfulURL = url
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(p *ContentfulProvider) {
		p.httpClient = httpClient
	}
}

func WithAccessToken(accessToken string) Option {
	return func(p *ContentfulProvider) {
		p.accessToken = accessToken
	}
}

func (p *ContentfulProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Contentful spaces and related configuration with Terraform.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "Contentful Management API base URL. Defaults to the public Contentful Management API. Can also be set with the CONTENTFUL_URL environment variable.",
				Optional:    true,
			},
			"access_token": schema.StringAttribute{
				Description: "Contentful Management API access token. Can also be set with the CONTENTFUL_MANAGEMENT_ACCESS_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *ContentfulProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ContentfulProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var contentfulURL string
	if !data.URL.IsNull() && !data.URL.IsUnknown() {
		contentfulURL = data.URL.ValueString()
	} else if contentfulURLFromEnv, found := os.LookupEnv("CONTENTFUL_URL"); found {
		contentfulURL = contentfulURLFromEnv
	}

	if p.contentfulURL != "" {
		contentfulURL = p.contentfulURL
	}

	if contentfulURL == "" {
		contentfulURL = cm.DefaultServerURL
	}

	if contentfulURL == "" {
		resp.Diagnostics.AddAttributeError(path.Root("url"), "Missing Contentful API URL", "Set the url provider attribute or the CONTENTFUL_URL environment variable.")
	} else {
		resp.Diagnostics.Append(validateContentfulURL(contentfulURL)...)
	}

	var accessToken string
	if !data.AccessToken.IsNull() && !data.AccessToken.IsUnknown() {
		accessToken = data.AccessToken.ValueString()
	} else {
		if accessTokenFromEnv, found := os.LookupEnv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN"); found {
			accessToken = accessTokenFromEnv
		}
	}

	if p.accessToken != "" {
		accessToken = p.accessToken
	}

	if accessToken == "" {
		resp.Diagnostics.AddAttributeError(path.Root("access_token"), "Missing Contentful management access token", "Set the access_token provider attribute or the CONTENTFUL_MANAGEMENT_ACCESS_TOKEN environment variable.")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	retryableClient := retryablehttp.NewClient()
	retryableClient.RetryMax = math.MaxInt
	retryableClient.Backoff = util.ContentfulRateLimitLinearJitterBackoff

	if p.httpClient != nil {
		retryableClient.HTTPClient = p.httpClient
	}

	contentfulManagementClient, err := cm.NewClient(
		contentfulURL,
		cm.NewAccessTokenSecuritySource(accessToken),
		cm.WithClient(util.NewClientWithUserAgent(retryableClient.StandardClient(), "terraform-provider-contentful/"+p.version)),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Contentful client", err.Error())

		return
	}

	providerData := ContentfulProviderData{
		client:                       contentfulManagementClient,
		editorInterfaceVersionOffset: &ContentfulContentTypeCounter{},
	}

	resp.ActionData = providerData
	resp.DataSourceData = providerData
	resp.ListResourceData = providerData
	resp.ResourceData = providerData
}

func validateContentfulURL(rawURL string) diag.Diagnostics {
	diags := diag.Diagnostics{}

	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		diags.AddAttributeError(path.Root("url"), "Invalid Contentful API URL", "The url provider attribute must be an absolute HTTP or HTTPS URL, such as https://api.contentful.com. It can also be set with the CONTENTFUL_URL environment variable.")

		return diags
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		diags.AddAttributeError(path.Root("url"), "Invalid Contentful API URL", "The url provider attribute must use the http or https scheme.")
	}

	return diags
}

func (p *ContentfulProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "contentful"
	resp.Version = p.version
}

func (p *ContentfulProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAppDefinitionDataSource,
		NewEnvironmentStatusReadyDataSource,
		NewMarketplaceAppDefinitionDataSource,
		NewPreviewAPIKeyDataSource,
	}
}

func (p *ContentfulProvider) ListResources(_ context.Context) []func() list.ListResource {
	return []func() list.ListResource{
		NewContentTypeListResource,
		NewEntryListResource,
	}
}

func (p *ContentfulProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAppDefinitionResource,
		NewAppSigningSecretResource,
		NewAppInstallationResource,
		NewContentTypeResource,
		NewDeliveryAPIKeyResource,
		NewEditorInterfaceResource,
		NewEnvironmentAliasResource,
		NewEnvironmentResource,
		NewEntryResource,
		NewExtensionResource,
		NewPersonalAccessTokenResource,
		NewResourceProviderResource,
		NewResourceTypeResource,
		NewRoleResource,
		NewSpaceEnablementsResource,
		NewTagResource,
		NewTeamResource,
		NewTeamSpaceMembershipResource,
		NewWebhookResource,
	}
}
