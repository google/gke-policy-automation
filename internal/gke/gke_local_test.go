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
	"testing"
)

func TestNewGKELocalClient(t *testing.T) {
	filename := "test.json"
	client := NewGKELocalClient(context.TODO(), filename)

	localClient, ok := client.(*gkeLocalClient)
	if !ok {
		t.Fatalf("client type is not *gkeLocalClient")
	}
	if localClient.dumpFile != filename {
		t.Errorf("client dumpFile = %v; want %v", localClient.dumpFile, filename)
	}
}

// TestLocalGetCluster() to test GetCluster()
func TestLocalGetCluster(t *testing.T) {
	clusterNames := []string{"cluster-1", "cluster-2"}
	client := NewGKELocalClient(context.TODO(), "test-fixtures/clusters_data.json")

	for _, clusterName := range clusterNames {
		cluster, err := client.GetCluster(clusterName)
		if err != nil {
			t.Fatalf("err = %v; want nil", err)
		}
		if cluster.Name != clusterName {
			t.Errorf("cluster name = %v; want %v", cluster.Name, clusterName)
		}
	}
}
