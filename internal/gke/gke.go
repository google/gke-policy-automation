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

// Package gke implements Google Kubernetes Engine specific features like cluster discovery
package gke

import (
	"fmt"
	"regexp"
)

func GetClusterID(project string, location string, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name)
}

func MustSliceClusterID(id string) (string, string, string) {
	p, l, c, err := SliceAndValidateClusterID(id)
	if err != nil {
		panic("invalid cluster id: " + err.Error())
	}
	return p, l, c
}

func SliceAndValidateClusterID(id string) (string, string, string, error) {
	r := regexp.MustCompile(`projects/(.+)/(locations|zones)/(.+)/clusters/(.+)`)
	if !r.MatchString(id) {
		return "", "", "", fmt.Errorf("input %q does not match regexp", id)
	}
	matches := r.FindStringSubmatch(id)
	if len(matches) != 5 {
		return "", "", "", fmt.Errorf("wrong number of matches, got %d, expected %d", len(matches), 5)
	}
	return matches[1], matches[3], matches[4], nil
}
