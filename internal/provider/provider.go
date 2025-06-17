// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/client/tls"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	odeclient "github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/client"
	odecfg "github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/config"
)

const (
	envTLSCAFile             = "ODE_TLS_CA_FILE"
	envTLSInsecureSkipVerify = "ODE_TLS_INSECURE_SKIP_VERIFY"
	envTLSServerName         = "ODE_TLS_SERVER_NAME"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var (
	_ provider.Provider = &odeProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &odeProvider{
			version: version,
		}
	}
}

// odeProviderModel maps provider schema data to a Go type.
type odeProviderModel struct {
	Username types.String `tfsdk:"ode_username"`
	Password types.String `tfsdk:"ode_password"`
	Host     types.String `tfsdk:"ode_host"`
	TLS      *odeTlsModel `tfsdk:"ode_tls"`
}

// odeTlsModel maps the nested ode_tls schema.
type odeTlsModel struct {
	CAFile             types.String `tfsdk:"ca_file"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	ServerName         types.String `tfsdk:"server_name"`
}

// odeProvider is the provider implementation.
type odeProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *odeProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ode"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *odeProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ode_username": schema.StringAttribute{
				Optional:    true,
				Description: "Username for On-Demand Environments authentication. Alternatively, this value can be sourced from the `ODE_USERNAME` environment variable.",
			},
			"ode_password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for On-Demand Environments authentication. Alternatively, this value can be sourced from the `ODE_PASSWORD` environment variable.",
			},
			"ode_host": schema.StringAttribute{
				Optional:    true,
				Description: "Host for On-Demand Environments. Alternatively, this value can be sourced from the `ODE_HOST` environment variable.",
			},
			"ode_tls": TLSAttribute(),
		},
	}
}

func TLSAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:            true,
		Description:         "Certificate for On-Demand Environments. Optional only when defined through environment variables.",
		MarkdownDescription: "Certificate for On-Demand Environments. Optional only when defined through environment variables.",
		Attributes: map[string]schema.Attribute{
			"ca_file": schema.StringAttribute{
				Optional:            true,
				Description:         "CA file for On-Demand Environments. Alternatively, this value can be sourced from the `ODE_TLS_CA_FILE` environment variable.",
				MarkdownDescription: "CA file for On-Demand Environments. Alternatively, this value can be sourced from the `ODE_TLS_CA_FILE` environment variable.",
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Optional:            true,
				Description:         "Insecure SSL certificate for On-Demand Environments. Alternatively, this value can be sourced from the `ODE_TLS_INSECURE_SKIP_VERIFY` environment variable.",
				MarkdownDescription: "Insecure SSL certificate for On-Demand Environments. Alternatively, this value can be sourced from the `ODE_TLS_INSECURE_SKIP_VERIFY` environment variable.",
			},
			"server_name": schema.StringAttribute{
				Optional:            true,
				Description:         "Server name for On-Demand Environments. Alternatively, this value can be sourced from the `ODE_TLS_SERVER_NAME` environment variable.",
				MarkdownDescription: "Server name for On-Demand Environments. Alternatively, this value can be sourced from the `ODE_TLS_SERVER_NAME` environment variable.",
			},
		},
	}
}

func GetStringAttr(attr types.String, envKey, defaultValue string) string {
	if !attr.IsNull() && !attr.IsUnknown() {
		if value := attr.ValueString(); value != "" {
			return value
		}
	}
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	return defaultValue
}

func GetBoolAttr(attr types.Bool, envKey string, defaultValue bool) bool {
	if !attr.IsNull() && !attr.IsUnknown() {
		return attr.ValueBool()
	}
	if envValue := os.Getenv(envKey); envValue != "" {
		return strings.ToLower(envValue) == "true"
	}
	return defaultValue
}

func (p *odeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	tflog.Info(ctx, "Configuring Terraformz Go client")
	var config odeProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("ode_username"),
			"Unknown Username",
			"Unknown configuration value for On-Demand Environments username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ODE_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("ode_password"),
			"Unknown Password",
			"Unknown configuration value for On-Demand Environments password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ODE_PASSWORD environment variable.",
		)
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("ode_host"),
			"Unknown Host",
			"Unknown configuration value for On-Demand Environments host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ODE_HOST environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	username := os.Getenv("ODE_USERNAME")
	password := os.Getenv("ODE_PASSWORD")
	host := os.Getenv("ODE_HOST")

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// TLS
	var ca string
	var skipInsecure bool
	var sni string

	if config.TLS == nil {
		config.TLS = &odeTlsModel{}
	}
	ca = GetStringAttr(config.TLS.CAFile, envTLSCAFile, "")
	sni = GetStringAttr(config.TLS.ServerName, envTLSServerName, "")
	skipInsecure = GetBoolAttr(config.TLS.InsecureSkipVerify, envTLSInsecureSkipVerify, false)

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("ode_username"),
			"Missing Username",
			"The provider cannot create the On-Demand Environments API client as there is a missing or empty value for the username. "+
				"Set the username value in the configuration or use the ODE_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("ode_password"),
			"Missing Password",
			"The provider cannot create the On-Demand Environments API client as there is a missing or empty value for the password. "+
				"Set the password value in the configuration or use the ODE_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("ode_host"),
			"Missing Host",
			"The provider cannot create the On-Demand Environments API client as there is a missing or empty value for the host. "+
				"Set the host value in the configuration or use the ODE_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// TLS configuration
	tlsParams := &tls.Params{CA: ca, SkipVerify: skipInsecure, SNI: sni}
	tlsConfig, err := tls.NewBuilder(tlsParams).Build()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create TLS Configuration",
			"The provider failed to build a TLS configuration required to connect "+
				"to the On-Demand Environments API.\n\n"+
				"Underlying error: "+err.Error(),
		)
		return
	}

	cfg := odecfg.Config{
		BaseURL:   host,
		User:      username,
		Pass:      password,
		TLSConfig: tlsConfig,
	}

	// Create a terraformz go client using the configuration values
	c, err := odeclient.New(&cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create ODE API Client",
			"An unexpected error occurred when creating the On-Demand Environments API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"ODE Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c

}

// Resources defines the resources implemented in the provider.
func (p *odeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProvisionedInstanceResource,
		NewODETargetResource,
	}
}

func (p *odeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOdeImageDataSource,
	}
}
