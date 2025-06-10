// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/httpz"
	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/service/testutil"
)

func init()          { httpz.SetAllowInsecure(true) }
func ptr(i int) *int { return &i }

func TestImageService_List(t *testing.T) {
	tests := []struct {
		name       string
		input      ListInput
		mockGroups []Group
		want       []StockImage
	}{
		{
			name: "no filters should return all images.",
			input: ListInput{
				SourceEnvironmentUUIDs: []string{"env-1"},
				Versions:               true,
			},
			mockGroups: []Group{
				{Name: "g1", StockImages: []StockImage{
					{UUID: "uuid11", Name: "img11", Version: 1},
					{UUID: "uuid22", Name: "img22", Version: 2},
					{UUID: "uuid33", Name: "img33", Version: 2},
				}},
			},
			want: []StockImage{
				{UUID: "uuid11", Name: "img11", Version: 1},
				{UUID: "uuid22", Name: "img22", Version: 2},
				{UUID: "uuid33", Name: "img33", Version: 2},
			},
		},
		{
			name: "filter by UUID should return a single image matching the provided UUID.",
			input: ListInput{
				UUID: "uuid22",
			},
			mockGroups: []Group{
				{Name: "group11", StockImages: []StockImage{
					{UUID: "uuid11", Name: "img11", Version: 1},
					{UUID: "uuid22", Name: "img22", Version: 2},
					{UUID: "uuid33", Name: "img33", Version: 2},
				}},
			},
			want: []StockImage{
				{UUID: "uuid22", Name: "img22", Version: 2},
			},
		},
		{
			name: "filter by label and version should return a single image matching the provided label and version.",
			input: ListInput{
				Label:   "img1",
				Version: ptr(1),
			},
			mockGroups: []Group{
				{Name: "g1", StockImages: []StockImage{
					{UUID: "u1", Name: "img1", Version: 1},
					{UUID: "u2", Name: "img1", Version: 2},
					{UUID: "u3", Name: "img2", Version: 1},
				}},
			},
			want: []StockImage{
				{UUID: "u1", Name: "img1", Version: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms := testutil.NewServer(t)
				ms.RegisterJSON(http.MethodGet, pathImages, http.StatusOK, tt.mockGroups)

				svc := New(testutil.NewMockAPI(ms))

				got, err := svc.List(context.Background(), tt.input)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("got: %v, want: %v", got, tt.want)
				}
			},
		)
	}
}
