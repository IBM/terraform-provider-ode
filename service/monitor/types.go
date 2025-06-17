// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package monitor

type Data struct {
	Tasks             []Task       `json:"tasks"`
	CurrentTask       int          `json:"current-task"`
	OverallPercentage float64      `json:"overall-percentage"`
	Done              bool         `json:"done"`
	Error             *ErrorRecord `json:"error"`
}

type Task struct {
	ID         int `json:"id"`
	Percentage int `json:"percentage"`
}

type ErrorRecord struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Cause      string `json:"cause"`
	Resolution string `json:"resolution"`
	Level      string `json:"level"`
}
