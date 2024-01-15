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

// Package inputs implements data inputs used by the application
package inputs

import (
	"fmt"
	"sync"

	"github.com/google/gke-policy-automation/internal/log"
)

const (
	defaultMaxDataGetCoroutines = 20
)

type Input interface {
	GetID() string
	GetDataSourceName() string
	GetDescription() string
	GetData(clusterID string) (interface{}, error)
	Close() error
}

type Cluster struct {
	Name string                 `json:"name"`
	Data map[string]interface{} `json:"data"`
}

type getDataTask struct {
	input     Input
	clusterID string
}

type getDataTaskResult struct {
	clusterID      string
	inputID        string
	dataSourceName string
	result         interface{}
	err            error
}

// GetAllInputsData fetches data from given inputs for all given clusters in a concurrent manner
func GetAllInputsData(inputs []Input, clusterIDs []string) (map[string]*Cluster, []error) {
	return GetAllInputsDataWithMaxGoRoutines(inputs, clusterIDs, defaultMaxDataGetCoroutines)
}

// GetAllInputsDataWithMaxGoRoutines fetches data from given inputs for all given clusters
// in a concurrent manner. The maxGoRoutines parameter determines concurrency level
func GetAllInputsDataWithMaxGoRoutines(inputs []Input, clusterIDs []string, maxGoRoutines int) (map[string]*Cluster, []error) {
	log.Infof("Fetching data from %d inputs for %d clusters", len(inputs), len(clusterIDs))
	log.Debugf("using %d maxGoRoutines", maxGoRoutines)
	tasksChan := make(chan *getDataTask, maxGoRoutines)
	resultsChan := make(chan *getDataTaskResult, maxGoRoutines)
	errorsChan := make(chan *getDataTaskResult, maxGoRoutines)

	log.Debugf("starting tasks producing goroutine")
	go func() {
		for _, input := range inputs {
			for _, clusterID := range clusterIDs {
				tasksChan <- &getDataTask{input: input, clusterID: clusterID}
			}
		}
		close(tasksChan)
	}()

	log.Debugf("starting tasks consuming goroutine")
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < maxGoRoutines; i++ {
			wg.Add(1)
			go getInputData(i, &wg, tasksChan, resultsChan, errorsChan)
		}
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()
	log.Debugf("processing results and errors")
	results := processResults(resultsChan)
	errors := processErrors(errorsChan)
	return results, errors
}

func getInputData(i int, wg *sync.WaitGroup, tasks chan *getDataTask, results chan *getDataTaskResult, errors chan *getDataTaskResult) {
	defer wg.Done()
	for task := range tasks {
		log.Debugf("goroutine %d fetching input %s for cluster %s", i, task.input.GetID(), task.clusterID)
		result, err := task.input.GetData(task.clusterID)
		if err != nil {
			log.Debugf("goroutine %d fetch error %s", i, err)
			errors <- &getDataTaskResult{clusterID: task.clusterID, inputID: task.input.GetID(), err: err}
		} else {
			log.Debugf("goroutine %d fetch success", i)
			results <- &getDataTaskResult{clusterID: task.clusterID, inputID: task.input.GetID(), dataSourceName: task.input.GetDataSourceName(), result: result}
		}
	}
	log.Debugf("goroutine %d done", i)
}

func processResults(resultsChan chan *getDataTaskResult) map[string]*Cluster {
	results := make(map[string]*Cluster)
	for result := range resultsChan {
		data, ok := results[result.clusterID]
		if !ok {
			data = &Cluster{Name: result.clusterID, Data: make(map[string]interface{})}
		}
		data.Data[result.dataSourceName] = result.result
		results[result.clusterID] = data
	}
	return results
}

func processErrors(errorsChan chan *getDataTaskResult) []error {
	errors := make([]error, 0, len(errorsChan))
	for err := range errorsChan {
		errors = append(errors,
			fmt.Errorf("failed to fetch data for cluster %s, input %s: %s", err.clusterID, err.inputID, err.err),
		)
	}
	return errors
}
