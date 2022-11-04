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

	"github.com/google/gke-policy-automation/internal/inputs"
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

	cluster := &inputs.Cluster{Data: make(map[string]interface{})}

	cluster.Data["gkeAPI"] = &container.Cluster{Name: "cluster-one"}

	collector := fileClusterDumpCollector{}
	collector.RegisterCluster(cluster)
	if !reflect.DeepEqual(collector.clusters[0], cluster) {
		t.Fatalf("collector clusters = %v; want %v", collector.clusters[0], cluster)
	}
}

func TestFileClusterDumpCollectorClose(t *testing.T) {
	fileName := "test.json"
	cluster := &inputs.Cluster{Data: make(map[string]interface{})}
	cluster.Data["gkeAPI"] = &container.Cluster{Name: "cluster-one"}
	collector := fileClusterDumpCollector{
		filename: fileName,
		clusters: []*inputs.Cluster{cluster},
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
	cluster := &inputs.Cluster{Data: make(map[string]interface{})}
	cluster.Data["gkeAPI"] = &container.Cluster{Name: "cluster-one"}
	collector := NewOutputClusterDumpCollector(output)
	collector.RegisterCluster(cluster)

	collector.Close()
	if len(buff.String()) <= 0 {
		t.Fatalf("len of output buffer is 0")
	}
}
