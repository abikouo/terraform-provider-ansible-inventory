package provider

import (
    "context"
    "fmt"
	"slices"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)


// Ensure the implementation satisfies the expected interfaces.
var (
    _ datasource.DataSource              = &inventoryDataSource{}
    _ datasource.DataSourceWithConfigure = &inventoryDataSource{}
)

// NewInventoryDataSource is a helper function to simplify the provider implementation.
func NewInventoryDataSource() datasource.DataSource {
    return &inventoryDataSource{}
}

// inventoryDataSource is the data source implementation.
type inventoryDataSource struct {
    client *hashicups.Client
}

// Metadata returns the data source type name.
func (d *inventoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_inventory"
}

// Schema defines the schema for the data source.
func (d *inventoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
			"meta": map[string]schema.Attribute{
				"hostvars": schema.MapAttribute{
					Required: false,
					Computed: true,
					ElementType: types.MapType{
						ElemType: types.StringType,
						Computed: true,
					}
				}
            },
			"groups": map[string]schema.Attribute{
				"hosts": schema.ListAttribute{
					ElementType: types.StringType,
					Required: true,
					Computed: true,
				}
            }
        }
    }
}


// Read refreshes the Terraform state with the latest data.
func (d *inventoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var state inventoryDataSourceModel

    hosts, err := d.client.GetHosts()
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Read Ansible hosts",
            err.Error(),
        )
        return
    }

    // Map response
    for _, host := range hosts {
		groups, ok := host.Groups
		if ok {
			for _, group := range groups {
				state_groups, ok := state.Groups[group]
				if ! ok {
					state.Groups[group] := groupsModel{
						Hosts: 		[],
					}
				}
				if ! slices.Contains(state.Groups[group].Hosts, host.Name){
					state.Groups[group].Hosts = append(state.Groups[group].Hosts, host.Name)
				}
			}
		}
	}

    // Set state
    diags := resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}


// Configure adds the provider configured client to the data source.
func (d *inventoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*hashicups.Client)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Data Source Configure Type",
            fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )

        return
    }

    d.client = client
}

// inventoryDataSourceModel maps the data source schema data.
type inventoryDataSourceModel struct {
    Meta   		metaModel 					`tfsdk:"meta"`
	Groups 		map[string]groupsModel 		`tfsdk:"groups"`
}

// groupsModel maps groups schema data.
type groupsModel struct {
    Hosts          []types.String       `tfsdk:"hosts"`
}

// metaModel maps meta schema data
type metaModel struct {
    HostVars       types.ObjectType      `tfsdk:"hostvars"`
}
