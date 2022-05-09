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
	"time"

	"github.com/google/gke-policy-automation/internal/policy"
)

type StorageClient interface {
	BucketExists(bucketName string) bool
	Write(bucketName, objectName string, content []byte) error
	Close() error
}

type Mapper func(evaluationResult []*policy.PolicyEvaluationResult, time time.Time) ([]byte, error)

type CloudStorageResultCollector struct {
	client            StorageClient
	mapper            Mapper
	bucketName        string
	objectName        string
	evaluationResults []*policy.PolicyEvaluationResult
}

func NewCloudStorageResultCollector(client StorageClient, mapper Mapper, bucketName string, objectName string) (*CloudStorageResultCollector, error) {

	if !client.BucketExists(bucketName) {
		return nil, fmt.Errorf("bucket does not exist: %s", bucketName)
	}

	return &CloudStorageResultCollector{
		client:     client,
		mapper:     mapper,
		bucketName: bucketName,
		objectName: objectName,
	}, nil
}

func (p *CloudStorageResultCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {
	p.evaluationResults = append(p.evaluationResults, results...)
	return nil
}

func (p *CloudStorageResultCollector) Close() error {

	res, err := p.mapper(p.evaluationResults, time.Now())
	if err != nil {
		return err
	}

	if err := p.client.Write(p.bucketName, p.objectName, res); err != nil {
		return err
	}

	return p.client.Close()
}
