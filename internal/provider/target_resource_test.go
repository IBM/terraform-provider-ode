// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTargetAttributes_basic(t *testing.T) {
	params := targetConfigParams{
		Label:               os.Getenv("TARGET_LABEL"),
		SSHUser:             os.Getenv("TARGET_SSH_USER"),
		SSHPassword:         os.Getenv("TARGET_SSH_PASSWORD"),
		Hostname:            os.Getenv("TARGET_HOSTNAME"),
		DNSIPPrimary:        os.Getenv("TARGET_DNS_IP_PRIMARY"),
		DNSDomainOrigin:     os.Getenv("TARGET_DNS_DOMAIN_ORIGIN"),
		InstallDir:          "/opt",
		ICPort:              8443,
		SSHPort:             22,
		ConcurrentTransfers: 4,
		TerminalPortStart:   3270,
		ZosIPAddress:        os.Getenv("TARGET_ZOS_IP"),
		ZosSSHRoutePort:     2022,
		TCPForwardPorts: []forwardPorts{
			{StartPort: 0, EndPort: 21},
			{StartPort: 23, EndPort: 2021},
			{StartPort: 2023, EndPort: 3269},
			{StartPort: 3271, EndPort: 8442},
			{StartPort: 8444, EndPort: 9449},
			{StartPort: 9452, EndPort: 65535},
		},
		UDPForwardPorts: []forwardPorts{
			{StartPort: 111, EndPort: 111},
			{StartPort: 514, EndPort: 514},
			{StartPort: 1023, EndPort: 1023},
			{StartPort: 1044, EndPort: 1049},
			{StartPort: 2049, EndPort: 2049},
		},
		TCPReroutePorts: []reroutePorts{
			{LinuxPort: 2022, ZosPort: 22},
		},
		UDPReroutePorts: []reroutePorts{
			{LinuxPort: 2022, ZosPort: 22},
		},
	}

	config := testAccProviderConfig() + testAccTargetResourceConfig(params)

	resource.Test(
		t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectNonEmptyPlan(),
							plancheck.ExpectResourceAction("ode_target.test", plancheck.ResourceActionCreate),

							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("label"), knownvalue.StringExact(params.Label),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("hostname"),
								knownvalue.StringExact(params.Hostname),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("dns_domain_origin"),
								knownvalue.StringExact(params.DNSDomainOrigin),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("dns_ip_primary"),
								knownvalue.StringExact(params.DNSIPPrimary),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("install_dir"),
								knownvalue.StringExact(params.InstallDir),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("tcp_forward_ports").AtSliceIndex(0).AtMapKey("start_port"),
								knownvalue.Int64Exact(int64(params.TCPForwardPorts[0].StartPort)),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("tcp_forward_ports").AtSliceIndex(0).AtMapKey("end_port"),
								knownvalue.Int64Exact(int64(params.TCPForwardPorts[0].EndPort)),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("udp_forward_ports").AtSliceIndex(0).AtMapKey("start_port"),
								knownvalue.Int64Exact(int64(params.UDPForwardPorts[0].StartPort)),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("udp_forward_ports").AtSliceIndex(0).AtMapKey("end_port"),
								knownvalue.Int64Exact(int64(params.UDPForwardPorts[0].EndPort)),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("tcp_reroute_ports").AtSliceIndex(0).AtMapKey("linux_port"),
								knownvalue.Int64Exact(int64(params.TCPReroutePorts[0].LinuxPort)),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("tcp_reroute_ports").AtSliceIndex(0).AtMapKey("zos_port"),
								knownvalue.Int64Exact(int64(params.TCPReroutePorts[0].ZosPort)),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("udp_reroute_ports").AtSliceIndex(0).AtMapKey("linux_port"),
								knownvalue.Int64Exact(int64(params.UDPReroutePorts[0].LinuxPort)),
							),
							plancheck.ExpectKnownValue(
								"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("udp_reroute_ports").AtSliceIndex(0).AtMapKey("zos_port"),
								knownvalue.Int64Exact(int64(params.UDPReroutePorts[0].ZosPort)),
							),

							plancheck.ExpectUnknownValue("ode_target.test", tfjsonpath.New("status")),
							plancheck.ExpectUnknownValue("ode_target.test", tfjsonpath.New("id")),
						},
					},
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("label"), knownvalue.StringExact(params.Label),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("hostname"), knownvalue.StringExact(params.Hostname),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("dns_domain_origin"),
							knownvalue.StringExact(params.DNSDomainOrigin),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("dns_ip_primary"),
							knownvalue.StringExact(params.DNSIPPrimary),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("install_dir"),
							knownvalue.StringExact(params.InstallDir),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("zos_ip_address"),
							knownvalue.StringExact(params.ZosIPAddress),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("zos_ssh_route_port"),
							knownvalue.Int64Exact(int64(params.ZosSSHRoutePort)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("tcp_forward_ports").AtSliceIndex(0).AtMapKey("start_port"),
							knownvalue.Int64Exact(int64(params.TCPForwardPorts[0].StartPort)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("tcp_forward_ports").AtSliceIndex(0).AtMapKey("end_port"),
							knownvalue.Int64Exact(int64(params.TCPForwardPorts[0].EndPort)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("udp_forward_ports").AtSliceIndex(0).AtMapKey("start_port"),
							knownvalue.Int64Exact(int64(params.UDPForwardPorts[0].StartPort)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("udp_forward_ports").AtSliceIndex(0).AtMapKey("end_port"),
							knownvalue.Int64Exact(int64(params.UDPForwardPorts[0].EndPort)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("tcp_reroute_ports").AtSliceIndex(0).AtMapKey("linux_port"),
							knownvalue.Int64Exact(int64(params.TCPReroutePorts[0].LinuxPort)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("tcp_reroute_ports").AtSliceIndex(0).AtMapKey("zos_port"),
							knownvalue.Int64Exact(int64(params.TCPReroutePorts[0].ZosPort)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("udp_reroute_ports").AtSliceIndex(0).AtMapKey("linux_port"),
							knownvalue.Int64Exact(int64(params.UDPReroutePorts[0].LinuxPort)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("iptable_setting").AtMapKey("udp_reroute_ports").AtSliceIndex(0).AtMapKey("zos_port"),
							knownvalue.Int64Exact(int64(params.UDPReroutePorts[0].ZosPort)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("terminal_port_start"),
							knownvalue.Int64Exact(int64(params.TerminalPortStart)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("concurrent_transfers"),
							knownvalue.Int64Exact(int64(params.ConcurrentTransfers)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("ssh_port"),
							knownvalue.Int64Exact(int64(params.SSHPort)),
						),
						statecheck.ExpectKnownValue(
							"ode_target.test", tfjsonpath.New("ic_port"),
							knownvalue.Int64Exact(int64(params.ICPort)),
						),
					},
				},
			},
		},
	)
}

func testAccPreCheck(t *testing.T) {
	required := []string{
		"ODE_USERNAME",
		"ODE_PASSWORD",
		"ODE_HOST",
		"SSH_TARG_ROOT_PASS",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			t.Fatalf("%s must be set for acceptance tests", env)
		}
	}
}

type targetConfigParams struct {
	Label               string
	Hostname            string
	Description         string
	SSHUser             string
	SSHPassword         string
	DNSIPPrimary        string
	DNSDomainOrigin     string
	InstallDir          string
	ICPort              int
	SSHPort             int
	ConcurrentTransfers int
	TerminalPortStart   int

	ZosIPAddress    string
	ZosSSHRoutePort int
	TCPForwardPorts []forwardPorts
	UDPForwardPorts []forwardPorts
	TCPReroutePorts []reroutePorts
	UDPReroutePorts []reroutePorts
}

type forwardPorts struct {
	StartPort int
	EndPort   int
}
type reroutePorts struct {
	LinuxPort int
	ZosPort   int
}

func testAccTargetResourceConfig(p targetConfigParams) string {
	return fmt.Sprintf(
		`
resource "ode_target" "test" {
  label                 = "%s"
  hostname              = "%s"
  description			= "%s"
  ssh_target_user       = "%s"
  ssh_target_password   = "%s"
  dns_ip_primary        = "%s"
  dns_domain_origin     = "%s"
  install_dir           = "%s"
  ic_port               = %d
  ssh_port              = %d
  concurrent_transfers  = %d
  terminal_port_start   = %d

  iptable_setting = {
    zos_ip_address       = "%s"
    zos_ssh_route_port   = %d

    tcp_forward_ports = [
      { start_port = %d, end_port = %d},
      { start_port = %d, end_port = %d},
      { start_port = %d, end_port = %d},
      { start_port = %d, end_port = %d},
      { start_port = %d, end_port = %d},
      { start_port = %d, end_port = %d}
    ]

    udp_forward_ports = [
      { start_port = %d, end_port = %d },
	  { start_port = %d, end_port = %d },
	  { start_port = %d, end_port = %d },
	  { start_port = %d, end_port = %d },
	  { start_port = %d, end_port = %d }
    ]

    tcp_reroute_ports = [
      { linux_port = %d, zos_port = %d }
    ]

    udp_reroute_ports = [
      { linux_port = %d, zos_port = %d }
    ]
  }
}
`,
		p.Label,
		p.Hostname,
		p.Description,
		p.SSHUser,
		p.SSHPassword,
		p.DNSIPPrimary,
		p.DNSDomainOrigin,
		p.InstallDir,
		p.ICPort,
		p.SSHPort,
		p.ConcurrentTransfers,
		p.TerminalPortStart,
		p.ZosIPAddress,
		p.ZosSSHRoutePort,

		p.TCPForwardPorts[0].StartPort, p.TCPForwardPorts[0].EndPort,
		p.TCPForwardPorts[1].StartPort, p.TCPForwardPorts[1].EndPort,
		p.TCPForwardPorts[2].StartPort, p.TCPForwardPorts[2].EndPort,
		p.TCPForwardPorts[3].StartPort, p.TCPForwardPorts[3].EndPort,
		p.TCPForwardPorts[4].StartPort, p.TCPForwardPorts[4].EndPort,
		p.TCPForwardPorts[5].StartPort, p.TCPForwardPorts[5].EndPort,

		p.UDPForwardPorts[0].StartPort, p.UDPForwardPorts[0].EndPort,
		p.UDPForwardPorts[1].StartPort, p.UDPForwardPorts[1].EndPort,
		p.UDPForwardPorts[2].StartPort, p.UDPForwardPorts[2].EndPort,
		p.UDPForwardPorts[3].StartPort, p.UDPForwardPorts[3].EndPort,
		p.UDPForwardPorts[4].StartPort, p.UDPForwardPorts[4].EndPort,

		p.TCPReroutePorts[0].LinuxPort, p.TCPReroutePorts[0].ZosPort,

		p.UDPReroutePorts[0].LinuxPort, p.UDPReroutePorts[0].ZosPort,
	)
}
