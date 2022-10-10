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
	"errors"
	"reflect"
	"sync"
	"testing"
)

type inputMock struct {
	getIDFn          func() string
	getDescriptionFn func() string
	getDataFn        func(clusterID string) (interface{}, error)
}

func (m inputMock) GetID() string {
	return m.getIDFn()
}

func (m inputMock) GetDescription() string {
	return m.getDescriptionFn()
}

func (m inputMock) GetData(clusterID string) (interface{}, error) {
	return m.getDataFn(clusterID)
}

func TestGetAllInputsData(t *testing.T) {
	clusterIDs := []string{"cluster-one", "cluster-two", "cluster-three", "cluster-four"}
	inputs := []Input{
		&inputMock{getIDFn: func() string { return "gke-api" }, getDataFn: func(clusterID string) (interface{}, error) { return "data", nil }},
		&inputMock{getIDFn: func() string { return "kube-api" }, getDataFn: func(clusterID string) (interface{}, error) { return "data", nil }},
		&inputMock{getIDFn: func() string { return "metrics-api" }, getDataFn: func(clusterID string) (interface{}, error) { return "data", nil }},
		&inputMock{getIDFn: func() string { return "bad-api" }, getDataFn: func(clusterID string) (interface{}, error) { return nil, errors.New("error") }},
	}
	results, errors := GetAllInputsData(inputs, clusterIDs)
	if len(results) != len(clusterIDs) {
		t.Fatalf("number of results = %v; want %v", len(results), len(clusterIDs))
	}
	if len(errors) != len(clusterIDs) {
		t.Fatalf("number of errors = %v; want %v", len(errors), len(clusterIDs))
	}
}

func TestGetInputData(t *testing.T) {
	tasksNo := 2
	errClusterID := "cluster-one"
	okClusterID := "cluster-two"
	okResult := "test data 123"
	tasksChan := make(chan *getDataTask, tasksNo)
	tasksChan <- &getDataTask{
		input: &inputMock{
			getIDFn: func() string {
				return "gke-api"
			},
			getDataFn: func(clusterID string) (interface{}, error) {
				return nil, errors.New("test error")
			},
		},
		clusterID: errClusterID,
	}
	tasksChan <- &getDataTask{
		input: &inputMock{
			getIDFn: func() string {
				return "gke-api"
			},
			getDataFn: func(clusterID string) (interface{}, error) {
				return okResult, nil
			},
		},
		clusterID: okClusterID,
	}
	close(tasksChan)
	resultsChan := make(chan *getDataTaskResult, tasksNo)
	errorsChan := make(chan *getDataTaskResult, tasksNo)
	var wg sync.WaitGroup
	wg.Add(1)
	getInputData(0, &wg, tasksChan, resultsChan, errorsChan)
	close(resultsChan)
	close(errorsChan)
	if len(resultsChan) != 1 {
		t.Fatalf("number of results = %v; want %v", len(resultsChan), 1)
	}
	if len(errorsChan) != 1 {
		t.Fatalf("number of errors = %v; want %v", len(errorsChan), 1)
	}
	for result := range resultsChan {
		if result.clusterID != okClusterID {
			t.Errorf("results clusterID = %v; want %v", result.clusterID, okClusterID)
		}
		if !reflect.DeepEqual(result.result, okResult) {
			t.Errorf("results result is not %v", result)
		}
	}
	for result := range errorsChan {
		if result.clusterID != errClusterID {
			t.Errorf("errors clusterID = %v; want %v", result.clusterID, errClusterID)
		}
		if result.err == nil {
			t.Errorf("errors err is nil; want error")
		}
	}
}

func TestProcessResults(t *testing.T) {
	resultsChan := make(chan *getDataTaskResult, 3)
	resultsChan <- &getDataTaskResult{
		clusterID: "cluster-one",
		inputID:   "gke-api",
		result:    "cluster-one-gcp-data",
	}
	resultsChan <- &getDataTaskResult{
		clusterID: "cluster-one",
		inputID:   "kube-api",
		result:    "cluster-one-kube-data",
	}
	resultsChan <- &getDataTaskResult{
		clusterID: "cluster-two",
		inputID:   "gke-api",
		result:    "cluster-two-gcp-data",
	}
	close(resultsChan)
	result := processResults(resultsChan)
	cOneResults, ok := result["cluster-one"]
	if !ok {
		t.Fatalf("no results for cluster-one")
	}
	cTwoResults, ok := result["cluster-two"]
	if !ok {
		t.Fatalf("no results for cluster-two")
	}
	if data := cOneResults.data["gke-api"]; data != "cluster-one-gcp-data" {
		t.Errorf("cluster-one results for gcp-api: %v; want %v", data, "cluster-one-gcp-data")
	}
	if data := cOneResults.data["kube-api"]; data != "cluster-one-kube-data" {
		t.Errorf("cluster-one results for kube-api: %v; want %v", data, "cluster-one-kube-data")
	}
	if data := cTwoResults.data["gke-api"]; data != "cluster-two-gcp-data" {
		t.Errorf("cluster-two results for kube-api: %v; want %v", data, "cluster-two-gcp-data")
	}
}

func TestProcessErrors(t *testing.T) {
	errorsChan := make(chan *getDataTaskResult, 2)
	errorsChan <- &getDataTaskResult{
		clusterID: "cluster-one",
		inputID:   "gke-api",
		err:       errors.New("error"),
	}
	errorsChan <- &getDataTaskResult{
		clusterID: "cluster-two",
		inputID:   "gke-api",
		err:       errors.New("error"),
	}
	close(errorsChan)
	errors := processErrors(errorsChan)
	if len(errors) != 2 {
		t.Fatalf("number of errors = %v; want %v", len(errors), 2)
	}
}
