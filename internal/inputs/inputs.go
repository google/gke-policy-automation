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
	"sync"

	"github.com/google/gke-policy-automation/internal/log"
)

const (
	defaultMaxDataGetCoroutines = 20
)

type Input interface {
	GetID() string
	GetDescription() string
	GetData(clusterID string) (interface{}, error)
	Close() error
}

type Cluster struct {
	Data map[string]interface{}
}

type getDataTask struct {
	input     Input
	clusterID string
}

type getDataTaskResult struct {
	clusterID string
	inputID   string
	result    interface{}
	err       error
}

//GetAllInputsData fetches data from given inputs for all given clusters in a concurrent manner
func GetAllInputsData(inputs []Input, clusterIDs []string) (map[string]*Cluster, []error) {
	tasksNo := len(inputs) * len(clusterIDs)
	tasksChan := make(chan *getDataTask, tasksNo)
	resultsChan := make(chan *getDataTaskResult, tasksNo)
	errorsChan := make(chan *getDataTaskResult, tasksNo)
	log.Infof("creating %d get data tasks", tasksNo)

	for _, input := range inputs {
		for _, clusterID := range clusterIDs {
			tasksChan <- &getDataTask{input: input, clusterID: clusterID}
		}
	}
	close(tasksChan)

	log.Debugf("starting %d goroutines", defaultMaxDataGetCoroutines)
	var wg sync.WaitGroup
	for i := 0; i < defaultMaxDataGetCoroutines; i++ {
		wg.Add(1)
		go getInputData(i, &wg, tasksChan, resultsChan, errorsChan)
	}
	wg.Wait()
	close(resultsChan)
	close(errorsChan)
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
			results <- &getDataTaskResult{clusterID: task.clusterID, inputID: task.input.GetID(), result: result}
		}
	}
	log.Debugf("goroutine %d done", i)
}

func processResults(resultsChan chan *getDataTaskResult) map[string]*Cluster {
	results := make(map[string]*Cluster)
	for result := range resultsChan {
		data, ok := results[result.clusterID]
		if !ok {
			data = &Cluster{Data: make(map[string]interface{})}
		}
		data.Data[result.inputID] = result.result
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
