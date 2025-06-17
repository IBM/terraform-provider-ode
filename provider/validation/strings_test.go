// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateExactlyOneOf(t *testing.T) {
	tests := []struct {
		name           string
		attributes     []StringAttribute // For setting attributes through literals
		envVars        map[string]string // For setting attributes env vars
		expectedErrors []string
	}{
		{
			name: "Only Key File Set - Literal. (i.e., ssh_target_key_passphrase = 'hello')",
			attributes: []StringAttribute{
				{Path: path.Root("ssh_target_key_file"), Value: types.StringValue("/path/to/key"), EnvVar: SSHKeyFile},
				{Path: path.Root(SSHPassword), Value: types.StringNull(), EnvVar: SSHPassword},
			},
			expectedErrors: nil,
		},
		{
			name: "Only Password Set - EnvVar. (i.e., export SSH_TARGET_PASSWORD='hello')",
			attributes: []StringAttribute{
				{Path: path.Root(SSHPassword), Value: types.StringValue("testPass"), EnvVar: SSHPassword},
				{Path: path.Root("ssh_target_key_file"), Value: types.StringNull(), EnvVar: SSHKeyFile},
			},
			envVars: map[string]string{
				SSHPassword: "secret",
			},
			expectedErrors: nil,
		},
		{
			name: "Three Attributes. Only One Is Set",
			attributes: []StringAttribute{
				{Path: path.Root(SSHPassword), Value: types.StringNull(), EnvVar: SSHPassword},
				{Path: path.Root("ssh_target_key_file"), Value: types.StringValue("/path/to/key"), EnvVar: SSHKeyFile},
				{Path: path.Root("ssh_target_token"), Value: types.StringNull(), EnvVar: "SSH_TARGET_TOKEN"},
			},
			expectedErrors: nil,
		},
		{
			name: "Multiple Set. Mixed Literal And Env",
			attributes: []StringAttribute{
				{Path: path.Root(SSHPassword), Value: types.StringValue("secret"), EnvVar: SSHPassword},
				{Path: path.Root("ssh_target_key_file"), Value: types.StringNull(), EnvVar: SSHKeyFile},
			},
			envVars: map[string]string{
				SSHKeyFile: "/path/to/key",
			},
			expectedErrors: []string{
				`Invalid Configuration`,
			},
		},
		{
			name: "None Set - Null",
			attributes: []StringAttribute{
				{Path: path.Root(SSHPassword), Value: types.StringNull(), EnvVar: SSHPassword},
				{Path: path.Root("ssh_target_key_file"), Value: types.StringNull(), EnvVar: SSHKeyFile},
			},
			expectedErrors: []string{
				`Missing Configuration`,
			},
		},
		{
			name: "Single Attribute",
			attributes: []StringAttribute{
				{Path: path.Root(SSHPassword), Value: types.StringValue("secret"), EnvVar: SSHPassword},
			},
			expectedErrors: []string{
				`Invalid Validation Call`,
			},
		},
		{
			name: "Unknown Value Skips Validation",
			attributes: []StringAttribute{
				{Path: path.Root(SSHPassword), Value: types.StringUnknown(), EnvVar: SSHPassword},
				{Path: path.Root("ssh_target_key_file"), Value: types.StringNull(), EnvVar: SSHKeyFile},
			},
			expectedErrors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// Set up environment variables
				for key, value := range tt.envVars {
					os.Setenv(key, value) //nolint:all
					defer os.Unsetenv(key)
				}

				// Run validation
				var diagnostics diag.Diagnostics
				ValidateExactlyOneOf(&diagnostics, tt.attributes...)

				// Check diag
				if len(tt.expectedErrors) == 0 {
					if diagnostics.HasError() {
						t.Fatalf("want: no errors, got: %v", diagnostics.Errors())
					}
				} else {
					if !diagnostics.HasError() {
						t.Fatal("want: errors, got: none")
					}
					if len(diagnostics.Errors()) != len(tt.expectedErrors) {
						t.Fatalf(
							"want: %d, got %d: %v", len(tt.expectedErrors), len(diagnostics.Errors()),
							diagnostics.Errors(),
						)
					}
					for i, expected := range tt.expectedErrors {
						err := diagnostics.Errors()[i]
						if !regexp.MustCompile(expected).MatchString(err.Summary()) {
							t.Errorf("want: %q, got: %q", expected, err.Summary())
						}
					}
				}
			},
		)
	}
}

