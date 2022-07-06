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
	"reflect"
	"testing"
)

func TestNewLocalDiscoveryClient(t *testing.T) {
	filename := "test.json"
	client := NewLocalDiscoveryClient(filename)
	localClient, ok := client.(*localDiscoveryClient)
	if !ok {
		t.Fatalf("client type is not *localDiscoveryClient")
	}
	if localClient.filename != filename {
		t.Errorf("client filename = %v; want %v", localClient.filename, filename)
	}
}

func TestLocalDiscoveryClientGetClusters(t *testing.T) {
	data := `[{"Name":"cluster-one"},{"Name":"cluster-two"}]`
	filename := "test.json"
	client := &localDiscoveryClient{
		filename: filename,
		readFileFunc: func(name string) ([]byte, error) {
			return []byte(data), nil
		},
	}
	clusters, err := client.getClusters()
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	expected := []string{"cluster-one", "cluster-two"}
	if !reflect.DeepEqual(clusters, expected) {
		t.Fatalf("clusters = %v; want %v", clusters, expected)
	}
}
