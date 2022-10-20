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

package clients

import (
	"context"
	"testing"

	"golang.org/x/oauth2"
)

type tsMock struct {
	tokenFn func() (*oauth2.Token, error)
}

func (m *tsMock) Token() (*oauth2.Token, error) {
	return m.tokenFn()
}

func TestNewGoogleTokenSourceWithCredentials(t *testing.T) {
	ts, err := NewGoogleTokenSourceWithCredentials(context.Background(), "../test-fixtures/test_credentials.json")
	if err != nil {
		t.Fatalf("error = %v; want nil", err)
	}
	_, ok := ts.(*googleTokenSource)
	if !ok {
		t.Fatalf("token source is not *googleTokenSource")
	}
}

func TestGetAuthToken(t *testing.T) {
	testAccessToken := "test-token"
	ts := &googleTokenSource{
		ctx: context.Background(),
		ts: &tsMock{
			tokenFn: func() (*oauth2.Token, error) {
				return &oauth2.Token{
					AccessToken: testAccessToken,
				}, nil
			},
		},
	}
	token, err := ts.GetAuthToken()
	if err != nil {
		t.Fatalf("error = %v; want nil", err)
	}
	if token != testAccessToken {
		t.Errorf("token = %v; want %v", token, testAccessToken)
	}
}
