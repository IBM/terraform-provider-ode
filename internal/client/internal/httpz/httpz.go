// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package httpz

import "net/http"

// APIClient Is mainly to solve the issue of circular dependency between client and services.
// It also makes Mock tests easier
type APIClient interface {
	HTTPClient() *http.Client
	BaseURL() string
}
