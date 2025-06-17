// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package httpz

import "errors"

var (
	ErrInvalidURL   = errors.New("httpz: invalid URL")
	ErrInvalidParam = errors.New("httpz: invalid parameter")

	ErrNilWriter   = errors.New("httpz: nil writer")
	ErrNilReader   = errors.New("httpz: nil reader")
	ErrNilResponse = errors.New("httpz: nil response")
)
