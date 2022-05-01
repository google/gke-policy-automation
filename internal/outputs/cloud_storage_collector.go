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
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/gke-policy-automation/internal/policy"
	"google.golang.org/api/option"
)

type CloudStorageResultCollector struct {
	ctx               context.Context
	bucket            *storage.BucketHandle
	objectName        string
	validationResults ValidationResults
}

func BuildCloudStorageResultCollector(ctx context.Context, credentialsFile string, bucketName string, objectName string) (ValidationResultCollector, error) {

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, err
	}

	bucket := client.Bucket(bucketName)
	if _, err := bucket.Attrs(ctx); err != nil {
		return nil, err
	}

	return &CloudStorageResultCollector{
		ctx:        ctx,
		bucket:     bucket,
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

	w := p.bucket.Object(p.objectName).NewWriter(p.ctx)
	if _, err := w.Write(res); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
