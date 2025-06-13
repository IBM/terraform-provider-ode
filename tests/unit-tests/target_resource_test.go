// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/target"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/provider"
)

func TestPlanToInput(t *testing.T) {
	fmt.Printf("Test case: %s ", t.Name())

	ipTablesSetting := provider.IptablesSettingModel{
		ZosIPAddress:    types.StringValue("192.168.1.2"),
		ZosSSHRoutePort: types.Int64Value(2222),
		TCPForwardPorts: []provider.ForwardPortModel{
			{StartPort: types.Int64Value(8080), EndPort: types.Int64Value(8089)},
			{StartPort: types.Int64Value(50000), EndPort: types.Int64Value(50050)},
		},
		UDPForwardPorts: []provider.ForwardPortModel{
			{StartPort: types.Int64Value(50051), EndPort: types.Int64Value(50100)},
		},
		TCPReroutePorts: []provider.ReroutePortModel{
			{LinuxPort: types.Int64Value(80), ZosPort: types.Int64Value(2001)},
			{LinuxPort: types.Int64Value(443), ZosPort: types.Int64Value(5050)},
		},
		UDPReroutePorts: []provider.ReroutePortModel{
			{LinuxPort: types.Int64Value(53), ZosPort: types.Int64Value(2053)},
		},
	}

	plan := provider.OdeTargetModel{
		ID:                   types.StringValue("planId"),
		Label:                types.StringValue("Test ODE Target"),
		Description:          types.StringValue("This is a test ODE target model."),
		Hostname:             types.StringValue("test-ode-target.example.com"),
		SSHPort:              types.Int64Value(22),
		ICPort:               types.Int64Value(8080),
		InstallDir:           types.StringValue("/opt/ode"),
		ConcurrentTransfers:  types.Int64Value(5),
		DNSIPPrimary:         types.StringValue("127.0.0.1"),
		DNSDomainOrigin:      types.StringValue("ibm.com"),
		TerminalPortStart:    types.Int64Value(50000),
		TerminalPortEnd:      types.Int64Value(50100),
		NetworkType:          types.StringValue("eth0"),
		PrivilegeCommandUUID: types.StringValue("abc123def456"),

		SSHTargetUser:       types.StringValue("odeuser"),
		SSHTargetPassword:   types.StringValue("password123"),
		SSHTargetKeyFile:    types.StringValue("/home/odeuser/.ssh/id_rsa"),
		SSHTargetPassphrase: types.StringValue("passphrase123"),

		Status:          types.StringValue("active"),
		IPTablesSetting: &ipTablesSetting,
	}

	expectedResult := target.CreateTargetInput{
		Request: target.IPTablesRequest{
			Label:               "Test ODE Target",
			Description:         "This is a test ODE target model.",
			Hostname:            "test-ode-target.example.com",
			SSHPort:             22,
			ICPort:              8080,
			DownloadDirectory:   "/opt/ode",
			DNSIPPrimary:        "127.0.0.1",
			DNSDomainOrigin:     "ibm.com",
			ConcurrentTransfers: 5,
			TerminalPortStart:   50000,
			NetworkType:         "eth0",
			IPTablesSetting: target.IPTablesSetting{
				ZosIPAddress:    "192.168.1.2",
				ZosSSHRoutePort: 2222,
				TCPForwardPorts: []target.ForwardPortRange{
					{StartPort: 8080, EndPort: 8089},
					{StartPort: 50000, EndPort: 50050},
				},
				UDPForwardPorts: []target.ForwardPortRange{
					{StartPort: 50051, EndPort: 50100},
				},
				TCPReroutePorts: []target.ReroutePortMapping{
					{LinuxPort: 80, ZosPort: 2001},
					{LinuxPort: 443, ZosPort: 5050},
				},
				UDPReroutePorts: []target.ReroutePortMapping{
					{LinuxPort: 53, ZosPort: 2053},
				},
			},
		},
		Auth: target.SSHCredentials{},
	}

	actualResult, actualErr := provider.PlanToInput(plan)

	if !reflect.DeepEqual(expectedResult, actualResult) {
		t.Errorf(`PlanToInput failed. Expected: %v, Actual: %v`, expectedResult, actualResult)
	} else if actualErr != nil {
		t.Errorf(`PlanToInput failed. Expected: %v, Actual: %s`, nil, actualErr.Error())
	} else {
		fmt.Println("passed")
	}
}

func TestPlanToInputMissingIpTable(t *testing.T) {
	fmt.Printf("Test case: %s ", t.Name())
	plan := provider.OdeTargetModel{}

	expectedResult := target.CreateTargetInput{}
	expectedError := "missing required iptable_setting block"

	actualResult, actualErr := provider.PlanToInput(plan)

	if !reflect.DeepEqual(expectedResult, actualResult) {
		t.Errorf(`PlanToInput failed. Expected: %v, Actual: %v`, expectedResult, actualResult)
	} else if expectedError != actualErr.Error() {
		t.Errorf(`PlanToInput failed. Expected: %s, Actual: %s`, expectedError, actualErr.Error())
	} else {
		fmt.Println("passed")
	}
}

func TestMapForwardPorts(t *testing.T) {
	fmt.Printf("Test case: %s ", t.Name())

	ports := []provider.ForwardPortModel{
		{StartPort: types.Int64Value(80), EndPort: types.Int64Value(8080)},
		{StartPort: types.Int64Value(50), EndPort: types.Int64Value(5050)},
	}

	expected := []target.ForwardPortRange{
		{StartPort: 80, EndPort: 8080},
		{StartPort: 50, EndPort: 5050},
	}

	actual := provider.MapForwardPorts(ports)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(`MapForwardPorts failed. Expected: %v, Actual: %v`, expected, actual)
	} else {
		fmt.Println("passed")
	}
}

func TestMapReroutePorts(t *testing.T) {
	fmt.Printf("Test case: %s ", t.Name())

	ports := []provider.ReroutePortModel{
		{LinuxPort: types.Int64Value(80), ZosPort: types.Int64Value(8080)},
		{LinuxPort: types.Int64Value(50), ZosPort: types.Int64Value(5050)},
	}

	expected := []target.ReroutePortMapping{
		{LinuxPort: 80, ZosPort: 8080},
		{LinuxPort: 50, ZosPort: 5050},
	}

	actual := provider.MapReroutePorts(ports)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(`MapForwardPorts failed. Expected: %v, Actual: %v`, expected, actual)
	} else {
		fmt.Println("passed")
	}
}
