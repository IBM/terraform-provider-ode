// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/instance"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/provider"
)

func TestGetStatusStringSuccessful(t *testing.T) {
	fmt.Printf("Test case: %s ", t.Name())
	data := instance.Data{
		Successful: true,
		Failed:     false,
		Cancelled:  false,
		InProgress: false,
	}

	expected := "completed"

	actual := provider.GetStatusString(data)

	if actual != "completed" {
		t.Errorf(`GetStatusString failed. Expected: %s, Actual: %s`, expected, actual)
	} else {
		fmt.Println("passed")
	}

}

func TestGetStatusStringFailed(t *testing.T) {
	fmt.Printf("Test case:  %s ", t.Name())
	data := instance.Data{
		Successful: false,
		Failed:     true,
		Cancelled:  false,
		InProgress: false,
	}

	expected := "failed"

	actual := provider.GetStatusString(data)

	if actual != "failed" {
		t.Errorf(`GetStatusString failed. Expected: %s, Actual: %s`, expected, actual)
	} else {
		fmt.Println("passed")
	}
}

func TestGetStatusStringCancelled(t *testing.T) {
	fmt.Printf("Test case:  %s ", t.Name())
	data := instance.Data{
		Successful: false,
		Failed:     false,
		Cancelled:  true,
		InProgress: false,
	}

	expected := "cancelled"

	result := provider.GetStatusString(data)

	if result != "cancelled" {
		t.Errorf(`GetStatusString failed. Expected: %s, Actual: %s`, expected, result)
	} else {
		fmt.Println("passed")
	}

}

func TestGetStatusStringInProgress(t *testing.T) {
	fmt.Printf("Test case:  %s ", t.Name())
	data := instance.Data{
		Successful: false,
		Failed:     false,
		Cancelled:  false,
		InProgress: true,
	}

	expected := "in_progress"

	result := provider.GetStatusString(data)

	if result != "in_progress" {
		t.Errorf(`GetStatusString failed. Expected: %s, Actual: %s`, expected, result)
	} else {
		fmt.Println("passed")
	}
}

func TestGetStatusStringEmpty(t *testing.T) {
	fmt.Printf("Test case:  %s ", t.Name())
	data := instance.Data{
		Successful: false,
		Failed:     false,
		Cancelled:  false,
		InProgress: false,
	}

	expected := ""

	actual := provider.GetStatusString(data)

	if actual != "" {
		t.Errorf(`GetStatusString failed. Expected: %s, Actual: %s`, expected, actual)
	} else {
		fmt.Println("passed")
	}
}

func TestIsNotFoundErrorWithCode(t *testing.T) {
	fmt.Printf("Test case:  %s ", t.Name())

	err := errors.New("404 - some message")

	expected := provider.IsNotFoundError(err)

	actual := true

	if expected != actual {
		t.Errorf(`IsNotFoundError failed. Expected: %t, Actual: %t`, expected, actual)
	} else {
		fmt.Println("passed")
	}
}

func TestIsNotFoundErrorWithMessage(t *testing.T) {
	fmt.Printf("Test case:  %s ", t.Name())

	err := errors.New("does not exist")

	expected := provider.IsNotFoundError(err)

	actual := true

	if expected != actual {
		t.Errorf(`IsNotFoundError failed. Expected: %t, Actual: %t`, expected, actual)
	} else {
		fmt.Println("passed")
	}
}

func TestIsNotFoundErrorWrongCode(t *testing.T) {
	fmt.Printf("Test case:  %s ", t.Name())

	err := errors.New("403 - Forbidden")

	expected := provider.IsNotFoundError(err)

	actual := false

	if expected != actual {
		t.Errorf(`IsNotFoundError failed. Expected: %t, Actual: %t`, expected, actual)
	} else {
		fmt.Println("passed")
	}
}
