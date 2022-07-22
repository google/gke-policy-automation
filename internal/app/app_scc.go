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
	"errors"

	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/outputs/scc"
)

func (p *PolicyAutomationApp) ConfigureSCC(orgNumber string) error {
	if orgNumber == "" {
		return errors.New("organization number is not set")
	}
	cli, err := scc.NewSecurityCommandCenterClient(p.ctx, orgNumber)
	if err != nil {
		return err
	}
	p.out.ColorPrintf("\u2139 [light_gray][bold]Searching for GKE Policy Automation in SCC organization... [%s]\n", orgNumber)
	log.Infof("Searching for GKE Policy Automation in SCC organization %s", orgNumber)
	id, err := cli.FindSource()
	if err != nil {
		p.out.ErrorPrint("could not fetch SCC sources", err)
		return err
	}
	if id != nil {
		p.out.ColorPrintf("\u2139 [light_gray][bold]Found GKE Policy Automation in SCC... [%s]\n", *id)
		log.Infof("Found GKE Policy Automation in SCC: %s", *id)
		return nil
	}
	p.out.ColorPrintf("\u2139 [light_gray][bold]GKE Policy Automation was not found in SCC, creating it...\n")
	log.Info("Creating GKE Policy Automation in SCC")
	*id, err = cli.CreateSource()
	if err != nil {
		p.out.ErrorPrint("could not create GKE Policy Automation source in SCC", err)
		return err
	}
	p.out.ColorPrintf("\u2139 [light_gray][bold]Created GKE Policy Automation in SCC... [%s]\n", *id)
	log.Infof("Created GKE Policy Automation in SCC: %s", *id)
	return nil
}
