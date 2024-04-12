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
	"fmt"
	"regexp"
	"testing"
)

func TestGetClusterID(t *testing.T) {
	projectID := "test-project"
	clusterLocation := "europe-central2"
	clusterName := "warsaw"
	name := GetClusterID(projectID, clusterLocation, clusterName)
	re := regexp.MustCompile(`^projects/([^/]+)/locations/([^/]+)/clusters/([^/]+)$`)
	if !re.MatchString(name) {
		t.Fatalf("name: %q, does not match regexp: %q", name, re.String())
	}
	matches := re.FindStringSubmatch(name)
	if matches[1] != projectID {
		t.Errorf("match[1] = %v; want %v", matches[1], projectID)
	}
	if matches[2] != clusterLocation {
		t.Errorf("match[2] = %v; want %v", matches[2], clusterLocation)
	}
	if matches[3] != clusterName {
		t.Errorf("match[3] = %v; want %v", matches[3], clusterName)
	}
}

func TestSliceAndValidateClusterID(t *testing.T) {
	projectID := "demo-project-123"
	clusterLocation := "europe-central2"
	clusterName := "cluster-waw"
	input := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, clusterLocation, clusterName)

	resultProjectID, resultClusterLocation, resultClusterName, err := SliceAndValidateClusterID(input)
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	if resultProjectID != projectID {
		t.Errorf("projectID = %v; want %v", resultProjectID, projectID)
	}
	if resultClusterLocation != clusterLocation {
		t.Errorf("clusterLocation = %v; want %v", resultProjectID, clusterLocation)
	}
	if resultClusterName != clusterName {
		t.Errorf("clusterName = %v; want %v", resultProjectID, clusterName)
	}
}

func TestSliceAndValidateClusterID_negative(t *testing.T) {
	input := ("projects/demo-project-123/regions/europe-central2/clusters/cluster-waw")
	_, _, _, err := SliceAndValidateClusterID(input)
	if err == nil {
		t.Fatalf("err = nil; want err")
	}
}

func TestMustSliceClusterID(t *testing.T) {
	input := "projects/demo-project-123/locations/europe-central2/clusters/cluster-waw"
	MustSliceClusterID(input)
}
