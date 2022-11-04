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

package gke

import (
	"encoding/json"
	"os"

	"github.com/google/gke-policy-automation/internal/inputs"
)

type localDiscoveryClient struct {
	readFileFunc func(name string) ([]byte, error)
	filename     string
}

func (c *localDiscoveryClient) Close() error {
	return nil
}

func (c *localDiscoveryClient) GetClustersInFolder(number string) ([]string, error) {
	return c.getClusters()
}

func (c *localDiscoveryClient) GetClustersInOrg(number string) ([]string, error) {
	return c.getClusters()
}

func (c *localDiscoveryClient) GetClustersInProject(name string) ([]string, error) {
	return c.getClusters()

}

func NewLocalDiscoveryClient(filename string) DiscoveryClient {
	return &localDiscoveryClient{
		readFileFunc: os.ReadFile,
		filename:     filename,
	}
}

func (c *localDiscoveryClient) getClusters() ([]string, error) {
	data, err := c.readFileFunc(c.filename)
	if err != nil {
		return nil, err
	}

	var clusters []*inputs.Cluster
	if err = json.Unmarshal(data, &clusters); err != nil {
		return nil, err
	}
	names := make([]string, len(clusters))
	for i := range clusters {
		names[i] = clusters[i].Name
	}
	return names, nil
}
