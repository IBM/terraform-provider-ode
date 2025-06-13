// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	odeclient "github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/client"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/target"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/provider/validation"
)

// odeTargetResource implements resource.Resource and resource.ResourceWithConfigure.
var (
	_ resource.Resource              = &odeTargetResource{}
	_ resource.ResourceWithConfigure = &odeTargetResource{}
)

type odeTargetResource struct {
	client *odeclient.Client
}

func NewODETargetResource() resource.Resource {
	return &odeTargetResource{}
}

func (r *odeTargetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_target"
}

func (r *odeTargetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*odeclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *ode.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}
	r.client = client
}

func (r *odeTargetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an ODE Target Environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique UUID of the target environment",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Target creation status (e.g. SUCCEEDED, FAILED).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"label": schema.StringAttribute{
				Required:    true,
				Description: "Label for the target environment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Optional free form description of the target.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"hostname": schema.StringAttribute{
				Required:    true,
				Description: "Linux host or IP for the target.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssh_port": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "SSH port to connect to the Linux host. Default is 22.",
				Default:     int64default.StaticInt64(22),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
			},

			"ssh_target_user": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Description: "Linux username for SSH access. This can also be sourced from the `SSH_TARGET_USER` environment variable.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"ssh_target_password": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Description: "Linux password for SSH access (Must specify if not using key). This can also be sourced from the `SSH_TARGET_PASSWORD` environment variable.",
			},

			"ssh_target_key_file": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Description: "Linux private key file content (Must specify if using key based SSH). This can also be sourced from the `SSH_KEY_FILE_PATH` environment variable.",
			},

			"ssh_target_passphrase": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Description: "Passphrase for the SSH key file, if encrypted. This can also be sourced from the `SSH_TARGET_PASSPHRASE` environment variable.",
			},

			"ic_port": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Linux environment instance controller port number (default 8443).",
				Default:     int64default.StaticInt64(8443),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"install_dir": schema.StringAttribute{
				Optional:    true,
				Description: "Directory for ODE files. Maps to `download directory`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"concurrent_transfers": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of concurrent volume transfers. Must be >= 1. Default is 1",
				Default:     int64default.StaticInt64(1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"resume": schema.BoolAttribute{
				Computed:    true,
				Description: "Resume a previously failed or in progress action.",
				Default:     booldefault.StaticBool(false),
			},
			"dns_ip_primary": schema.StringAttribute{
				Required:    true,
				Description: "The IP address of the name server serving the Linux environment..",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"dns_domain_origin": schema.StringAttribute{
				Required:    true,
				Description: "The domain origin that will be appended to host names passed to the resolver.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"terminal_port_start": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The starting port of an available range for accessing 3270 terminal.",
				Default:     int64default.StaticInt64(3270),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"terminal_port_end": schema.Int64Attribute{
				Optional:    true,
				Description: "Ending port of an available range for 3270 terminal access.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"privilege_command_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the privilege escalation command, if required.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"network_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Always 'IPTABLE' for this resource.",
				Default:     stringdefault.StaticString("IPTABLE"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},

			"iptable_setting": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Settings specific to iptables based networking configuration.",
				Attributes: map[string]schema.Attribute{
					"zos_ip_address": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The emulated z/OS IP address. Default is 172.26.1.2",
						Default:     stringdefault.StaticString("172.26.1.2"),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"zos_ssh_route_port": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "The Linux port that routes to z/OS ssh. Default is 2022",
						Default:     int64default.StaticInt64(2022),
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplaceIfConfigured(),
						},
					},

					"tcp_forward_ports": schema.ListNestedAttribute{
						Required:    true,
						Description: "The port forwarding rules of iptables for TCP protocols.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplaceIfConfigured(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"start_port": schema.Int64Attribute{
									Required:    true,
									Description: "Starting port.",
								},
								"end_port": schema.Int64Attribute{
									Required:    true,
									Description: "Ending port",
								},
							},
						},
					},

					"udp_forward_ports": schema.ListNestedAttribute{
						Required:    true,
						Description: "The port forwarding rules of iptables for UDP protocols.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplaceIfConfigured(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"start_port": schema.Int64Attribute{
									Required:    true,
									Description: "Starting port.",
								},
								"end_port": schema.Int64Attribute{
									Required:    true,
									Description: "Ending port",
								},
							},
						},
					},

					"tcp_reroute_ports": schema.ListNestedAttribute{
						Required:    true,
						Description: "The port rerouting rules of iptables for TCP protocols.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplaceIfConfigured(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"linux_port": schema.Int64Attribute{
									Required:    true,
									Description: "Linux port.",
								},
								"zos_port": schema.Int64Attribute{
									Required:    true,
									Description: "z/OS port",
								},
							},
						},
					},

					"udp_reroute_ports": schema.ListNestedAttribute{
						Required:    true,
						Description: "The port rerouting rules of iptables for UDP protocols.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplaceIfConfigured(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"linux_port": schema.Int64Attribute{
									Required:    true,
									Description: "Linux port.",
								},
								"zos_port": schema.Int64Attribute{
									Required:    true,
									Description: "z/OS port",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *odeTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OdeTargetModel
	var sshPassword, sshUsername, sshKeyFile, sshPassphrase string

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, err := PlanToInput(plan)
	if err != nil {
		resp.Diagnostics.AddError("Plan Mapping Failed", err.Error())
		return
	}

	// get writeOnly values from config
	req.Config.GetAttribute(ctx, path.Root("ssh_target_password"), &sshPassword)
	req.Config.GetAttribute(ctx, path.Root("ssh_target_user"), &sshUsername)
	req.Config.GetAttribute(ctx, path.Root("ssh_target_key_file"), &sshKeyFile)
	req.Config.GetAttribute(ctx, path.Root("ssh_target_passphrase"), &sshPassphrase)

	// set input.auth values
	if sshPassword == "" {
		sshPassword = os.Getenv("SSH_TARGET_PASSWORD")
	}
	input.Auth.Password = sshPassword
	// set sshPassphrase
	if sshPassphrase == "" {
		sshPassphrase = os.Getenv("SSH_TARGET_PASSPHRASE")
	}
	input.Auth.Passphrase = &sshPassphrase
	// set key_file
	var sshKeyFileBytes []byte
	if sshKeyFile != "" {
		sshKeyFileBytes = []byte(sshKeyFile)
	} else {
		sshKeyFile_env := os.Getenv("SSH_KEY_FILE_PATH")
		sshKeyFileBytes = []byte(sshKeyFile_env)
	}
	input.Auth.KeyFile = &sshKeyFileBytes
	// set username
	if sshUsername == "" {
		sshUsername = os.Getenv("SSH_TARGET_USER")
	}
	input.Auth.Username = sshUsername

	// make createIPTables call
	targetUUID, err := r.client.Target.CreateIPTables(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	// Fetch the complete resource details
	resourceDetails, err := r.client.Target.Get(ctx, targetUUID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading resource after creation", err.Error())
		return
	}

	// Populate the state with complete resource details
	plan.ID = types.StringValue(resourceDetails.UUID)
	plan.Status = types.StringValue(resourceDetails.Status)
	plan.Hostname = types.StringValue(resourceDetails.Hostname)
	plan.Label = types.StringValue(resourceDetails.Label)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *odeTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OdeTargetModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	targ, err := r.client.Target.Get(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	if targ == nil || targ.UUID == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state = OdeTargetModel{
		ID:                  types.StringValue(targ.UUID),
		Status:              types.StringValue(targ.Status),
		Hostname:            types.StringValue(targ.Hostname),
		Label:               types.StringValue(targ.Label),
		Description:         types.StringValue(targ.Description),
		SSHPort:             types.Int64Value(int64(targ.SSHPort)),
		ICPort:              types.Int64Value(int64(targ.ICPort)),
		InstallDir:          types.StringValue(targ.DownloadDirectory),
		ConcurrentTransfers: types.Int64Value(int64(targ.ConcurrentTransfers)),
		DNSIPPrimary:        types.StringValue(targ.DNSIPPrimary),
		DNSDomainOrigin:     types.StringValue(targ.DNSDomainOrigin),
		TerminalPortStart:   types.Int64Value(int64(targ.TerminalPortStart)),
		NetworkType:         types.StringValue(targ.NetworkType),
		Resume:              types.BoolValue(targ.Resume),

		IPTablesSetting: &IptablesSettingModel{
			ZosIPAddress:    types.StringValue(targ.ZosIPAddress),
			ZosSSHRoutePort: types.Int64Value(int64(targ.ZosSSHRoutePort)),
		},
	}

	for _, TCPForwardPort := range targ.TCPForwardPorts {
		state.IPTablesSetting.TCPForwardPorts = append(state.IPTablesSetting.TCPForwardPorts, ForwardPortModel{
			StartPort: types.Int64Value(int64(TCPForwardPort.StartPort)),
			EndPort:   types.Int64Value(int64(TCPForwardPort.EndPort)),
		})
	}

	for _, UDPForwardPort := range targ.UDPForwardPorts {
		state.IPTablesSetting.UDPForwardPorts = append(state.IPTablesSetting.UDPForwardPorts, ForwardPortModel{
			StartPort: types.Int64Value(int64(UDPForwardPort.StartPort)),
			EndPort:   types.Int64Value(int64(UDPForwardPort.EndPort)),
		})
	}

	for _, TCPReroutePort := range targ.TCPReroutePorts {
		state.IPTablesSetting.TCPReroutePorts = append(state.IPTablesSetting.TCPReroutePorts, ReroutePortModel{
			LinuxPort: types.Int64Value(int64(TCPReroutePort.LinuxPort)),
			ZosPort:   types.Int64Value(int64(TCPReroutePort.ZosPort)),
		})
	}

	for _, UDPReroutePort := range targ.UDPReroutePorts {
		state.IPTablesSetting.UDPReroutePorts = append(state.IPTablesSetting.UDPReroutePorts, ReroutePortModel{
			LinuxPort: types.Int64Value(int64(UDPReroutePort.LinuxPort)),
			ZosPort:   types.Int64Value(int64(UDPReroutePort.ZosPort)),
		})
	}

	// refresh state
	resp.State.Set(ctx, &state)
}

func (r *odeTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OdeTargetModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Target.Delete(ctx, state.ID.ValueString(), false, false)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
	}
}

func (r *odeTargetResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {

}
func (r *odeTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

type OdeTargetModel struct {
	ID                   types.String `tfsdk:"id"`
	Label                types.String `tfsdk:"label"`
	Description          types.String `tfsdk:"description"`
	Hostname             types.String `tfsdk:"hostname"`
	SSHPort              types.Int64  `tfsdk:"ssh_port"`
	ICPort               types.Int64  `tfsdk:"ic_port"`
	InstallDir           types.String `tfsdk:"install_dir"`
	ConcurrentTransfers  types.Int64  `tfsdk:"concurrent_transfers"`
	DNSIPPrimary         types.String `tfsdk:"dns_ip_primary"`
	DNSDomainOrigin      types.String `tfsdk:"dns_domain_origin"`
	TerminalPortStart    types.Int64  `tfsdk:"terminal_port_start"`
	TerminalPortEnd      types.Int64  `tfsdk:"terminal_port_end"`
	NetworkType          types.String `tfsdk:"network_type"`
	Resume               types.Bool   `tfsdk:"resume"`
	PrivilegeCommandUUID types.String `tfsdk:"privilege_command_uuid"`

	// Sensitive Auth Inputs
	SSHTargetUser       types.String `tfsdk:"ssh_target_user"`
	SSHTargetPassword   types.String `tfsdk:"ssh_target_password"`
	SSHTargetKeyFile    types.String `tfsdk:"ssh_target_key_file"`
	SSHTargetPassphrase types.String `tfsdk:"ssh_target_passphrase"`

	Status types.String `tfsdk:"status"`

	IPTablesSetting *IptablesSettingModel `tfsdk:"iptable_setting"`
}

type IptablesSettingModel struct {
	ZosIPAddress    types.String       `tfsdk:"zos_ip_address"`
	ZosSSHRoutePort types.Int64        `tfsdk:"zos_ssh_route_port"`
	TCPForwardPorts []ForwardPortModel `tfsdk:"tcp_forward_ports"`
	UDPForwardPorts []ForwardPortModel `tfsdk:"udp_forward_ports"`
	TCPReroutePorts []ReroutePortModel `tfsdk:"tcp_reroute_ports"`
	UDPReroutePorts []ReroutePortModel `tfsdk:"udp_reroute_ports"`
}

type ForwardPortModel struct {
	StartPort types.Int64 `tfsdk:"start_port"`
	EndPort   types.Int64 `tfsdk:"end_port"`
}

type ReroutePortModel struct {
	LinuxPort types.Int64 `tfsdk:"linux_port"`
	ZosPort   types.Int64 `tfsdk:"zos_port"`
}

func PlanToInput(plan OdeTargetModel) (target.CreateTargetInput, error) {

	if plan.IPTablesSetting == nil {
		return target.CreateTargetInput{}, fmt.Errorf("missing required iptable_setting block")
	}

	auth := target.SSHCredentials{}

	iptableConfig := target.IPTablesSetting{
		ZosIPAddress:    plan.IPTablesSetting.ZosIPAddress.ValueString(),
		ZosSSHRoutePort: int(plan.IPTablesSetting.ZosSSHRoutePort.ValueInt64()),
		TCPForwardPorts: MapForwardPorts(plan.IPTablesSetting.TCPForwardPorts),
		UDPForwardPorts: MapForwardPorts(plan.IPTablesSetting.UDPForwardPorts),
		TCPReroutePorts: MapReroutePorts(plan.IPTablesSetting.TCPReroutePorts),
		UDPReroutePorts: MapReroutePorts(plan.IPTablesSetting.UDPReroutePorts),
	}

	req := target.IPTablesRequest{
		Label:       plan.Label.ValueString(),
		Description: plan.Description.ValueString(),
		Hostname:    plan.Hostname.ValueString(),
		SSHPort:     int(plan.SSHPort.ValueInt64()),
		ICPort:      int(plan.ICPort.ValueInt64()),

		DownloadDirectory:   plan.InstallDir.ValueString(),
		DNSIPPrimary:        plan.DNSIPPrimary.ValueString(),
		DNSDomainOrigin:     plan.DNSDomainOrigin.ValueString(),
		ConcurrentTransfers: int(plan.ConcurrentTransfers.ValueInt64()),
		Resume:              plan.Resume.ValueBool(),
		TerminalPortStart:   int(plan.TerminalPortStart.ValueInt64()),
		NetworkType:         plan.NetworkType.ValueString(),
		IPTablesSetting:     iptableConfig,
	}

	return target.CreateTargetInput{
		Request: req,
		Auth:    auth,
	}, nil
}

func MapForwardPorts(in []ForwardPortModel) []target.ForwardPortRange {
	out := make([]target.ForwardPortRange, len(in))
	for i, p := range in {
		out[i] = target.ForwardPortRange{
			StartPort: int(p.StartPort.ValueInt64()),
			EndPort:   int(p.EndPort.ValueInt64()),
		}
	}
	return out
}

func MapReroutePorts(in []ReroutePortModel) []target.ReroutePortMapping {
	out := make([]target.ReroutePortMapping, len(in))
	for i, p := range in {
		out[i] = target.ReroutePortMapping{
			LinuxPort: int(p.LinuxPort.ValueInt64()),
			ZosPort:   int(p.ZosPort.ValueInt64()),
		}
	}
	return out
}

func (r *odeTargetResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data OdeTargetModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	validation.ValidateExactlyOneOf(
		&resp.Diagnostics,
		validation.StringAttribute{
			Path:   path.Root("ssh_target_password"),
			Value:  data.SSHTargetPassword,
			EnvVar: SSHPassword,
		},
		validation.StringAttribute{
			Path:   path.Root("ssh_target_key_file"),
			Value:  data.SSHTargetKeyFile,
			EnvVar: SSHKeyFile,
		},
	)
	validation.ValidateAllSet(
		&resp.Diagnostics,
		validation.StringAttribute{
			Path:   path.Root("ssh_target_user"),
			Value:  data.SSHTargetUser,
			EnvVar: SSHUser,
		},
	)
}
