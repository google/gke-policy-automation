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

package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

type CollectorPubSubClient struct {
	ctx    context.Context
	client *pubsub.Client
}

func NewPubSubClient(ctx context.Context, project string) (*CollectorPubSubClient, error) {
	client, err := pubsub.NewClient(ctx, project)

	if err != nil {
		return nil, err
	}

	return &CollectorPubSubClient{
		ctx:    ctx,
		client: client,
	}, nil
}

func NewPubSubClientWithCredentialsFile(ctx context.Context, project string, credentialsFile string) (*CollectorPubSubClient, error) {
	client, err := pubsub.NewClient(ctx, project, option.WithCredentialsFile(credentialsFile))

	if err != nil {
		return nil, err
	}

	return &CollectorPubSubClient{
		ctx:    ctx,
		client: client,
	}, nil
}

func (c *CollectorPubSubClient) Publish(topicName string, message []byte) (string, error) {
	topic := c.client.Topic(topicName)
	pubResult := topic.Publish(c.ctx, &pubsub.Message{
		Data: message,
	})
	return pubResult.Get(c.ctx)
}

func (c *CollectorPubSubClient) Close() error {
	return c.client.Close()
}
