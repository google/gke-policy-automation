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

	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

// TestLocalGetClusterName() to test GetClusterName()
func TestLocalGetClusterName(t *testing.T) {
	var clusterName string
	client, err := NewGKELocalClient(context.TODO(), "../app/test-fixtures/backup.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if clusterName, err = client.GetClusterName(); err != nil {
		t.Fatalf(err.Error())
	}
	if clusterName == "" {
		t.Errorf("unable to get cluster name")
	}
}

// TestLocalGetCluster() to test GetCluster()
func TestLocalGetCluster(t *testing.T) {
	var cluster *containerpb.Cluster
	client, err := NewGKELocalClient(context.TODO(), "../app/test-fixtures/backup.json")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if cluster, err = client.GetCluster(); err != nil {
		t.Fatalf(err.Error())
	}
	if cluster == nil || cluster.Network != "default" {
		t.Errorf("unable to read cluster data")
	}

}
