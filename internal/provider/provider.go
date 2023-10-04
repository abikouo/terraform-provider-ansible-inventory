package provider

import (
	"context"
	"os"
    "strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
    _ provider.Provider = &ansibleInventoryProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
    return func() provider.Provider {
        return &ansibleInventoryProvider{
            version: version,
        }
    }
}

// ansibleInventoryProvider is the provider implementation.
type ansibleInventoryProvider struct {
    // version is set to the provider version on release, "dev" when the
    // provider is built and ran locally, and "test" when running acceptance
    // testing.
    version string
}

// Metadata returns the provider type name.
func (p *ansibleInventoryProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
    resp.TypeName = "ansibleinventory"
    resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *ansibleInventoryProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "host": schema.StringAttribute{
                Required: true,
            },
            "username": schema.StringAttribute{
                Optional: true,
            },
            "password": schema.StringAttribute{
                Optional:  true,
                Sensitive: true,
            },
            "insecure_skip_verify": schema.StringAttribute{
                Optional:   true,
                Default:    false,
            },
        },
    }
}

func (p *ansibleInventoryProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    // Retrieve provider data from configuration
    var config ansibleInventoryProviderModel
    diags := req.Config.Get(ctx, &config)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // If practitioner provided a configuration value for any of the
    // attributes, it must be a known value.

    if config.Host.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("host"),
            "Unknown Controller API Host",
            "The provider cannot create the Controller API client as there is an unknown configuration value for the Controller API host. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the CONTROLLER_HOST environment variable.",
        )
    }

    if config.Username.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("username"),
            "Unknown Controller API Username",
            "The provider cannot create the Controller API client as there is an unknown configuration value for the Controller API username. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the CONTROLLER_USERNAME environment variable.",
        )
    }

    if config.Password.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("password"),
            "Unknown Controller API Password",
            "The provider cannot create the Controller API client as there is an unknown configuration value for the Controller API password. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the CONTROLLER_PASSWORD environment variable.",
        )
    }

    if config.InsecureSkipVerify.IsUnknown() {
        resp.Diagnostics.AddAttributeError(
            path.Root("insecure_skip_verify"),
            "Unknown Controller API insecure_skip_verify",
            "The provider cannot create the Controller API client as there is an unknown configuration value for the Controller API insecure_skip_verify. "+
                "Either target apply the source of the value first, set the value statically in the configuration, or use the CONTROLLER_INSECURE_SKIP_VERIFY environment variable.",
        )
    }

    if resp.Diagnostics.HasError() {
        return
    }

    // Default values to environment variables, but override
    // with Terraform configuration value if set.

    host := os.Getenv("CONTROLLER_HOST")
    username := os.Getenv("CONTROLLER_USERNAME")
    password := os.Getenv("CONTROLLER_PASSWORD")
    var insecure_skip_verify bool = false
    raw_insecure_skip_verify := os.Getenv("CONTROLLER_INSECURE_SKIP_VERIFY")
    if raw_insecure_skip_verify != "" {
        insecure_skip_verify, err := strconv.ParseBool(raw_insecure_skip_verify)
        if err != nil {
            resp.Diagnostics.AddAttributeError(
                path.Root("insecure_skip_verify"),
                "Invalid value for insecure_skip_verify",
                "The provider cannot create the Controller API client as the value provided for insecure_skip_verify is not a valid boolean.",
            )
            return
        }
    }

    if !config.Host.IsNull() {
        host = config.Host.ValueString()
    }

    if !config.Username.IsNull() {
        username = config.Username.ValueString()
    }

    if !config.Password.IsNull() {
        password = config.Password.ValueString()
    }

    if !config.InsecureSkipVerify.IsNull() {
        insecure_skip_verify = config.InsecureSkipVerify.Value()
    }

    // If any of the expected configurations are missing, return
    // errors with provider-specific guidance.

    if host == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("host"),
            "Missing Controller API Host",
            "The provider cannot create the Controller API client as there is a missing or empty value for the Controller API host. "+
                "Set the host value in the configuration or use the CONTROLLER_HOST environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if username == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("username"),
            "Missing Controller API Username",
            "The provider cannot create the Controller API client as there is a missing or empty value for the Controller API username. "+
                "Set the username value in the configuration or use the CONTROLLER_USERNAME environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if password == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("password"),
            "Missing Controller API Password",
            "The provider cannot create the Controller API client as there is a missing or empty value for the Controller API password. "+
                "Set the password value in the configuration or use the CONTROLLER_PASSWORD environment variable. "+
                "If either is already set, ensure the value is not empty.",
        )
    }

    if resp.Diagnostics.HasError() {
        return
    }

    // Create a new http client using the configuration values
    client, err := NewClient(host, &username, &password, insecure_skip_verify)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Controller API Client",
            "An unexpected error occurred when creating the Controller API client. "+
                "If the error is not clear, please contact the provider developers.\n\n"+
                "http Client Error: "+err.Error(),
        )
        return
    }

    // Make the http client available during DataSource and Resource
    // type Configure methods.
    resp.DataSourceData = client
    resp.ResourceData = client
}


// DataSources defines the data sources implemented in the provider.
func (p *ansibleInventoryProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource {
        NewInventoryDataSource,
    }
}


// Resources defines the resources implemented in the provider.
func (p *ansibleInventoryProvider) Resources(_ context.Context) []func() resource.Resource {
    return nil
}

// ansibleInventoryProviderModel maps provider schema data to a Go type.
type ansibleInventoryProviderModel struct {
    Host                types.String    `tfsdk:"host"`
    Username            types.String    `tfsdk:"username"`
    Password            types.String    `tfsdk:"password"`
    InsecureSkipVerify  types.Bool      `tfsdk:"insecure_skip_verify"`
}

