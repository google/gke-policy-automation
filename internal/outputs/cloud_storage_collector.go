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
	"fmt"

	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/policy"
)

type StorageClient interface {
	BucketExists(bucketName string) bool
	Write(bucketName, objectName string, content []byte) error
	Close() error
}

type reportMapper interface {
	AddResults(results []*policy.PolicyEvaluationResult)
	GetJsonReport() ([]byte, error)
}

type cloudStorageResultCollector struct {
	client       StorageClient
	bucketName   string
	objectName   string
	reportMapper reportMapper
}

func NewCloudStorageResultCollector(client StorageClient, bucketName string, objectName string) (ValidationResultCollector, error) {
	if !client.BucketExists(bucketName) {
		return nil, fmt.Errorf("bucket does not exist: %s", bucketName)
	}
	return &cloudStorageResultCollector{
		client:       client,
		bucketName:   bucketName,
		objectName:   objectName,
		reportMapper: NewValidationReportMapper(),
	}, nil
}

func (p *cloudStorageResultCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {
	p.reportMapper.AddResults(results)
	return nil
}

func (p *cloudStorageResultCollector) Close() error {
	reportData, err := p.reportMapper.GetJsonReport()
	if err != nil {
		return err
	}
	if err := p.client.Write(p.bucketName, p.objectName, reportData); err != nil {
		return err
	}
	log.Infof("Validation results stored in [%s] object in [%s] Cloud Storage bucket", p.objectName, p.bucketName)
	return p.client.Close()
}

func (p *cloudStorageResultCollector) Name() string {
	return "Cloud Storage bucket " + p.bucketName
}
