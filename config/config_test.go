// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package config

import "testing"

func TestFillDefaultsAndValidate(t *testing.T) {
	tests := []struct {
		name    string
		in      Config
		wantErr bool
	}{
		{"missing base", Config{User: "u", Pass: "p"}, true},
		{"missing user", Config{BaseURL: "https://ode.com"}, true},
		{"missing pass", Config{BaseURL: "https://ode.com", User: "u"}, true},
		{"valid minimal", Config{BaseURL: "https://ode.com", User: "u", Pass: "p"}, false},
	}

	for _, tc := range tests {
		tc.in.FillDefaults()
		err := tc.in.Validate()
		if tc.wantErr && err == nil {
			t.Fatalf("want: %s: expected error", tc.name)
		}
		if !tc.wantErr && err != nil {
			t.Fatalf("want: %s: got unexpected error %v", tc.name, err)
		}
		if !tc.wantErr && tc.in.RequestTimeout != DefaultRequestTimeout {
			t.Fatalf("want: %s: defaults not applied", tc.name)
		}
	}
}