func TestValidateAllSet(t *testing.T) {
	tests := []struct {
		name           string
		attributes     []StringAttribute
		envVars        map[string]string
		expectedErrors []string
	}{
		{
			name: "Only One Attribute Set Out of Three - String Literal Case - Error",
			attributes: []StringAttribute{
				{Path: path.Root("ssh_target_user"), Value: types.StringValue("/path/to/key"), EnvVar: SSHUser},
				{Path: path.Root("ssh_target_password"), Value: types.StringNull(), EnvVar: SSHPassword},
				{Path: path.Root("ssh_target_key"), Value: types.StringNull(), EnvVar: SSHKeyFile},
			},
			expectedErrors: []string{
				`Missing Configuration`,
			},
		},
		{
			name: "All Attribute Set Out of Three - String Literal Case - Pass",
			attributes: []StringAttribute{
				{Path: path.Root("ssh_target_user"), Value: types.StringValue("/path/to/key"), EnvVar: SSHUser},
				{Path: path.Root("ssh_target_password"), Value: types.StringValue("testPass"), EnvVar: SSHPassword},
				{Path: path.Root("ssh_target_key_file"), Value: types.StringValue("test"), EnvVar: SSHKeyFile},
			},
			expectedErrors: nil,
		},
		{
			name: "All Attribute Set Out of Three - Env Var Case - Pass",
			attributes: []StringAttribute{
				{Path: path.Root("ssh_target_user"), Value: types.StringNull(), EnvVar: SSHUser},
				{Path: path.Root("ssh_target_password"), Value: types.StringNull(), EnvVar: SSHPassword},
				{Path: path.Root("ssh_target_key_file"), Value: types.StringNull(), EnvVar: SSHKeyFile},
			},
			envVars: map[string]string{
				SSHUser:     "Izzy",
				SSHKeyFile:  "/path/to/key",
				SSHPassword: "secret",
			},
			expectedErrors: nil,
		},
		{
			name: "No Attribute Set Out of Three - ENV Var Case - Error",
			attributes: []StringAttribute{
				{Path: path.Root("ssh_target_user"), Value: types.StringNull(), EnvVar: SSHUser},
				{Path: path.Root("ssh_target_password"), Value: types.StringNull(), EnvVar: SSHPassword},
				{Path: path.Root("ssh_target_key_file"), Value: types.StringNull(), EnvVar: SSHKeyFile},
			},
			envVars: map[string]string{
				SSHUser:     "Izzy",
				SSHPassword: "secret",
				SSHKeyFile:  "/path/to/key",
			},
			expectedErrors: nil,
		},
		{
			name: "None Set Out of Three - String Literal Case - Error",
			attributes: []StringAttribute{
				{Path: path.Root("ssh_target_password"), Value: types.StringNull(), EnvVar: SSHPassword},
				{Path: path.Root("ssh_target_key_file"), Value: types.StringNull(), EnvVar: SSHKeyFile},
			},
			envVars: nil,
			expectedErrors: []string{
				`Missing Configuration`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				for key, value := range tt.envVars {
					t.Setenv(key, value)
				}

				var diagnostics diag.Diagnostics
				ValidateAllSet(&diagnostics, tt.attributes...)

				if len(tt.expectedErrors) == 0 {
					if diagnostics.HasError() {
						t.Fatalf("want: no errors, got: %v", diagnostics.Errors())
					}
				} else {
					if !diagnostics.HasError() {
						t.Fatal("want: errors, got: none")
					}
					if len(diagnostics.Errors()) != len(tt.expectedErrors) {
						t.Fatalf(
							"want: %d, got %d: %v", len(tt.expectedErrors), len(diagnostics.Errors()),
							diagnostics.Errors(),
						)
					}
					for i, expected := range tt.expectedErrors {
						err := diagnostics.Errors()[i]
						if !regexp.MustCompile(expected).MatchString(err.Summary()) {
							t.Errorf("want: %q, got: %q", expected, err.Summary())
						}
					}
				}
			},
		)
	}
}
