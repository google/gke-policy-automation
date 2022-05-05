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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/gke-policy-automation/internal/policy"
)

type CloudStorageResultCollector struct {
	ctx               context.Context
	client            StorageClient
	bucketName        string
	objectName        string
	validationResults ValidationResults
}

func NewCloudStorageResultCollector(ctx context.Context, client StorageClient, bucketName string, objectName string) (ValidationResultCollector, error) {

	if !client.BucketExists(bucketName) {
		return nil, fmt.Errorf("bucket does not exist %s", bucketName)
	}

	return &CloudStorageResultCollector{
		ctx:        ctx,
		client:     client,
		bucketName: bucketName,
		objectName: objectName,
	}, nil
}

func (p *CloudStorageResultCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {
	for _, r := range results {
		p.validationResults.ClusterValidationResults = append(p.validationResults.ClusterValidationResults, MapClusterToJson(r))
	}

	return nil
}

func (p *CloudStorageResultCollector) Close() error {
	p.validationResults.ValidationDate = time.Now()

	res, err := json.Marshal(p.validationResults)
	if err != nil {
		return err
	}

	if err := p.client.Write(p.bucketName, p.objectName, res); err != nil {
		return err
	}

	return p.client.Close()
}
