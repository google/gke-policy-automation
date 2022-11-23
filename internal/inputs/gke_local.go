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
	"encoding/json"
	"fmt"
	"os"

	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

const (
	gkeLocalInputID          = "gkeLocal"
	gkeLocalDataSourceName   = "gke"
	gkeLocalInputDescription = "GKE cluster data from JSON dump"
)

type gkeLocalInput struct {
	readFileFunc func(name string) ([]byte, error)
	dumpFile     string
}

func NewGKELocalInput(dumpFile string) Input {
	return &gkeLocalInput{
		readFileFunc: os.ReadFile,
		dumpFile:     dumpFile,
	}
}

func (i *gkeLocalInput) GetID() string {
	return gkeLocalInputID
}

func (i *gkeLocalInput) GetDescription() string {
	return gkeLocalInputDescription
}

func (i *gkeLocalInput) GetDataSourceName() string {
	return gkeLocalDataSourceName
}

func (i *gkeLocalInput) GetData(clusterID string) (interface{}, error) {
	var clusters []*containerpb.Cluster
	data, err := i.readFileFunc(i.dumpFile)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &clusters); err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		if cluster.Name == clusterID {
			return cluster, nil
		}
	}
	return nil, fmt.Errorf("cluster %s not found in a dump file", clusterID)
}

func (i *gkeLocalInput) Close() error {
	return nil
}
