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
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/policy"
)

type PolicyAutomation interface {
	LoadCliConfig(cliConfig *CliConfig, validateFn ValidateConfig) error
	Close() error
	ClusterReview() error
	ClusterJSONData() error
	Version() error
	PolicyCheck() error
}

type PolicyAutomationApp struct {
	ctx    context.Context
	config *Config
	out    *Output
	gke    *gke.GKEClient
}

func NewPolicyAutomationApp() PolicyAutomation {
	return &PolicyAutomationApp{
		ctx:    context.Background(),
		config: &Config{},
		out:    NewSilentOutput(),
	}
}

func (p *PolicyAutomationApp) LoadCliConfig(cliConfig *CliConfig, validateFn ValidateConfig) error {
	var config *Config
	var err error
	if cliConfig.ConfigFile != "" {
		if config, err = newConfigFromFile(cliConfig.ConfigFile); err != nil {
			return err
		}
	} else {
		config = newConfigFromCli(cliConfig)
	}
	if validateFn != nil {
		if err := validateFn(*config); err != nil {
			return err
		}
	}
	return p.LoadConfig(config)
}

func (p *PolicyAutomationApp) LoadConfig(config *Config) (err error) {
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

func (p *PolicyAutomationApp) ClusterJSONData() error {
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
		p.out.ColorPrintf("[light_gray][bold]Printing GKE cluster JSON data... [%s]\n",
			cluster.Id)
		data, error := prettyJson(cluster)
		if error != nil {
			log.Errorf("could not print cluster data: %s", err)
			return err
		}
		p.out.Printf("%s\n", (data))
	}
	return nil
}

func (p *PolicyAutomationApp) Version() error {
	p.out.Printf("%s\n", Version)
	return nil
}

func (p *PolicyAutomationApp) PolicyCheck() error {
	files, err := p.loadPolicyFiles()
	if err != nil {
		p.out.ErrorPrint("loading policy files failed: ", err)
		log.Errorf("loading policy files failed: %s", err)
		return err
	}
	pa := policy.NewPolicyAgent(p.ctx)
	if err := pa.WithFiles(files); err != nil {
		p.out.ErrorPrint("could not parse policy files", err)
		log.Errorf("could not parse policy files: %s", err)
		return err
	}
	p.out.ColorPrintf("[bold][green] All policies validated correctly \n")
	log.Info("All policies validated correctly")
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

func newConfigFromFile(path string) (*Config, error) {
	return ReadConfig(path, os.ReadFile)
}

func newConfigFromCli(cliConfig *CliConfig) *Config {
	config := &Config{}
	config.SilentMode = cliConfig.SilentMode
	config.CredentialsFile = cliConfig.CredentialsFile
	config.Clusters = []ConfigCluster{
		{
			Name:     cliConfig.ClusterName,
			Location: cliConfig.ClusterLocation,
			Project:  cliConfig.ProjectName,
		},
	}
	if cliConfig.LocalDirectory == "" && cliConfig.GitRepository == "" {
		log.Debugf("using default git policy source: repo %s, branch %s, directory %s", DefaultGitRepository, DefaultGitBranch, DefaultGitPolicyDir)
		config.Policies = append(config.Policies, ConfigPolicy{
			GitRepository: DefaultGitRepository,
			GitBranch:     DefaultGitBranch,
			GitDirectory:  DefaultGitPolicyDir,
		})
	} else {
		config.Policies = append(config.Policies, ConfigPolicy{
			LocalDirectory: cliConfig.LocalDirectory,
			GitRepository:  cliConfig.GitRepository,
			GitBranch:      cliConfig.GitBranch,
			GitDirectory:   cliConfig.GitDirectory,
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

func prettyJson(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}
