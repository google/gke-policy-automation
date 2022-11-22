// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package inputs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/gke-policy-automation/internal/version"
)

type mockRoundTripper struct {
	RoundTripFn func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return m.RoundTripFn(r)
}

func TestRestInputGetID(t *testing.T) {
	input := restInput{}
	if id := input.GetID(); id != restInputID {
		t.Errorf("id = %v; want %v", id, restInputID)
	}
}

func TestRestInputGetDescription(t *testing.T) {
	input := restInput{}
	if desc := input.GetDescription(); desc != restInputDescription {
		t.Errorf("desc = %v; want %v", desc, restInputDescription)
	}
}

func TestRestInputClose(t *testing.T) {
	input := restInput{}
	if err := input.Close(); err != nil {
		t.Errorf("err = %v; want nil", err)
	}
}

func TestRestInputGetData(t *testing.T) {
	input := NewRestInput(context.Background(), "http://blabla.com/CLUSTER_ID/metrics")
	restInput := input.(*restInput)
	clusterID := "some/cluster/id"
	restInput.client = &http.Client{
		Transport: &mockRoundTripper{
			RoundTripFn: func(r *http.Request) (*http.Response, error) {
				expectedURL := fmt.Sprintf("http://blabla.com/%s/metrics", clusterID)
				if r.URL.String() != expectedURL {
					t.Fatalf("request URL = %v; want %v", r.URL.String(), expectedURL)
				}
				responseData := "{\"id\":\"demo\"}"
				reader := strings.NewReader(responseData)
				readerCloser := io.NopCloser(reader)
				return &http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Body:       readerCloser,
				}, nil
			},
		},
	}

	_, err := input.GetData(clusterID)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
}

func TestRestInputReadResponseBody(t *testing.T) {
	key := "id"
	value := "test"
	data := fmt.Sprintf("{%q:%q}", key, value)
	reader := strings.NewReader(data)
	readerCloser := io.NopCloser(reader)

	result, err := readResponseBody(readerCloser)
	if err != nil {
		t.Errorf("err = %v; want nil", err)
	}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not  = %v; want nil", err)
	}
	if resultMap[key] != value {
		t.Errorf("result key %v = %v; want %v", key, resultMap[key], value)
	}
}

func TestRestInputCreateGetRequest(t *testing.T) {
	endpoint := "https://endpoint.com/resource/item"
	req, err := createGetRequest(context.Background(), endpoint)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	if req.UserAgent() != version.UserAgent {
		t.Errorf("userAgent = %v; want %v", req.UserAgent(), version.UserAgent)
	}
	if req.Method != http.MethodGet {
		t.Errorf("method = %v; want %v", req.Method, http.MethodGet)
	}
	if req.URL.String() != endpoint {
		t.Errorf("URL = %v; want %v", req.URL.String(), endpoint)
	}
}

func TestRestInputReplaceWildcard(t *testing.T) {
	wildcard := clusterIDWildcard
	input := fmt.Sprintf("https://test.com/%s/data", wildcard)
	cluster := "projects/test-project/locations/test-region/clusters/test-cluster"
	expected := fmt.Sprintf("https://test.com/%s/data", cluster)

	result := replaceWildcard(input, wildcard, cluster)
	if result != expected {
		t.Errorf("result = %v; want %v", result, expected)
	}
}
