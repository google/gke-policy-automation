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
	"github.com/google/gke-policy-automation/internal/outputs"
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
	ctx       context.Context
	config    *Config
	out       *outputs.Output
	collector outputs.ValidationResultCollector
	gke       *gke.GKEClient
	discovery gke.DiscoveryClient
}

func NewPolicyAutomationApp() PolicyAutomation {
	out := outputs.NewSilentOutput()
	return &PolicyAutomationApp{
		ctx:       context.Background(),
		config:    &Config{},
		out:       out,
		collector: outputs.NewConsoleResultCollector(out)}
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
	setConfigDefaults(config)
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
		p.out = outputs.NewStdOutOutput()
		p.collector = outputs.NewConsoleResultCollector(p.out)
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
	if p.discovery != nil {
		return p.discovery.Close()
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

	clusterIds, err := p.getClusters()
	if err != nil {
		p.out.ErrorPrint("could not get clusters", err)
		log.Errorf("could not get clusters: %s", err)
	}
	evalResults := make([]*policy.PolicyEvaluationResult, 0)
	for _, clusterId := range clusterIds {
		p.out.ColorPrintf("[light_gray][bold]Fetching GKE cluster details... [%s]\n", clusterId)
		cluster, err := p.gke.GetCluster(clusterId)
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
		evalResult.ClusterName = clusterId
		evalResults = append(evalResults, evalResult)
	}
	err = p.collector.RegisterResult(evalResults)
	if err != nil {
		p.out.ErrorPrint("failed to register evaluation results", err)
		log.Errorf("could not register evaluation results: %s", err)
		return err
	}
	err = p.collector.Close()
	if err != nil {
		p.out.ErrorPrint("failed to close results registration", err)
		log.Errorf("could not finalize registering evaluation results: %s", err)
		return err
	}
	return nil
}

func (p *PolicyAutomationApp) ClusterJSONData() error {
	clusterIds, err := p.getClusters()
	if err != nil {
		p.out.ErrorPrint("could not get clusters", err)
		log.Errorf("could not get clusters: %s", err)
	}
	for _, clusterId := range clusterIds {
		p.out.ColorPrintf("[light_gray][bold]Fetching GKE cluster details... [%s]\n", clusterId)
		cluster, err := p.gke.GetCluster(clusterId)
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

//getClusters retrieves lists of a clusters for further processing
//from the sources that are defined in a configuration.
func (p *PolicyAutomationApp) getClusters() ([]string, error) {
	if p.config.ClusterDiscovery.Enabled {
		var dc gke.DiscoveryClient
		var err error
		if p.config.CredentialsFile != "" {
			log.Debugf("instantiating cluster discovery client with a credentials file")
			dc, err = gke.NewDiscoveryClientWithCredentialsFile(p.ctx, p.config.CredentialsFile)
		} else {
			log.Debugf("instantiating cluster discovery client")
			dc, err = gke.NewDiscoveryClient(p.ctx)
		}
		if err != nil {
			return nil, err
		}
		p.discovery = dc
		return p.discoverClusters()
	}
	clusters := make([]string, 0, len(p.config.Clusters))
	for _, configCluster := range p.config.Clusters {
		clusterName, err := getClusterName(configCluster)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, clusterName)
	}
	return clusters, nil
}

//discoverClusters discovers clusters according to the cluster discovery configuration.
func (p *PolicyAutomationApp) discoverClusters() ([]string, error) {
	if p.config.ClusterDiscovery.Organization != "" {
		log.Infof("discovering clusters for organization %s", p.config.ClusterDiscovery.Organization)
		p.out.ColorPrintf("[light_gray][bold]Discovering clusters in for an organization... [%s]\n", p.config.ClusterDiscovery.Organization)
		return p.discovery.GetClustersInOrg(p.config.ClusterDiscovery.Organization)
	}
	clusters := make([]string, 0)
	for _, folder := range p.config.ClusterDiscovery.Folders {
		log.Infof("discovering clusters in a folder %s", folder)
		p.out.ColorPrintf("[light_gray][bold]Discovering clusters in a folder... [%s]\n", folder)
		results, err := p.discovery.GetClustersInFolder(folder)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, results...)
	}
	for _, project := range p.config.ClusterDiscovery.Projects {
		log.Infof("discovering clusters in a project %s", project)
		p.out.ColorPrintf("[light_gray][bold]Discovering clusters in a project... [%s]\n", project)
		results, err := p.discovery.GetClustersInProject(project)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, results...)
	}
	log.Debugf("discovered %v clusters in projects and folders", len(clusters))
	return clusters, nil
}

func newConfigFromFile(path string) (*Config, error) {
	return ReadConfig(path, os.ReadFile)
}

func newConfigFromCli(cliConfig *CliConfig) *Config {
	config := &Config{}
	config.SilentMode = cliConfig.SilentMode
	config.CredentialsFile = cliConfig.CredentialsFile
	if cliConfig.ClusterName != "" || cliConfig.ClusterLocation != "" || cliConfig.ProjectName != "" {
		config.Clusters = []ConfigCluster{
			{
				Name:     cliConfig.ClusterName,
				Location: cliConfig.ClusterLocation,
				Project:  cliConfig.ProjectName,
			},
		}
	}
	if cliConfig.LocalDirectory != "" || cliConfig.GitRepository != "" {
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

func prettyJson(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}
