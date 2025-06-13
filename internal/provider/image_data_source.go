// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	odeclient "github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/client"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/image"
)

var (
	_ datasource.DataSource              = &odeImageDataSource{}
	_ datasource.DataSourceWithConfigure = &odeImageDataSource{}
)

type odeImageDataSource struct {
	client *odeclient.Client
}

func NewOdeImageDataSource() datasource.DataSource {
	return &odeImageDataSource{}
}

// Metadata returns the data source type name.
func (d *odeImageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

// Configure implements the WithConfigure interface.
func (d *odeImageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*odeclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *odeclient.Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}
	d.client = client
}

// Schema defines the schema for the data source.
func (d *odeImageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"flatten": schema.BoolAttribute{
				Optional:    true,
				Description: "When true, expose a flat layer for the first image (image_list[0]).",
			},
			"uuid": schema.StringAttribute{
				Optional:    true,
				Description: "Exact image UUID lookup. Cannot be used with filter.",
			},
			"filter": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Filter by label and optional version. Omit to return all images.",
				Attributes: map[string]schema.Attribute{
					"label": schema.StringAttribute{
						Required:    true,
						Description: "Image label to filter on.",
					},
					"version": schema.Int64Attribute{
						Optional:    true,
						Description: "Image version to filter on (requires label).",
					},
				},
			},
			"image_list": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of images matching the query.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed:    true,
							Description: "UUID of the image.",
						},
						"label": schema.StringAttribute{
							Computed:    true,
							Description: "Label (name) of the image.",
						},
						"version": schema.Int64Attribute{
							Computed:    true,
							Description: "Version number of the image.",
						},
						"sysres_component_uuid": schema.StringAttribute{
							Computed:    true,
							Description: "SYSRES component UUID associated with the image.",
						},
						"ipl_parameter": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "IPL parameters for the image.",
							Attributes: map[string]schema.Attribute{
								"sysres_device": schema.StringAttribute{
									Computed:    true,
									Description: "SYSRES device number for IPL.",
								},
								"iodf_device": schema.StringAttribute{
									Computed:    true,
									Description: "IODF device number for IPL.",
								},
								"load_suffix": schema.StringAttribute{
									Computed:    true,
									Description: "Load suffix used in the IPL.",
								},
							},
						},
					},
				},
			},
			"label": schema.StringAttribute{
				Computed:    true,
				Description: "Shortcut to image_list[0].label when flatten=true.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "Shortcut to image_list[0].version when flatten=true.",
			},
			"sysres_component_uuid": schema.StringAttribute{
				Computed:    true,
				Description: "Shortcut to image_list[0].sysres_component_uuid when flatten=true.",
			},
			"ipl_parameter": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Shortcut to image_list[0].ipl_parameter when flatten=true.",
				Attributes: map[string]schema.Attribute{
					"sysres_device": schema.StringAttribute{
						Computed:    true,
						Description: "SYSRES device number for IPL.",
					},
					"iodf_device": schema.StringAttribute{
						Computed:    true,
						Description: "IODF device number for IPL.",
					},
					"load_suffix": schema.StringAttribute{
						Computed:    true,
						Description: "Load suffix used in the IPL.",
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *odeImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data odeImagesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := image.ListInput{
		Versions: true,
		UUID:     data.UUID.ValueString(),
	}

	if data.Filter != nil {
		if !data.Filter.Label.IsNull() {
			input.Label = data.Filter.Label.ValueString()
		}
		if !data.Filter.Version.IsNull() {
			v := int(data.Filter.Version.ValueInt64())
			input.Version = &v
		}
	}

	images, err := d.client.Image.List(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error Getting Images", err.Error())
		return
	}

	if data.Flatten.ValueBool() {
		if data.UUID.IsNull() && (data.Filter == nil || data.Filter.Label.IsNull() || data.Filter.Version.IsNull()) {
			resp.Diagnostics.AddError(
				"Invalid flatten usage",
				"`flatten = true` requires either a `uuid` or both `filter.label` and `filter.version` to uniquely identify one image.",
			)
			return
		}

		if len(images) != 1 {
			resp.Diagnostics.AddError(
				"Flatten requires exactly one image",
				fmt.Sprintf("Expected exactly one image, but found %d", len(images)),
			)
			return
		}

		img := images[0]
		data.UUID = types.StringValue(img.UUID)
		data.Label = types.StringValue(img.Name)
		data.Version = types.Int64Value(int64(img.Version))
		data.SysResComponentUUID = types.StringValue(img.SysResComponentUUID)

		if img.IPLParameter != nil {
			data.IPLParameter = &IplModel{
				SysResDevice: types.StringValue(img.IPLParameter.SysResDevice),
				IODFDevice:   types.StringValue(img.IPLParameter.IODFDevice),
				LoadSuffix:   types.StringValue(img.IPLParameter.LoadSuffix),
			}
		} else {
			data.IPLParameter = nil
		}

		data.ImageList = nil
	} else {
		data.ImageList = MapToState(images)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type filterModel struct {
	Label   types.String `tfsdk:"label"`
	Version types.Int64  `tfsdk:"version"`
}
type odeImagesDataSourceModel struct {
	Flatten             types.Bool   `tfsdk:"flatten"`
	UUID                types.String `tfsdk:"uuid"`
	Filter              *filterModel `tfsdk:"filter"`
	ImageList           []ImageModel `tfsdk:"image_list"`
	Label               types.String `tfsdk:"label"`
	Version             types.Int64  `tfsdk:"version"`
	SysResComponentUUID types.String `tfsdk:"sysres_component_uuid"`
	IPLParameter        *IplModel    `tfsdk:"ipl_parameter"`
}
type ImageModel struct {
	UUID                types.String `tfsdk:"uuid"`
	Label               types.String `tfsdk:"label"`
	Version             types.Int64  `tfsdk:"version"`
	SysResComponentUUID types.String `tfsdk:"sysres_component_uuid"`
	IPLParameter        *IplModel    `tfsdk:"ipl_parameter"`
}

type IplModel struct {
	SysResDevice types.String `tfsdk:"sysres_device"`
	IODFDevice   types.String `tfsdk:"iodf_device"`
	LoadSuffix   types.String `tfsdk:"load_suffix"`
}

func MapToState(images []image.StockImage) []ImageModel {
	models := make([]ImageModel, len(images))
	for i, img := range images {
		var ip *IplModel
		if p := img.IPLParameter; p != nil {
			ip = &IplModel{
				SysResDevice: types.StringValue(p.SysResDevice),
				IODFDevice:   types.StringValue(p.IODFDevice),
				LoadSuffix:   types.StringValue(p.LoadSuffix),
			}
		}
		models[i] = ImageModel{
			UUID:                types.StringValue(img.UUID),
			Label:               types.StringValue(img.Name),
			Version:             types.Int64Value(int64(img.Version)),
			SysResComponentUUID: types.StringValue(img.SysResComponentUUID),
			IPLParameter:        ip,
		}
	}
	return models
}
