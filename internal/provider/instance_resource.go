// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	odeclient "github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/client"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/instance"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/provider/validation"
)

// provisionedInstanceResource implements resource.Resource and resource.ResourceWithConfigure. We can implement more here
// depending on use case. More about in the tutorial or docs for terraform.
var (
	_ resource.Resource                   = &provisionedInstanceResource{}
	_ resource.ResourceWithConfigure      = &provisionedInstanceResource{}
	_ resource.ResourceWithValidateConfig = &provisionedInstanceResource{}
)

const (
	SSHPassword   = "SSH_TARGET_PASSWORD"
	SSHKeyFile    = "SSH_KEY_FILE_PATH"
	SSHUser       = "SSH_TARGET_USER"
	SSHPassphrase = "SSH_TARGET_PASSPHRASE"
)

// provisionedInstanceResource is the resource implementation.
type provisionedInstanceResource struct {
	client *odeclient.Client
}

func NewProvisionedInstanceResource() resource.Resource {
	return &provisionedInstanceResource{}
}

func (r *provisionedInstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	// Terraform will refer to this resource as: <provider_name>_provisioned_instance
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *provisionedInstanceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the provisioning of a z/OS instance onto a Linux target environment.",
		Attributes: map[string]schema.Attribute{
			// This is the unique provision ID
			"provision_uuid": schema.StringAttribute{
				Description: "The unique provision ID returned by On-Demand Environments after provisioning starts",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				Description: "The IP address assigned to the z/OS instance",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// The current provisioning status
			"status": schema.StringAttribute{
				Description: "Current provision status (e.g., 'in_progress', 'completed', 'failed')",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ssh_target_user": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Description: "Linux target username for SSH access. Alternatively, this value can be sourced from the `SSH_TARGET_USER` environment variable.",
			},
			"ssh_target_password": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Description: "Linux target password for SSH access (Must specify if not using SSH key). Alternatively, this value can be sourced from the `SSH_TARGET_PASSWORD` environment variable.",
			},

			"ssh_target_key_file": schema.StringAttribute{
				Optional:    true,
				WriteOnly:   true,
				Sensitive:   true,
				Description: "Linux private key file content (Must specify if using key based SSH). Alternatively, this value can be sourced from the `SSH_KEY_FILE_PATH` environment variable.",
			},

			"ssh_target_passphrase": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Description: "Passphrase for the SSH key file, if encrypted. Alternatively, this value can be sourced from the `SSH_TARGET_PASSPHRASE` environment variable.",
			},

			// -- NESTED BLOCKS -- //

			// 'general' block
			"general": schema.SingleNestedAttribute{
				Description: "General properties of provision (label, target, image, SSH key, etc.)",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"label": schema.StringAttribute{
						Description: "Label or name for provisioned instance",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"description": schema.StringAttribute{
						Description: "Optional description for the provisioned instance",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"target_uuid": schema.StringAttribute{
						Description: "UUID of the Linux target environment",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"image_uuid": schema.StringAttribute{
						Description: "The UUID of image (stock image only)",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"ssh_public_key": schema.StringAttribute{
						Description: "Public SSH key to inject if provisioning a stock image",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"deployment_directory": schema.StringAttribute{
						Description: "Directory path on the target system for deployment",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"sysres_component_uuid": schema.StringAttribute{
						Description: "UUID of the system component that will be used to provision an instance",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},

			// 'emulator' block
			"emulator": schema.SingleNestedAttribute{
				Description: "Emulator or hardware resource allocations (CPU, RAM, etc.)",
				Optional:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"cp": schema.Int64Attribute{
						Description: "Number of CP (general-purpose) engines",
						Required:    true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
					"ziip": schema.Int64Attribute{
						Description: "Number of zIIP engines to allocate, 0 if none",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
					"ram": schema.Int64Attribute{
						Description: "Amount of RAM in bytes (e.g., 8589934592 for 8 GiB)",
						Required:    true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
				},
			},

			// 'zos_creds' block
			"zos_creds": schema.SingleNestedAttribute{
				Description: "Credentials for z/OS",
				Optional:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "z/OS system username",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"password": schema.StringAttribute{
						Description: "z/OS system password",
						Required:    true,
						Sensitive:   true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},

			// 'ipl' block
			"ipl": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "All IPL values will be used from the image specified ",
				Attributes: map[string]schema.Attribute{
					"device_address": schema.StringAttribute{
						Required:    true,
						Description: "The device address in the devmap of the volume that contains the initial load program of z/OS",
					},
					"iodf_address": schema.StringAttribute{
						Required:    true,
						Description: "The device address where IPL process will look for the IODF data set and SYSn.IPLPARM",
					},
					"load_suffix": schema.StringAttribute{
						Required:    true,
						Description: "The suffix for the LOADxx member in SYSn.IPLPARM or SYS1.PARMLIB",
					},
				},
			},
			"timeouts": timeouts.Attributes(
				ctx,
				timeouts.Opts{
					Create: true,
					Delete: true,
				},
			),
		},
	}
}

// Top-level resource model, maps directly to the schema above
// Data model for state/plan.
type provisionedInstanceResourceModel struct {
	// Computed
	ProvisionUUID types.String `tfsdk:"provision_uuid"`
	Hostname      types.String `tfsdk:"hostname"`
	Status        types.String `tfsdk:"status"`
	// Linux Target Inputs
	SSHTargetUser       types.String `tfsdk:"ssh_target_user"`
	SSHTargetPassword   types.String `tfsdk:"ssh_target_password"`
	SSHTargetKeyFile    types.String `tfsdk:"ssh_target_key_file"`
	SSHTargetPassphrase types.String `tfsdk:"ssh_target_passphrase"`

	General  *generalBlockModel  `tfsdk:"general"`
	Emulator *emulatorBlockModel `tfsdk:"emulator"`
	ZosCreds *zosCredsBlockModel `tfsdk:"zos_creds"`
	Ipl      *iplBlockModel      `tfsdk:"ipl"`
	Timeouts timeouts.Value      `tfsdk:"timeouts"`
}

type generalBlockModel struct {
	Label               types.String `tfsdk:"label"`
	Description         types.String `tfsdk:"description"`
	SSHPublicKey        types.String `tfsdk:"ssh_public_key"`
	TargetUUID          types.String `tfsdk:"target_uuid"`
	ImageUUID           types.String `tfsdk:"image_uuid"`
	DeploymentDirectory types.String `tfsdk:"deployment_directory"`
	SysResComponentUUID types.String `tfsdk:"sysres_component_uuid"`
}

type emulatorBlockModel struct {
	CP   types.Int64 `tfsdk:"cp"`
	Ziip types.Int64 `tfsdk:"ziip"`
	Ram  types.Int64 `tfsdk:"ram"`
}

type zosCredsBlockModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type iplBlockModel struct {
	DeviceAddress types.String `tfsdk:"device_address"`
	IODFAddress   types.String `tfsdk:"iodf_address"`
	LoadSuffix    types.String `tfsdk:"load_suffix"`
}

func (r *provisionedInstanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*odeclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *ode.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData,
			),
		)
		return
	}
	r.client = client
}

func (r *provisionedInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan provisionedInstanceResourceModel
	var sshTargetUser, sshTargetPassword, sshKeyFilePath, sshPassphrase string
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	//Check that we have a client
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Client not configured", "Provider did not configure the On Demand Environment client.",
		)
		return
	}

	// get write-only attribute values from the config
	req.Config.GetAttribute(ctx, path.Root("ssh_target_user"), &sshTargetUser)
	req.Config.GetAttribute(ctx, path.Root("ssh_target_password"), &sshTargetPassword)
	req.Config.GetAttribute(ctx, path.Root("ssh_target_key_file"), &sshKeyFilePath)
	req.Config.GetAttribute(ctx, path.Root("ssh_target_passphrase"), &sshPassphrase)

	// set input.auth values
	auth := instance.SSHCredentials{}
	// set username
	if sshTargetUser == "" {
		sshTargetUser = os.Getenv("SSH_TARGET_USER")
	}
	auth.Username = sshTargetUser
	// set password
	if sshTargetPassword == "" {
		sshTargetPassword = os.Getenv("SSH_TARGET_PASSWORD")
	}
	auth.Password = sshTargetPassword
	// set key_file
	var k []byte
	if sshKeyFilePath != "" {
		k = []byte(sshKeyFilePath)
	} else {
		sshKeyFilePath_env := os.Getenv("SSH_KEY_FILE_PATH")
		k = []byte(sshKeyFilePath_env)
	}
	auth.KeyFile = &k
	// set sshPassphrase
	if sshPassphrase == "" {
		sshPassphrase = os.Getenv("SSH_TARGET_PASSPHRASE")
	}
	auth.Passphrase = &sshPassphrase
	// fill rest of request
	zosCreds := instance.ZosCreds{}
	if plan.ZosCreds != nil {
		zosCreds.Username = plan.ZosCreds.Username.ValueString()
		zosCreds.Password = plan.ZosCreds.Password.ValueString()
	}
	request := instance.LinuxProvisionRequest{
		General: instance.General{
			Label:               plan.General.Label.ValueString(),
			Description:         plan.General.Description.ValueString(),
			TargetUUID:          plan.General.TargetUUID.ValueString(),
			ImageUUID:           plan.General.ImageUUID.ValueString(),
			SysResComponentUUID: plan.General.SysResComponentUUID.ValueString(),
			SSHPublicKey:        plan.General.SSHPublicKey.ValueString(),
		},
		ZosCreds:            zosCreds,
		DeploymentDirectory: plan.General.DeploymentDirectory.ValueString(),
		ValidateLinux:       true,
		Emulator: instance.Emulator{
			CP:   plan.Emulator.CP.ValueInt64(),
			Ziip: plan.Emulator.Ziip.ValueInt64(),
			Ram:  plan.Emulator.Ram.ValueInt64(),
		},
	}

	// Fill ipl
	if plan.Ipl != nil {
		request.IPL = &instance.IPL{
			DeviceAddress: plan.Ipl.DeviceAddress.ValueString(),
			IODFAddress:   plan.Ipl.IODFAddress.ValueString(),
			LoadSuffix:    plan.Ipl.LoadSuffix.ValueString(),
		}
	}
	input := instance.CreateInput{
		Request: request,
		Auth:    auth,
	}
	// Create a context with a timeout
	createTimeout, diags := plan.Timeouts.Create(ctx, 120*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()
	// create instance
	uuid, err := r.client.Instance.Create(ctx, input)
	if err != nil {
		if strings.Contains(err.Error(), "monitor") {
			plan.Status = types.StringValue("in_progress")
		}
		resp.Diagnostics.AddError("Error Creating Provision", err.Error())
		return
	}

	targ, err := r.client.Target.Get(ctx, plan.General.TargetUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Getting Target", err.Error())
	}

	if targ != nil && targ.Hostname != "" {
		plan.Hostname = types.StringValue(targ.Hostname)
	}

	plan.ProvisionUUID = types.StringValue(uuid)
	plan.Status = types.StringValue("completed")

	// Final state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *provisionedInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state provisionedInstanceResourceModel
	var targetHost string
	diags := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get refreshed state
	provisionData, err := r.client.Instance.Get(ctx, state.ProvisionUUID.ValueString())

	if err != nil {
		if IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Checking Status", err.Error())
		return
	}

	updatedStatus := GetStatusString(provisionData)
	targ, err := r.client.Target.Get(ctx, provisionData.TargetUUID)
	if err != nil {
		resp.Diagnostics.AddError("Error Getting Target Status", err.Error())
	}
	if targ != nil && targ.Hostname != "" {
		targetHost = targ.Hostname
	}

	state = provisionedInstanceResourceModel{
		ProvisionUUID: types.StringValue(provisionData.ProvisionUUID),
		Status:        types.StringValue(updatedStatus),
		Hostname:      types.StringValue(targetHost),
		General: &generalBlockModel{
			Label:               types.StringValue(provisionData.Label),
			Description:         types.StringValue(provisionData.Description),
			ImageUUID:           types.StringValue(provisionData.ImageUUID),
			DeploymentDirectory: types.StringValue(provisionData.DeploymentDirectory),
			SysResComponentUUID: types.StringValue(provisionData.SysResComponentUUID),
			TargetUUID:          types.StringValue(provisionData.TargetUUID),
			SSHPublicKey:        types.StringValue(provisionData.SSHPublicKey),
		},
		Emulator: &emulatorBlockModel{
			CP:   types.Int64Value(provisionData.Emulator.CP),
			Ziip: types.Int64Value(provisionData.Emulator.Ziip),
			Ram:  types.Int64Value(provisionData.Emulator.Ram),
		},
	}
	if targ != nil && targ.Hostname != "" {
		state.Hostname = types.StringValue(targ.Hostname)
	}
	var timeouts timeouts.Value
	resp.State.GetAttribute(ctx, path.Root("timeouts"), &timeouts)
	state.Timeouts = timeouts

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *provisionedInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var timeoutsState, timeoutsConfig timeouts.Value
	var state provisionedInstanceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get timeouts value from state and config
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("timeouts"), &timeoutsState)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("timeouts"), &timeoutsConfig)...)

	if !timeoutsState.Equal(timeoutsConfig) {
		// update timeouts
		state.Timeouts = timeoutsConfig
		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

func (r *provisionedInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state provisionedInstanceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete a context with a timeout
	createTimeout, diags := state.Timeouts.Delete(ctx, 20*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()
	err := r.client.Instance.Delete(
		ctx, instance.DeleteInput{
			ProvisionUUID: state.ProvisionUUID.ValueString()},
	)

	if err != nil {
		if IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to Delete Resource", err.Error())
		return
	}
	var monitorTimeout timeouts.Value
	resp.State.GetAttribute(ctx, path.Root("timeouts"), &monitorTimeout)
	state.Timeouts = monitorTimeout

	// Final state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *provisionedInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("provision_uuid"), req, resp)
}

func GetStatusString(data instance.Data) string {
	if data.Successful {
		return "completed"
	} else if data.Failed {
		return "failed"
	} else if data.Cancelled {
		return "cancelled"
	} else if data.InProgress {
		return "in_progress"
	} else {
		return ""
	}
}

func (r *provisionedInstanceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data provisionedInstanceResourceModel
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

func IsNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "does not exist")
}
