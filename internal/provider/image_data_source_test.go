// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccStockImageData_All(t *testing.T) {
	resource.Test(
		t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 func() { testAccPreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testAccProviderConfig() + testAccStockImageAllConfig(),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("label"),
							knownvalue.StringExact("IBM Stock image (1.0.2)"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("uuid"),
							knownvalue.StringExact("e6daf8cd-894f-4a1e-afe0-d7056cd1b17b"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("version"),
							knownvalue.Int64Exact(1),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("sysres_component_uuid"),
							knownvalue.StringExact("499d5d92-fc99-41c8-94cf-28a92aad830f"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("ipl_parameter").AtMapKey("load_suffix"),
							knownvalue.StringExact("AU"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("ipl_parameter").AtMapKey("sysres_device"),
							knownvalue.StringExact("DE27"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("ipl_parameter").AtMapKey("iodf_device"),
							knownvalue.StringExact("DE28"),
						),

						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(2).AtMapKey("label"),
							knownvalue.StringExact("terraform-z-image"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(2).AtMapKey("uuid"),
							knownvalue.StringExact("8987520e-c3ad-4773-9b0d-b43b25dc0d87"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(2).AtMapKey("version"),
							knownvalue.Int64Exact(1),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(2).AtMapKey("sysres_component_uuid"),
							knownvalue.StringExact("cd2e13c4-c84c-4bc5-a208-77186ef4fab8"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(2).AtMapKey("ipl_parameter").AtMapKey("load_suffix"),
							knownvalue.StringExact("AU"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(2).AtMapKey("ipl_parameter").AtMapKey("sysres_device"),
							knownvalue.StringExact("DE27"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(2).AtMapKey("ipl_parameter").AtMapKey("iodf_device"),
							knownvalue.StringExact("DE28"),
						),
					},
				},
			},
		},
	)
}

func TestAccStockImageData_ByUUID(t *testing.T) {
	resource.Test(
		t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 func() { testAccPreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testAccProviderConfig() + testAccStockImageUUIDNoFlatConfig("e6daf8cd-894f-4a1e-afe0-d7056cd1b17b"),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("uuid"),
							knownvalue.StringExact("e6daf8cd-894f-4a1e-afe0-d7056cd1b17b"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("label"),
							knownvalue.StringExact("IBM Stock image (1.0.2)"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("version"),
							knownvalue.Int64Exact(1),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("sysres_component_uuid"),
							knownvalue.StringExact("499d5d92-fc99-41c8-94cf-28a92aad830f"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("ipl_parameter").AtMapKey("load_suffix"),
							knownvalue.StringExact("AU"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("ipl_parameter").AtMapKey("sysres_device"),
							knownvalue.StringExact("DE27"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("ipl_parameter").AtMapKey("iodf_device"),
							knownvalue.StringExact("DE28"),
						),
					},
				},
			},
		},
	)
}

func TestAccStockImageData_ByUUID_Flat(t *testing.T) {
	resource.Test(
		t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 func() { testAccPreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testAccProviderConfig() + testAccStockImageUUIDFlatConfig("e6daf8cd-894f-4a1e-afe0-d7056cd1b17b"),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("uuid"),
							knownvalue.StringExact("e6daf8cd-894f-4a1e-afe0-d7056cd1b17b"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("label"),
							knownvalue.StringExact("IBM Stock image (1.0.2)"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("version"),
							knownvalue.Int64Exact(1),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("sysres_component_uuid"),
							knownvalue.StringExact("499d5d92-fc99-41c8-94cf-28a92aad830f"),
						),

						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("ipl_parameter").AtMapKey("load_suffix"),
							knownvalue.StringExact("AU"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("ipl_parameter").AtMapKey("sysres_device"),
							knownvalue.StringExact("DE27"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("ipl_parameter").AtMapKey("iodf_device"),
							knownvalue.StringExact("DE28"),
						),
					},
				},
			},
		},
	)
}

func TestAccStockImageData_ByLabelVersion(t *testing.T) {
	resource.Test(
		t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 func() { testAccPreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testAccProviderConfig() + testAccStockImageFilterNoFlatConfig("IBM Stock image (1.0.2)", 1),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("label"),
							knownvalue.StringExact("IBM Stock image (1.0.2)"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("uuid"),
							knownvalue.StringExact("e6daf8cd-894f-4a1e-afe0-d7056cd1b17b"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("version"),
							knownvalue.Int64Exact(1),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test",
							tfjsonpath.New("image_list").AtSliceIndex(0).AtMapKey("sysres_component_uuid"),
							knownvalue.StringExact("499d5d92-fc99-41c8-94cf-28a92aad830f"),
						),

						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("image_list").AtSliceIndex(0).
								AtMapKey("ipl_parameter").AtMapKey("load_suffix"),
							knownvalue.StringExact("AU"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("image_list").AtSliceIndex(0).
								AtMapKey("ipl_parameter").AtMapKey("sysres_device"),
							knownvalue.StringExact("DE27"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("image_list").AtSliceIndex(0).
								AtMapKey("ipl_parameter").AtMapKey("iodf_device"),
							knownvalue.StringExact("DE28"),
						),
					},
				},
			},
		},
	)
}

func TestAccStockImageData_ByLabelVersion_Flat(t *testing.T) {
	resource.Test(
		t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 func() { testAccPreCheck(t) },
			Steps: []resource.TestStep{
				{
					Config: testAccProviderConfig() + testAccStockImageFilterFlatConfig("IBM Stock image (1.0.2)", 1),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("uuid"),
							knownvalue.StringExact("e6daf8cd-894f-4a1e-afe0-d7056cd1b17b"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("label"),
							knownvalue.StringExact("IBM Stock image (1.0.2)"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("version"),
							knownvalue.Int64Exact(1),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("sysres_component_uuid"),
							knownvalue.StringExact("499d5d92-fc99-41c8-94cf-28a92aad830f"),
						),

						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("ipl_parameter").AtMapKey("load_suffix"),
							knownvalue.StringExact("AU"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("ipl_parameter").AtMapKey("sysres_device"),
							knownvalue.StringExact("DE27"),
						),
						statecheck.ExpectKnownValue(
							"data.ode_image.test", tfjsonpath.New("ipl_parameter").AtMapKey("iodf_device"),
							knownvalue.StringExact("DE28"),
						),
					},
				},
			},
		},
	)
}

func testAccStockImageFilterNoFlatConfig(label string, version int) string {
	return fmt.Sprintf(
		`
data "ode_image" "test" {
  filter = {
    label   = "%s"
    version = %d
  }
}
`, label, version,
	)
}

func testAccStockImageFilterFlatConfig(label string, version int) string {
	return fmt.Sprintf(
		`
data "ode_image" "test" {
  flatten = true
  filter = {
    label   = "%s"
    version = %d
  }
}
`, label, version,
	)
}

func testAccStockImageUUIDFlatConfig(uuid string) string {
	return fmt.Sprintf(
		`
data "ode_image" "test" {
  flatten = true
  uuid    = "%s"
}
`, uuid,
	)
}

func testAccStockImageUUIDNoFlatConfig(uuid string) string {
	return fmt.Sprintf(
		`
data "ode_image" "test" {
  uuid = "%s"
}
`, uuid,
	)
}

func testAccStockImageAllConfig() string {
	return `
data "ode_image" "test" {}
`
}
