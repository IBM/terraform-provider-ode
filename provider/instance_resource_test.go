// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccProvisionAttributes_basic(t *testing.T) {
	if os.Getenv("SSH_PASSWORD") == "" {
		t.Fatal("SSH_PASSWORD environment variable must be set for test")
	}
	params := instanceConfigParams{
		Label:               os.Getenv("INSTANCE_LABEL"),
		Description:         os.Getenv("INSTANCE_DESCRIPTION"),
		TargetUUID:          os.Getenv("INSTANCE_TARGET_UUID"),
		ImageUUID:           os.Getenv("INSTANCE_IMAGE_UUID"),
		SSHUser:             os.Getenv("INSTANCE_SSH_USER"),
		SSHPassword:         os.Getenv("INSTANCE_SSH_PASSWORD"),
		SSHPublicKey:        os.Getenv("INSTANCE_SSH_PUBLIC_KEY"),
		SysresComponentUUID: os.Getenv("INSTANCE_SYSRES_COMPONENT_UUID"),
		DeploymentDirectory: os.Getenv("INSTANCE_DEPLOYMENT_DIRECTORY"),
		CP:                  3,
		RAM:                 5368709120,
		ZIIP:                0,
	}
	config := testAccProviderConfig() + testAccInstanceResourceConfig(params)
	resource.Test(
		t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectNonEmptyPlan(),
							plancheck.ExpectResourceAction("ode_instance.test", plancheck.ResourceActionCreate),

							plancheck.ExpectKnownValue(
								"ode_instance.test", tfjsonpath.New("general").AtMapKey("label"),
								knownvalue.StringExact(params.Label),
							),
							plancheck.ExpectKnownValue(
								"ode_instance.test", tfjsonpath.New("general").AtMapKey("description"),
								knownvalue.StringExact(params.Description),
							),
							plancheck.ExpectKnownValue(
								"ode_instance.test", tfjsonpath.New("general").AtMapKey("target_uuid"),
								knownvalue.StringExact(params.TargetUUID),
							),
							plancheck.ExpectKnownValue(
								"ode_instance.test", tfjsonpath.New("general").AtMapKey("image_uuid"),
								knownvalue.StringExact(params.ImageUUID),
							),
							plancheck.ExpectKnownValue(
								"ode_instance.test", tfjsonpath.New("general").AtMapKey("ssh_public_key"),
								knownvalue.StringExact(params.SSHPublicKey),
							),
							plancheck.ExpectKnownValue(
								"ode_instance.test",
								tfjsonpath.New("general").AtMapKey("sysres_component_uuid"),
								knownvalue.StringExact(params.SysresComponentUUID),
							),
							plancheck.ExpectKnownValue(
								"ode_instance.test",
								tfjsonpath.New("general").AtMapKey("deployment_directory"),
								knownvalue.StringExact(params.DeploymentDirectory),
							),
							plancheck.ExpectKnownValue(
								"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("cp"),
								knownvalue.Int64Exact(int64(params.CP)),
							),
							plancheck.ExpectKnownValue(
								"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ram"),
								knownvalue.Int64Exact(params.RAM),
							),
							plancheck.ExpectKnownValue(
								"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ziip"),
								knownvalue.Int64Exact(int64(params.ZIIP)),
							),

							plancheck.ExpectUnknownValue(
								"ode_instance.test", tfjsonpath.New("provision_uuid"),
							),
							plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("hostname")),
							plancheck.ExpectUnknownValue("ode_instance.test", tfjsonpath.New("status")),
						},
					},
					ConfigStateChecks: []statecheck.StateCheck{
						assertInt64ExactPath(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("cp"), int64(params.CP),
						),
						assertInt64ExactPath(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ram"), params.RAM,
						),
						assertInt64ExactPath(
							"ode_instance.test", tfjsonpath.New("emulator").AtMapKey("ziip"),
							int64(params.ZIIP),
						),
						assertStringExactPath(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("label"), params.Label,
						),
						assertStringExactPath(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("description"),
							params.Description,
						),

						assertNotNullPath(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("sysres_component_uuid"),
						),
						assertNotNullPath(
							"ode_instance.test", tfjsonpath.New("general").AtMapKey("image_uuid"),
						),
						assertNotNullPath("ode_instance.test", tfjsonpath.New("status")),
						assertNotNullPath("ode_instance.test", tfjsonpath.New("provision_uuid")),
						assertNotNullPath("ode_instance.test", tfjsonpath.New("hostname")),
					},
				},
			},
		},
	)
}

type instanceConfigParams struct {
	Label               string
	Description         string
	TargetUUID          string
	ImageUUID           string
	SSHUser             string
	SSHPassword         string
	SSHPublicKey        string
	SysresComponentUUID string
	DeploymentDirectory string
	CP                  int
	RAM                 int64
	ZIIP                int
}

func testAccInstanceResourceConfig(p instanceConfigParams) string {
	return fmt.Sprintf(
		`
resource "ode_instance" "test" {
  ssh_target_user     = "%s"
  ssh_target_password = "%s"

  emulator = {
    cp   = %d
    ram  = %d
    ziip = %d
  }

  general = {
    label                 = "%s"
    description           = "%s"
    target_uuid           = "%s"
    image_uuid            = "%s"
    ssh_public_key        = "%s"
    sysres_component_uuid = "%s"
    deployment_directory  = "%s"
  }
}
`,
		p.SSHUser,
		p.SSHPassword,
		p.CP,
		p.RAM,
		p.ZIIP,
		p.Label,
		p.Description,
		p.TargetUUID,
		p.ImageUUID,
		p.SSHPublicKey,
		p.SysresComponentUUID,
		p.DeploymentDirectory,
	)
}

//nolint:all
func assertNotNullPath(resourceName string, path tfjsonpath.Path) statecheck.StateCheck {
	return statecheck.ExpectKnownValue(resourceName, path, knownvalue.NotNull())
}

func assertStringExactPath(resourceName string, path tfjsonpath.Path, expected string) statecheck.StateCheck {
	return statecheck.ExpectKnownValue(resourceName, path, knownvalue.StringExact(expected))
}

func assertInt64ExactPath(resourceName string, path tfjsonpath.Path, expected int64) statecheck.StateCheck {
	return statecheck.ExpectKnownValue(resourceName, path, knownvalue.Int64Exact(expected))
}
