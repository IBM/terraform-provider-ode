// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package transport

import "net/http"

// defaultTransport Sets the next transport to default if nil.
func defaultTransport(next http.RoundTripper) http.RoundTripper {
	if next == nil {
		return http.DefaultTransport
	}
	return next
}
