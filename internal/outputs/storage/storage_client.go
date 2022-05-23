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

package storage

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/google/gke-policy-automation/internal/version"
	"google.golang.org/api/option"
)

type CloudStorageClient struct {
	ctx    context.Context
	client *storage.Client
}

func NewCloudStorageClient(ctx context.Context) (*CloudStorageClient, error) {
	return newCloudStorageClient(ctx)
}

func NewCloudStorageClientWithCredentialsFile(ctx context.Context, credentialsFile string) (*CloudStorageClient, error) {
	return newCloudStorageClient(ctx, option.WithCredentialsFile(credentialsFile))
}

func newCloudStorageClient(ctx context.Context, opts ...option.ClientOption) (*CloudStorageClient, error) {
	opts = append(opts, option.WithUserAgent(version.UserAgent))
	client, err := storage.NewClient(ctx, opts...)

	if err != nil {
		return nil, err
	}

	return &CloudStorageClient{
		ctx:    ctx,
		client: client,
	}, nil
}

func (c *CloudStorageClient) BucketExists(bucketName string) bool {
	_, err := c.client.Bucket(bucketName).Attrs(c.ctx)
	return err == nil
}

func (c *CloudStorageClient) Write(bucketName, objectName string, content []byte) error {
	w := c.client.Bucket(bucketName).Object(objectName).NewWriter(c.ctx)

	if _, err := w.Write(content); err != nil {
		w.Close()
		return err
	}

	return w.Close()
}

func (c *CloudStorageClient) Close() error {
	return c.client.Close()
}
