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

type PubSubMock struct {
	mock.Mock
}

func (m *PubSubMock) Publish(topicName string, message []byte) (string, error) {
	args := m.Called(topicName, message)
	return args.String(0), args.Error(1)
}

func (m *PubSubMock) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestPublishingToPubSub(t *testing.T) {
	policyResult := &policy.PolicyEvaluationResult{}
	policyResult.Policies = append(policyResult.Policies, &policy.Policy{
		Title:            "title",
		Description:      "description",
		Group:            "group",
		Valid:            true,
		Violations:       []string{},
		ProcessingErrors: []error{},
	})

	project := "my-project"
	topicName := "my-topic"
	mockMessageID := "xxxxxx"

	evalResults := make([]*policy.PolicyEvaluationResult, 0)
	evalResults = append(evalResults, policyResult)

	mockPubSub := &PubSubMock{}
	mockPubSub.On("Publish", topicName, mock.Anything).Return(mockMessageID, nil)
	mockPubSub.On("Close").Return(nil)

	var collector = NewPubSubResultCollector(mockPubSub, project, topicName)

	collector.RegisterResult(evalResults)
	collector.Close()

	mockPubSub.AssertExpectations(t)
}
