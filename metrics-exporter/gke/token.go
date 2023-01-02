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

package gke

import (
	"context"
	"os"

	"github.com/google/gke-policy-automation/metrics-exporter/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"k8s.io/client-go/util/retry"
)

type TokenSource interface {
	GetAuthToken() (string, error)
}

type googleTokenSource struct {
	ctx context.Context
	ts  oauth2.TokenSource
}

var defaultGoogleOAuthScopes = []string{"https://www.googleapis.com/auth/cloud-platform"}

func NewGoogleTokenSource(ctx context.Context) (TokenSource, error) {
	ts, err := google.DefaultTokenSource(ctx, defaultGoogleOAuthScopes...)
	if err != nil {
		return nil, err
	}
	return &googleTokenSource{
		ctx: ctx,
		ts:  ts,
	}, nil
}

func NewGoogleTokenSourceWithCredentials(ctx context.Context, credentialsFile string) (TokenSource, error) {
	credsB, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, err
	}
	creds, err := google.CredentialsFromJSONWithParams(ctx, credsB, google.CredentialsParams{Scopes: defaultGoogleOAuthScopes})
	if err != nil {
		return nil, err
	}
	return &googleTokenSource{
		ctx: ctx,
		ts:  creds.TokenSource,
	}, nil
}

func (s *googleTokenSource) GetAuthToken() (string, error) {
	var token *oauth2.Token
	err := retry.OnError(retry.DefaultBackoff, func(err error) bool { return true }, func() error {
		var err error
		token, err = s.ts.Token()
		if err != nil {
			log.Debugf("cannot construct google default token source: %s", err)
			return err
		}
		return nil
	})
	if err != nil {
		log.Debugf("getting google default token failed after multiple retries: %s", err)
		return "", err
	}
	return token.AccessToken, nil
}
