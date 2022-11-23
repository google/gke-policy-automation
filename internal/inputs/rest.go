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
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/google/gke-policy-automation/internal/version"
)

const (
	restInputID                 = "rest"
	restInputDescription        = "Generic REST API input with HTTPs transport and JSON encoding. CLUSTER_ID wildcard can be used in the endpoint path."
	restDataSourceName          = "rest"
	clusterIDWildcard           = "CLUSTER_ID"
	defaultClientTimeoutSeconds = 3
)

type restInput struct {
	ctx      context.Context
	client   *http.Client
	endpoint string
}

func NewRestInput(ctx context.Context, endpoint string) Input {
	client := &http.Client{
		Timeout: time.Duration(defaultClientTimeoutSeconds) * time.Second,
	}

	return &restInput{
		ctx:      ctx,
		client:   client,
		endpoint: endpoint,
	}
}

func (i *restInput) GetID() string {
	return restInputID
}

func (i *restInput) GetDescription() string {
	return restInputDescription
}

func (i *restInput) GetDataSourceName() string {
	return restDataSourceName
}

func (i *restInput) GetData(clusterID string) (interface{}, error) {
	endpoint := replaceWildcard(i.endpoint, clusterIDWildcard, clusterID)
	req, err := createGetRequest(i.ctx, endpoint)
	if err != nil {
		return nil, err
	}
	resp, err := i.client.Do(req)
	if err != nil {
		return nil, err
	}
	return readResponseBody(resp.Body)
}

func (i *restInput) Close() error {
	return nil
}

func readResponseBody(bodyReader io.ReadCloser) (interface{}, error) {
	defer bodyReader.Close()
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func createGetRequest(ctx context.Context, endpoint string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", version.UserAgent)
	return req, nil
}

func replaceWildcard(input, wildcard, value string) string {
	r := regexp.MustCompile(wildcard)
	return r.ReplaceAllString(input, value)
}
