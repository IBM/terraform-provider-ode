// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/image"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/provider"
)

func TestMapToState(t *testing.T) {
	fmt.Printf("Test case:  %s ", t.Name())

	stockipl := image.IPL{
		SysResDevice: "AA-AA",
		IODFDevice:   "BB-BB",
		LoadSuffix:   "CC-CC",
	}

	modelipl := provider.IplModel{
		SysResDevice: types.StringValue("AA-AA"),
		IODFDevice:   types.StringValue("BB-BB"),
		LoadSuffix:   types.StringValue("CC-CC"),
	}

	images := []image.StockImage{{
		UUID:                "testUUID0",
		Name:                "testName0",
		Description:         "testDescription0",
		Version:             1,
		Type:                "testType0",
		Size:                999,
		IPLParameter:        &stockipl,
		SysResComponentUUID: "testSysresUUID0",
		ProvisionUUIDs:      []string{"testProvision0", "testProvision1"},
	}, {
		UUID:                "testUUID1",
		Name:                "testName1",
		Description:         "testDescription1",
		Version:             2,
		Type:                "testType0",
		Size:                555,
		IPLParameter:        &stockipl,
		SysResComponentUUID: "testSysresUUID1",
		ProvisionUUIDs:      []string{"testProvision1", "testProvision2"},
	}}

	expectedModels := []provider.ImageModel{{
		UUID:                types.StringValue("testUUID0"),
		Label:               types.StringValue("testName0"),
		Version:             types.Int64Value(1),
		SysResComponentUUID: types.StringValue("testSysresUUID0"),
		IPLParameter:        &modelipl,
	}, {
		UUID:                types.StringValue("testUUID1"),
		Label:               types.StringValue("testName1"),
		Version:             types.Int64Value(2),
		SysResComponentUUID: types.StringValue("testSysresUUID1"),
		IPLParameter:        &modelipl,
	}}

	actualModels := provider.MapToState(images)

	if !reflect.DeepEqual(expectedModels, actualModels) {
		t.Errorf(`MapToState failed. Expected: %v, Actual: %v `, expectedModels, actualModels)
	} else {
		fmt.Println("passed")
	}
}
