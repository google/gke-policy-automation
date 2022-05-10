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
	"testing"
	"time"

	"github.com/google/gke-policy-automation/internal/policy"
	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) BucketExists(bucketName string) bool {
	args := m.Called(bucketName)
	return args.Bool(0)
}

func (m *Mock) Write(bucketName, objectName string, content []byte) error {
	args := m.Called(bucketName, objectName, content)
	return args.Error(0)
}

func (m *Mock) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *Mock) MockMapper(evaluationResult []*policy.PolicyEvaluationResult, time time.Time) ([]byte, error) {
	args := m.Called(evaluationResult, time)
	return args.Get(0).([]byte), args.Error(1)
}

func TestCloudStorageCollector(t *testing.T) {

	bucket := "mybucket"
	object := "myobject"
	mappedResult := []byte("result")

	policyResult := policy.NewPolicyEvaluationResult()
	policyResult.AddPolicy(&policy.Policy{
		Title:            "title",
		Description:      "description",
		Group:            "group",
		Valid:            true,
		Violations:       []string{},
		ProcessingErrors: []error{},
	})

	t.Run("happy path", func(t *testing.T) {

		evalResults := make([]*policy.PolicyEvaluationResult, 0)
		evalResults = append(evalResults, policyResult)

		mockStorage := &Mock{}
		mockStorage.On("BucketExists", bucket).Return(true)
		mockStorage.On("MockMapper", evalResults, mock.Anything).Return(mappedResult, nil)
		mockStorage.On("Write", bucket, object, mappedResult).Return(nil)
		mockStorage.On("Close").Return(nil)

		var collector, _ = NewCloudStorageResultCollector(
			mockStorage,
			mockStorage.MockMapper,
			bucket,
			object,
		)

		collector.RegisterResult(evalResults)
		collector.Close()

		mockStorage.AssertExpectations(t)
	})

	t.Run("bucket does not exist", func(t *testing.T) {

		mockStorage := &Mock{}
		mockStorage.On("BucketExists", bucket).Return(false)

		var _, err = NewCloudStorageResultCollector(
			mockStorage,
			mockStorage.MockMapper,
			bucket,
			object,
		)

		if err == nil {
			t.Errorf("should have returned error")
		}
	})

	t.Run("collect multiple results", func(t *testing.T) {

		evalResults := make([]*policy.PolicyEvaluationResult, 0)
		evalResults = append(evalResults, policyResult)

		expectedResultsCollected := []*policy.PolicyEvaluationResult{policyResult, policyResult}

		mockStorage := &Mock{}
		mockStorage.On("BucketExists", bucket).Return(true)
		mockStorage.On("MockMapper", expectedResultsCollected, mock.Anything).Return(mappedResult, nil)
		mockStorage.On("Write", bucket, object, mappedResult).Return(nil)
		mockStorage.On("Close").Return(nil)

		var collector, _ = NewCloudStorageResultCollector(
			mockStorage,
			mockStorage.MockMapper,
			bucket,
			object,
		)

		collector.RegisterResult(evalResults)
		collector.RegisterResult(evalResults)
		collector.Close()

		mockStorage.AssertExpectations(t)
	})
}
