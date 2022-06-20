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
	"time"

	"github.com/google/gke-policy-automation/internal/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/retry"
)

type cred struct {
	googleDefaultTokenSource func(ctx context.Context, scope ...string) (oauth2.TokenSource, error)
	k8sStartingConfig        func() (*clientcmdapi.Config, error)
}

func newCred() *cred {
	return &cred{
		googleDefaultTokenSource: google.DefaultTokenSource,
		k8sStartingConfig:        k8sStartingConfig,
	}
}

type Creds struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Status     Status `json:"status"`
}

type Status struct {
	ExpirationTimestamp time.Time `json:"expirationTimestamp"`
	Token               string    `json:"token"`
}

// getClusterToken returns the token needed to authentication to the k8s cluster
func getClusterToken(ctx context.Context) (string, error) {
	var token string
	var err error
	cred := newCred()

	if token, err = cred.defaultAccessToken(ctx); err != nil {
		log.Debugf("unable to retrieve default access token: %s", err)
		return "", err
	}

	return token, nil
}

// defaultAccessToken retrieves the access token with the application default credentials
func (c *cred) defaultAccessToken(ctx context.Context) (string, error) {
	var tok *oauth2.Token
	var defaultScopes = []string{
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/userinfo.email"}

	err := retry.OnError(retry.DefaultBackoff, func(err error) bool { return true }, func() error {
		ts, err := c.googleDefaultTokenSource(ctx, defaultScopes...)
		if err != nil {
			log.Debugf("cannot construct google default token source: %s", err)
			return err
		}

		tok, err = ts.Token()
		if err != nil {
			log.Debugf("cannot retrieve default token from google default token source: %s", err)
			return err
		}

		return nil
	})
	if err != nil {
		log.Debugf("getting google default token failed after multiple retries: %s", err)
		return "", err
	}

	return tok.AccessToken, nil
}

func k8sStartingConfig() (*clientcmdapi.Config, error) {
	po := clientcmd.NewDefaultPathOptions()
	return po.GetStartingConfig()
}
