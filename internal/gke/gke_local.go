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
	"io/ioutil"
	"os"
)

type gkeLocalClient struct {
	ctx      context.Context
	dumpFile string
}

func NewGKELocalClient(ctx context.Context, dumpFile string) GKEClient {
	return &gkeLocalClient{ctx: ctx, dumpFile: dumpFile}
}

// GetClusterName() returns ClusterName from the file
/*
func (c *gkeLocalClient) GetClusterName() (string, error) {
	var err error
	var cluster containerpb.Cluster

	clusterData, err := openData(c.dumpFile)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(clusterData, &cluster)
	if err != nil {
		return "", err
	}
	return cluster.Name, err
}
*/

// GetCluster() returns cluster data gathered from file
func (c *gkeLocalClient) GetCluster(name string) (*Cluster, error) {
	var err error
	var cluster Cluster

	clusterData, err := openData(c.dumpFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(clusterData, &cluster)
	if err != nil {
		return &cluster, err
	}
	return &cluster, err
}

func (c *gkeLocalClient) Close() error {
	return nil
}

func openData(fileName string) ([]byte, error) {
	clusterDumpFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer clusterDumpFile.Close()

	byteValue, err := ioutil.ReadAll(clusterDumpFile)
	if err != nil {
		return nil, err
	}
	return byteValue, nil
}
