// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"ode": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccProviderConfig() string {
	return fmt.Sprintf(
		`
provider "ode" {
  ode_host     = "%s"
  ode_username = "%s"
  ode_password = "%s"
}
`, os.Getenv("ODE_HOST"), os.Getenv("ODE_USERNAME"), os.Getenv("ODE_PASSWORD"),
	)
}
