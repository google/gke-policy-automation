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

package outputs

import (
	"encoding/json"
	"os"

	"github.com/google/gke-policy-automation/internal/inputs"
)

type fileClusterDumpCollector struct {
	writeFileFunc func(name string, data []byte, perm os.FileMode) error
	filename      string
	clusters      []*inputs.Cluster
}

func (c *fileClusterDumpCollector) RegisterCluster(cluster *inputs.Cluster) {
	c.clusters = append(c.clusters, cluster)
}

func (c *fileClusterDumpCollector) Close() error {
	data, err := json.MarshalIndent(c.clusters, "", "    ")
	if err != nil {
		return err
	}
	return c.writeFileFunc(c.filename, data, 0644)
}

func NewFileClusterDumpCollector(filename string) ClusterDumpCollector {
	return &fileClusterDumpCollector{
		writeFileFunc: os.WriteFile,
		filename:      filename,
	}
}

type outputClusterDumpCollector struct {
	output   *Output
	clusters []*inputs.Cluster
}

func (c *outputClusterDumpCollector) RegisterCluster(cluster *inputs.Cluster) {
	c.clusters = append(c.clusters, cluster)
}

func (c *outputClusterDumpCollector) Close() (err error) {
	data, err := json.MarshalIndent(c.clusters, "", "    ")
	if err != nil {
		return
	}
	_, err = c.output.Printf(string(data))
	return
}

func NewOutputClusterDumpCollector(output *Output) ClusterDumpCollector {
	return &outputClusterDumpCollector{
		output: output,
	}
}
