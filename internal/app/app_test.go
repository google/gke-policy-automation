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

package app

import (
	"fmt"
	"testing"

	cfg "github.com/google/gke-policy-automation/internal/config"
	"github.com/google/gke-policy-automation/internal/outputs"
)

type DiscoveryClientMock struct {
	GetClustersInProjectFn func(name string) ([]string, error)
	GetClustersInFolderFn  func(number string) ([]string, error)
	GetClustersInOrgFn     func(number string) ([]string, error)
	CloseFn                func() error
}

func (m DiscoveryClientMock) GetClustersInProject(name string) ([]string, error) {
	return m.GetClustersInProjectFn(name)
}

func (m DiscoveryClientMock) GetClustersInFolder(number string) ([]string, error) {
	return m.GetClustersInFolderFn(number)
}

func (m DiscoveryClientMock) GetClustersInOrg(number string) ([]string, error) {
	return m.GetClustersInOrgFn(number)
}

func (m DiscoveryClientMock) Close() error {
	return m.CloseFn()
}

func TestNewPolicyAutomationApp(t *testing.T) {
	pa := NewPolicyAutomationApp()
	paApp, ok := pa.(*PolicyAutomationApp)
	if !ok {
		t.Fatalf("Result of NewPolicyAutomationApp is not *PolicyAutomationApp")
	}
	if paApp.ctx == nil {
		t.Fatalf("policyAutomationApp ctx is nil")
	}
	if paApp.out == nil {
		t.Fatalf("policyAutomationApp output is nil")
	}
	if len(paApp.collectors) == 0 {
		t.Fatalf("policyAutomationApp collector is nil")
	}
	if len(paApp.collectors) <= 0 {
		t.Fatalf("policyAutomationApp has no output collectors")
	}
}

func TestClusterReviewWithNoPolicies(t *testing.T) {
	pa := PolicyAutomationApp{
		out: outputs.NewSilentOutput(),
		config: &cfg.Config{
			Policies: []cfg.ConfigPolicy{},
		},
	}

	err := pa.Check()
	if err != errNoPolicies {
		t.Fatalf("need noPoliciesError but err = %s", err)
	}
}

func TestPolicyAutomationAppClose_negative(t *testing.T) {
	closeErr := fmt.Errorf("close error")
	pa := PolicyAutomationApp{
		discovery: DiscoveryClientMock{
			CloseFn: func() error {
				return closeErr
			},
		},
	}
	err := pa.Close()
	if err == nil {
		t.Fatalf("error is nil; want error")
	}
	if err != closeErr {
		t.Errorf("error is %v; want %v", err, closeErr)
	}
}

type MockDocumentation struct {
	content string
}

func (m *MockDocumentation) GenerateDocumentation() string {
	return m.content
}
