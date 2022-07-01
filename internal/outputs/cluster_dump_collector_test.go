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
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/google/gke-policy-automation/internal/gke"
	"google.golang.org/genproto/googleapis/container/v1"
)

func TestFileClusterDumpCollectorNew(t *testing.T) {
	fileName := "myfile.json"
	collector := NewFileClusterDumpCollector(fileName)

	fileCollector, ok := collector.(*fileClusterDumpCollector)
	if !ok {
		t.Fatalf("collector type is not *fileClusterDumpCollector")
	}
	if fileCollector.filename != fileName {
		t.Errorf("collector filename = %v; want %v", fileCollector.filename, fileName)
	}
	if reflect.ValueOf(fileCollector.writeFileFunc).Pointer() != reflect.ValueOf(os.WriteFile).Pointer() {
		t.Fatalf("collector writeFileFunc is not os.WriteFile")
	}
}

func TestFileClusterDumpCollectorRegisterCluster(t *testing.T) {
	clusters := []*gke.Cluster{
		{
			Cluster:   &container.Cluster{Name: "cluster-one"},
			Resources: []*gke.Resource{},
		},
		{
			Cluster:   &container.Cluster{Name: "cluster-two"},
			Resources: []*gke.Resource{},
		},
	}

	collector := fileClusterDumpCollector{}
	for _, cluster := range clusters {
		collector.RegisterCluster(cluster)
	}
	if !reflect.DeepEqual(collector.clusters, clusters) {
		t.Fatalf("collector clusters = %v; want %v", collector.clusters, clusters)
	}
}

func TestFileClusterDumpCollectorClose(t *testing.T) {
	fileName := "test.json"
	collector := fileClusterDumpCollector{
		filename: fileName,
		clusters: []*gke.Cluster{
			{
				Cluster:   &container.Cluster{Name: "cluster-one"},
				Resources: []*gke.Resource{},
			},
		},
		writeFileFunc: func(name string, data []byte, perm os.FileMode) error {
			if name != fileName {
				t.Fatalf("filename = %v; want %v", name, fileName)
			}
			if perm != 0644 {
				t.Fatalf("perm = %v; want %v", perm, 0644)
			}
			return nil
		},
	}
	collector.Close()
}

func TestOutputClusterDumpCollectorNew(t *testing.T) {
	output := NewSilentOutput()
	collector := NewOutputClusterDumpCollector(output)

	outputCollector, ok := collector.(*outputClusterDumpCollector)
	if !ok {
		t.Fatalf("collector type is not *outputClusterDumpCollector")
	}
	if outputCollector.output != output {
		t.Errorf("collector output = %v; want %v", outputCollector.output, output)
	}
}

func TestOutputClusterDumpCollectorClose(t *testing.T) {
	var buff bytes.Buffer
	output := &Output{w: &buff}
	collector := NewOutputClusterDumpCollector(output)
	collector.RegisterCluster(&gke.Cluster{
		Cluster:   &container.Cluster{Name: "cluster-one"},
		Resources: []*gke.Resource{}})

	collector.Close()
	if len(buff.String()) <= 0 {
		t.Fatalf("len of output buffer is 0")
	}
}
