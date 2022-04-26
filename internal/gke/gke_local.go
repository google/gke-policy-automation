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
	"io/ioutil"
	"os"

	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

type GKELocalClient struct {
	ctx      context.Context
	dumpFile string
}

func NewLocalClient(ctx context.Context, dumpFile string) (*GKELocalClient, error) {
	fmt.Printf("log file: %s\n", dumpFile)
	return &GKELocalClient{ctx: ctx, dumpFile: dumpFile}, nil
}

func (c *GKELocalClient) GetClusterName() (string, error) {
	var cluster containerpb.Cluster

	clusterDumpFile, err := os.Open(c.dumpFile)
	if err != nil {
		return "", err
	}
	defer clusterDumpFile.Close()

	byteValue, err := ioutil.ReadAll(clusterDumpFile)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(byteValue, &cluster)
	if err != nil {
		return "", err
	}
	fmt.Println("Successfully Opened cluster data file")
	return cluster.Name, err
}

// to add json file path
func (c *GKELocalClient) GetCluster() (*containerpb.Cluster, error) {
	var cluster containerpb.Cluster

	clusterDumpFile, err := os.Open(c.dumpFile)
	if err != nil {
		return &cluster, err
	}
	defer clusterDumpFile.Close()

	byteValue, err := ioutil.ReadAll(clusterDumpFile)
	if err != nil {
		return &cluster, err
	}

	err = json.Unmarshal(byteValue, &cluster)
	if err != nil {
		return &cluster, err
	}
	fmt.Println("Successfully Opened cluster data file")
	// defer the closing of our jsonFile so that we can parse it later on

	return &cluster, err
}
