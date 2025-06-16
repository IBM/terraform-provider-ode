// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.ibm.com/terraformz-poc/terraform-provider-ode/internal/client/internal/httpz"
)

const pathImages = "/odtz/api/image-services/v1/images"

// Service represents an image service client.
type Service struct {
	api httpz.APIClient
}

// New creates a new image service client.
func New(client httpz.APIClient) *Service {
	return &Service{api: client}
}

// List retrieves a list of images based on the provided input.
// Filtering by UUID, label, or version is done client side.
func (s *Service) List(ctx context.Context, input ListInput) ([]StockImage, error) {
	query := url.Values{}

	if input.Versions {
		query.Set("versions", "true")
	}
	if len(input.SourceEnvironmentUUIDs) > 0 {
		query.Set("source-environment-uuid", strings.Join(input.SourceEnvironmentUUIDs, ","))
	}

	groups, err := httpz.Do[listResponse](
		ctx,
		s.api,
		http.MethodGet,
		pathImages,
		httpz.Query(query),
	)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}

	var images []StockImage
	for _, group := range groups {
		images = append(images, group.StockImages...)
	}

	return filterImages(images, input.UUID, input.Label, input.Version), nil
}

// filterImages applies client side filtering by UUID, label, and version.
func filterImages(images []StockImage, uuid, label string, version *int) []StockImage {
	var filtered []StockImage
	for _, img := range images {
		if (uuid == "" || img.UUID == uuid) &&
			(label == "" || img.Name == label) &&
			(version == nil || img.Version == *version) {
			filtered = append(filtered, img)
		}
	}
	return filtered
}
