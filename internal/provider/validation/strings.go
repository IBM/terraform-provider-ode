// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	SSHPassword = "SSH_TARGET_PASSWORD"
	SSHKeyFile  = "SSH_KEY_FILE_PATH"
	SSHUser     = "SSH_TARGET_USER"
)

// StringAttribute defines fields for a string attribute with an environment variable fallback.
type StringAttribute struct {
	Path   path.Path
	Value  types.String
	EnvVar string
}

// validateSetup performs common validation checks for config phase validations.
func validateSetup(diagnostics *diag.Diagnostics, attributes []StringAttribute, minAttrs int, errMsg string) bool {
	if len(attributes) < minAttrs {
		diagnostics.AddError(
			"Invalid Validation Call",
			fmt.Sprintf("%s Provide at least %d attribute(s) to validate.", errMsg, minAttrs),
		)
		return false
	}

	// Skip validation if any attribute is a variable. TF variable = Unknown.
	for _, attr := range attributes {
		if attr.Value.IsUnknown() {
			return false
		}
	}
	return true
}

// ValidateExactlyOneOf ensures exactly one of the strings is set through literal or env.
func ValidateExactlyOneOf(diagnostics *diag.Diagnostics, attributes ...StringAttribute) {
	if !validateSetup(diagnostics, attributes, 2, "At least two attributes must be provided.") {
		return
	}

	setCount := 0
	var setIndices []int
	for i, attr := range attributes {
		if isStringAttributeSet(attr) {
			setCount++
			setIndices = append(setIndices, i)
		}
	}

	attrNames := getAttributeNames(attributes)
	attrList := strings.Join(attrNames, ", ")
	envVars := getEnvVars(attributes)

	if setCount > 1 {
		setAttrNames := make([]string, len(setIndices))
		for i, id := range setIndices {
			setAttrNames[i] = attributes[id].Path.String()
		}
		diagnostics.AddError(
			"Invalid Configuration",
			fmt.Sprintf(
				"Exactly one of %s must be set, but %d were set (%s). "+
					"Remove all but one from the configuration or environment variables (%s).",
				attrList, setCount, strings.Join(setAttrNames, ", "), strings.Join(envVars, ", "),
			),
		)
	} else if setCount == 0 {
		diagnostics.AddError(
			"Missing Configuration",
			fmt.Sprintf(
				"Exactly one of %s is required. "+
					"Set one in the configuration (non-empty literal) or through environment variable (%s).",
				attrList, strings.Join(envVars, ", "),
			),
		)
	}
}

// ValidateAllSet ensures all provided strings are set through literal or env.
func ValidateAllSet(diagnostics *diag.Diagnostics, attributes ...StringAttribute) {
	if !validateSetup(diagnostics, attributes, 1, "At least one attribute must be provided.") {
		return
	}

	var unsetAttrs []string
	for _, attr := range attributes {
		if !isStringAttributeSet(attr) {
			unsetAttrs = append(unsetAttrs, attr.Path.String())
		}
	}

	if len(unsetAttrs) > 0 {
		attrNames := getAttributeNames(attributes)
		attrList := strings.Join(attrNames, ", ")
		envVars := getEnvVars(attributes)
		diagnostics.AddError(
			"Missing Configuration",
			fmt.Sprintf(
				"All of the following attributes: %s must be set, but %s were not set. "+
					"Ensure all are provided in the configuration (non empty literal) or through environment variables (%s).",
				attrList, strings.Join(unsetAttrs, ", "), strings.Join(envVars, ", "),
			),
		)
	}
}

// isStringAttributeSet checks if a string or its environment variable is set.
// Returns true if the value is present.
func isStringAttributeSet(attr StringAttribute) bool {
	if !attr.Value.IsUnknown() && !attr.Value.IsNull() && attr.Value.ValueString() != "" {
		return true
	}
	return os.Getenv(attr.EnvVar) != ""
}

// getAttributeNames returns a list of names for the given attributes.
func getAttributeNames(attributes []StringAttribute) []string {
	names := make([]string, len(attributes))
	for i, attr := range attributes {
		names[i] = attr.Path.String()
	}
	return names
}

// getEnvVars returns a list of env names for the given attributes.
func getEnvVars(attributes []StringAttribute) []string {
	envVars := make([]string, len(attributes))
	for i, attr := range attributes {
		envVars[i] = attr.EnvVar
	}
	return envVars
}
