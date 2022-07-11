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
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/outputs/scc"
)

func (p *PolicyAutomationApp) ConfigureSCC() error {
	p.out.Printf("suck\n")
	cli, err := scc.NewSecurityCommandCenterClient(p.ctx, "153963171798")
	if err != nil {
		return err
	}
	log.Infof("Searching for source")
	id, err := cli.FindSource()
	if err != nil {
		return err
	}
	if id == nil {
		log.Infof("Creating source: %v", id)
		*id, err = cli.CreateSource()
		if err != nil {
			return err
		}
	}

	log.Infof("Using source %v", *id)

	_, err = cli.UpsertFinding(&scc.Finding{
		SourceName:   *id,
		ResourceName: "//container.googleapis.com/projects/gke-policy-demo/zones/europe-central2/clusters/cluster-waw",
		Category:     "GKE_POLICY_AUTOMATION_TEST",
	})
	if err != nil {
		return err
	}
	return nil
}
