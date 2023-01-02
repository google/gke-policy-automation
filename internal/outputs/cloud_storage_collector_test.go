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

	"github.com/google/gke-policy-automation/internal/policy"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) BucketExists(bucketName string) bool {
	args := m.Called(bucketName)
	return args.Bool(0)
}

func (m *MockStorage) Write(bucketName, objectName string, content []byte) error {
	args := m.Called(bucketName, objectName, content)
	return args.Error(0)
}

func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockMapper struct {
	mock.Mock
}

func (m *MockMapper) AddResults(results []*policy.PolicyEvaluationResult) {
	m.Called()
}

func (m *MockMapper) GetJSONReport() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func TestCloudStorageCollector(t *testing.T) {
	bucket := "mybucket"
	object := "myobject"
	mappedResult := []byte("result")

	policyResult := &policy.PolicyEvaluationResult{ClusterID: "cluster-one"}
	policyResult.Policies = append(policyResult.Policies, &policy.Policy{
		Name:             "policy-one",
		Title:            "title",
		Description:      "description",
		Group:            "group",
		Valid:            true,
		Violations:       []string{},
		ProcessingErrors: []error{},
	})

	t.Run("happy path", func(t *testing.T) {
		evalResults := []*policy.PolicyEvaluationResult{policyResult}

		mockStorage := &MockStorage{}
		mockStorage.On("BucketExists", bucket).Return(true)
		mockStorage.On("Write", bucket, object, mappedResult).Return(nil)
		mockStorage.On("Close").Return(nil)

		mockMapper := &MockMapper{}
		mockMapper.On("AddResults").Return()
		mockMapper.On("GetJSONReport").Return(mappedResult, nil)

		collector, err := NewCloudStorageResultCollector(mockStorage, bucket, object)
		if err != nil {
			t.Fatalf("error = %v; want nil", err)
		}
		csCollector := collector.(*cloudStorageResultCollector)
		csCollector.reportMapper = mockMapper

		collector.RegisterResult(evalResults)
		collector.Close()

		mockStorage.AssertExpectations(t)
		mockMapper.AssertExpectations(t)
	})

	t.Run("bucket does not exist", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockStorage.On("BucketExists", bucket).Return(false)

		var _, err = NewCloudStorageResultCollector(
			mockStorage,
			bucket,
			object,
		)

		if err == nil {
			t.Errorf("should have returned error")
		}
	})

	t.Run("collect multiple results", func(t *testing.T) {
		evalResults := []*policy.PolicyEvaluationResult{policyResult}

		mockStorage := &MockStorage{}
		mockStorage.On("BucketExists", bucket).Return(true)
		mockStorage.On("Write", bucket, object, mappedResult).Return(nil)
		mockStorage.On("Close").Return(nil)

		mockMapper := &MockMapper{}
		mockMapper.On("AddResults").Return()
		mockMapper.On("GetJSONReport").Return(mappedResult, nil)

		collector, err := NewCloudStorageResultCollector(mockStorage, bucket, object)
		if err != nil {
			t.Fatalf("error = %v; want nil", err)
		}
		csCollector := collector.(*cloudStorageResultCollector)
		csCollector.reportMapper = mockMapper

		collector.RegisterResult(evalResults)
		collector.RegisterResult(evalResults)
		collector.Close()

		mockStorage.AssertExpectations(t)
		mockMapper.AssertExpectations(t)
	})
}
