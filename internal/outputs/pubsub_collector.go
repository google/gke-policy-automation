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
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/policy"
)

type PubSubClient interface {
	Publish(topicName string, message []byte) (string, error)
	Close() error
}

type pubSubResultCollector struct {
	client       PubSubClient
	project      string
	topic        string
	reportMapper *validationReportMapper
}

func NewPubSubResultCollector(client PubSubClient, project string, topic string) ValidationResultCollector {
	return &pubSubResultCollector{
		client:       client,
		project:      project,
		topic:        topic,
		reportMapper: NewValidationReportMapper(),
	}
}

func (p *pubSubResultCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {
	p.reportMapper.AddResults(results)
	return nil
}

func (p *pubSubResultCollector) Close() error {
	reportData, err := p.reportMapper.GetJsonReport()
	if err != nil {
		return err
	}
	id, err := p.client.Publish(p.topic, reportData)
	if err != nil {
		return err
	}
	log.Info("Validation results published to Pub/Sub topic [", p.topic, "] with message id [", id, "]")
	return p.client.Close()
}
