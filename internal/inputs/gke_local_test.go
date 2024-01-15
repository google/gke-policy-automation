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
	"fmt"
	"testing"

	"cloud.google.com/go/container/apiv1/containerpb"
)

func TestNewGKELocalInput(t *testing.T) {
	filename := "test.json"
	input := NewGKELocalInput(filename)

	gkeLocalInput, ok := input.(*gkeLocalInput)
	if !ok {
		t.Fatalf("input type is not *gkeLocalInput")
	}
	if gkeLocalInput.dumpFile != filename {
		t.Errorf("input dumpFile = %v; want %v", gkeLocalInput.dumpFile, filename)
	}
}

func TestGKELocalGetId(t *testing.T) {
	input := gkeLocalInput{}
	if id := input.GetID(); id != gkeLocalInputID {
		t.Fatalf("id = %v; want %v", id, gkeLocalInputID)
	}
}

func TestGKELocalGetDescription(t *testing.T) {
	input := gkeLocalInput{}
	if id := input.GetDescription(); id != gkeLocalInputDescription {
		t.Fatalf("id = %v; want %v", id, gkeLocalInputDescription)
	}
}

func TestGKEGetData(t *testing.T) {
	fileName := "dump/file.json"
	clusterName := "cluster-test-01"
	clusterCIDR := "10.84.0.0/14"
	clusterJSON := fmt.Sprintf("[\n"+
		"  {\n"+
		"    \"name\": %q,\n"+
		"    \"network\": \"default\",\n"+
		"    \"cluster_ipv4_cidr\": %q\n"+
		"  }\n,"+
		"  {\n"+
		"    \"name\": \"cluster-test-02\",\n"+
		"    \"network\": \"default\",\n"+
		"    \"cluster_ipv4_cidr\": \"127.16.0.0/14\"\n"+
		"  }\n"+
		"]", clusterName, clusterCIDR)

	input := gkeLocalInput{
		readFileFunc: func(name string) ([]byte, error) {
			if name != fileName {
				t.Errorf("fileName = %v; want %v", name, fileName)
			}
			return []byte(clusterJSON), nil
		},
		dumpFile: fileName,
	}
	data, err := input.GetData(clusterName)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	clusterData, ok := data.(*containerpb.Cluster)
	if !ok {
		t.Fatalf("data is not *containerpb.Cluster")
	}
	if clusterData.Name != clusterName {
		t.Errorf("name = %v; want %v", clusterData.Name, clusterName)
	}
	if clusterData.ClusterIpv4Cidr != clusterCIDR {
		t.Errorf("name = %v; want %v", clusterData.ClusterIpv4Cidr, clusterCIDR)
	}
}

func TestGKELocalClose(t *testing.T) {
	input := gkeLocalInput{}
	if err := input.Close(); err != nil {
		t.Errorf("gkeLocalInput close() error = %v ; want nil", err)
	}
}
