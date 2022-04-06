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
	"context"
	"fmt"
	"os"

	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/policy"
)

type PolicyAutomation interface {
	LoadCliConfig(cliConfig *CliConfig) error
	Close() error
	ClusterReview() error
	Version() error
}

type PolicyAutomationApp struct {
	ctx    context.Context
	config *ConfigNg
	out    *Output
	gke    *gke.GKEClient
}

func NewPolicyAutomationApp() PolicyAutomation {
	return &PolicyAutomationApp{
		ctx:    context.Background(),
		config: &ConfigNg{},
		out:    NewSilentOutput(),
	}
}

func (p *PolicyAutomationApp) LoadCliConfig(cliConfig *CliConfig) error {
	var config *ConfigNg
	var err error
	if cliConfig.ConfigFile != "" {
		if config, err = newConfigFromFile(cliConfig.ConfigFile); err != nil {
			return err
		}
	} else {
		config = newConfigFromCli(cliConfig)
	}
	return p.LoadConfig(config)
}

func (p *PolicyAutomationApp) LoadConfig(config *ConfigNg) (err error) {
	p.config = config
	if !p.config.SilentMode {
		p.out = NewStdOutOutput()
	}
	if p.config.CredentialsFile != "" {
		p.gke, err = gke.NewClientWithCredentialsFile(p.ctx, p.config.CredentialsFile)
	} else {
		p.gke, err = gke.NewClient(p.ctx)
	}
	return
}

func (p *PolicyAutomationApp) Close() error {
	if p.gke != nil {
		return p.gke.Close()
	}
	return nil
}

func (p *PolicyAutomationApp) ClusterReview() error {
	files, err := p.loadPolicyFiles()
	if err != nil {
		return err
	}
	pa := policy.NewPolicyAgent(p.ctx)
	p.out.ColorPrintf("[light_gray][bold]Parsing REGO policies...\n")
	log.Info("Parsing rego policies")
	if err := pa.WithFiles(files); err != nil {
		p.out.ErrorPrint("could not parse policy files", err)
		log.Errorf("could not parse policy files: %s", err)
		return err
	}

	evalResults := make([]*policy.PolicyEvaluationResult, 0)
	for _, cluster := range p.config.Clusters {
		clusterName, err := getClusterName(cluster)
		if err != nil {
			p.out.ErrorPrint("could not create cluster path", err)
			log.Errorf("could not create cluster path: %s", err)
			return err
		}
		p.out.ColorPrintf("[light_gray][bold]Fetching GKE cluster details... [projects/%s/locations/%s/clusters/%s]\n",
			cluster.Project,
			cluster.Location,
			cluster.Name)
		cluster, err := p.gke.GetCluster(clusterName)
		if err != nil {
			p.out.ErrorPrint("could not fetch the cluster details", err)
			log.Errorf("could not fetch cluster details: %s", err)
			return err
		}
		p.out.ColorPrintf("[light_gray][bold]Evaluating policies against GKE cluster... [%s]\n",
			cluster.Id)
		evalResult, err := pa.Evaluate(cluster)
		if err != nil {
			p.out.ErrorPrint("failed to evalute policies", err)
			log.Errorf("could not evaluate rego policies on cluster %s: %s", cluster.Id, err)
			return err
		}
		evalResult.ClusterName = clusterName
		evalResults = append(evalResults, evalResult)
	}
	p.printEvaluationResults(evalResults)
	return nil
}

func (p *PolicyAutomationApp) Version() error {
	p.out.Printf("%s\n", Version)
	return nil
}

func (p *PolicyAutomationApp) loadPolicyFiles() ([]*policy.PolicyFile, error) {
	policyFiles := make([]*policy.PolicyFile, 0)
	for _, policyConfig := range p.config.Policies {
		var policySrc policy.PolicySource
		if policyConfig.LocalDirectory != "" {
			policySrc = policy.NewLocalPolicySource(policyConfig.LocalDirectory)
		}
		if policyConfig.GitRepository != "" {
			policySrc = policy.NewGitPolicySource(policyConfig.GitRepository,
				policyConfig.GitBranch,
				policyConfig.GitDirectory)
		}
		p.out.ColorPrintf("[light_gray][bold]Reading policy files... [%s]\n", policySrc)
		log.Infof("Reading policy files from %s", policySrc)
		files, err := policySrc.GetPolicyFiles()
		if err != nil {
			p.out.ErrorPrint("could not read policy files", err)
			log.Errorf("could not read policy files: %s", err)
			return nil, err
		}
		policyFiles = append(policyFiles, files...)
	}
	return policyFiles, nil
}

func newConfigFromFile(path string) (*ConfigNg, error) {
	return ReadConfig(path, os.ReadFile)
}

func newConfigFromCli(cliConfig *CliConfig) *ConfigNg {
	config := &ConfigNg{}
	config.SilentMode = cliConfig.SilentMode
	config.CredentialsFile = cliConfig.CredentialsFile
	config.Clusters = []ConfigCluster{
		{
			Name:     cliConfig.ClusterName,
			Location: cliConfig.ClusterLocation,
			Project:  cliConfig.ProjectName,
		},
	}
	if cliConfig.LocalDirectory != "" {
		config.Policies = append(config.Policies, ConfigPolicy{LocalDirectory: cliConfig.LocalDirectory})
	}
	if cliConfig.GitRepository != "" {
		config.Policies = append(config.Policies, ConfigPolicy{
			GitRepository: cliConfig.GitRepository,
			GitBranch:     cliConfig.GitBranch,
			GitDirectory:  cliConfig.GitDirectory,
		})
	}
	return config
}

func getClusterName(c ConfigCluster) (string, error) {
	if c.ID != "" {
		return c.ID, nil
	}
	if c.Name != "" && c.Location != "" && c.Project != "" {
		return gke.GetClusterName(c.Project, c.Location, c.Name), nil
	}
	return "", fmt.Errorf("cluster mandatory parameters not set (project, name, location)")
}

func (p *PolicyAutomationApp) printEvaluationResults(results []*policy.PolicyEvaluationResult) {
	for _, result := range results {
		p.out.ColorPrintf("[yellow][bold]GKE Cluster [%s]:", result.ClusterName)
		for _, group := range result.Groups() {
			p.out.ColorPrintf("\n[light_gray][bold]Group %q:\n\n", group)
			for _, policy := range result.Valid[group] {
				p.out.ColorPrintf("[bold][green][\u2713] %s: [reset][green]%s\n", policy.Title, policy.Description)
			}
			for _, policy := range result.Violated[group] {
				p.out.ColorPrintf("[bold][red][x] %s: [reset][red]%s. [bold]Violations:[reset][red] %s\n", policy.Title, policy.Description, policy.Violations[0])
			}
		}
		p.out.ColorPrintf("\n[bold][green]GKE cluster [%s]: Policies: %d valid, %d violated, %d errored.\n",
			result.ClusterName,
			result.ValidCount(),
			result.ViolatedCount(),
			result.ErroredCount())
	}
}
