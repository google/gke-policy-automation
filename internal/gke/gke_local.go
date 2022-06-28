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
	"context"
	"encoding/json"
	"fmt"
	"os"
)

type gkeLocalClient struct {
	readFileFunc func(name string) ([]byte, error)
	dumpFile     string
}

func NewGKELocalClient(ctx context.Context, dumpFile string) GKEClient {
	return &gkeLocalClient{
		readFileFunc: os.ReadFile,
		dumpFile:     dumpFile,
	}
}

// GetCluster() returns cluster data gathered from file
func (c *gkeLocalClient) GetCluster(name string) (*Cluster, error) {
	var clusters []*Cluster
	clusterData, err := c.readFileFunc(c.dumpFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(clusterData, &clusters)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}
	return nil, fmt.Errorf("cluster %s not found in a dump file", name)
}

func (c *gkeLocalClient) Close() error {
	return nil
}
