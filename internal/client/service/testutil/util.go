// Copyright (c) IBM Corporation
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockServer struct {
	*httptest.Server
	handlers map[string]http.HandlerFunc
}

func NewServer(t testing.TB) *MockServer {
	ms := &MockServer{handlers: make(map[string]http.HandlerFunc)}
	ms.Server = httptest.NewServer(http.HandlerFunc(ms.handle))
	t.Cleanup(ms.Close)
	return ms
}

func (ms *MockServer) handle(w http.ResponseWriter, r *http.Request) {
	key := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	if h, ok := ms.handlers[key]; ok {
		h(w, r)
		return
	}
	http.Error(w, "not found", http.StatusNotFound)
}

func (ms *MockServer) RegisterHandler(method, path string, h http.HandlerFunc) {
	key := fmt.Sprintf("%s %s", method, path)
	ms.handlers[key] = h
}

func (ms *MockServer) RegisterJSON(method, path string, status int, body interface{}) {
	ms.RegisterHandler(
		method, path, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(body)
		},
	)
}

type APIClient interface {
	BaseURL() string
	HTTPClient() *http.Client
}

func NewMockAPI(ms *MockServer) APIClient {
	return &mockAPI{base: ms.URL, client: ms.Client()}
}

type mockAPI struct {
	base   string
	client *http.Client
}

func (m *mockAPI) BaseURL() string          { return m.base }
func (m *mockAPI) HTTPClient() *http.Client { return m.client }
