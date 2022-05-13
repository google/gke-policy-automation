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
	"reflect"

	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/outputs"
	"github.com/google/gke-policy-automation/internal/outputs/storage"
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
	ctx        context.Context
	config     *Config
	out        *outputs.Output
	collectors []outputs.ValidationResultCollector
	gke        *gke.GKEClient
}

func NewPolicyAutomationApp() PolicyAutomation {
	out := outputs.NewSilentOutput()
	return &PolicyAutomationApp{
		ctx:        context.Background(),
		config:     &Config{},
		out:        out,
		collectors: []outputs.ValidationResultCollector{outputs.NewConsoleResultCollector(out)},
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
		p.collectors = []outputs.ValidationResultCollector{outputs.NewConsoleResultCollector(p.out)}
	}
	if p.config.CredentialsFile != "" {
		p.gke, err = gke.NewClientWithCredentialsFile(p.ctx, p.config.CredentialsFile)
	} else {
		p.gke, err = gke.NewClient(p.ctx)
	}
	for _, o := range p.config.Outputs {
		if o.FileName != "" {
			p.collectors = append(p.collectors, outputs.NewJSONResultToFileCollector(o.FileName))
		}
		if o.CloudStorage.Bucket != "" && o.CloudStorage.Path != "" {

			var storageClient *storage.CloudStorageClient
			if p.config.CredentialsFile != "" {
				storageClient, err = storage.NewCloudStorageClientWithCredentialsFile(p.ctx, p.config.CredentialsFile)
			} else {
				storageClient, err = storage.NewCloudStorageClient(p.ctx)
			}
			storageCollector, err := outputs.NewCloudStorageResultCollector(storageClient, outputs.MapEvaluationResultsToJsonWithTime, o.CloudStorage.Bucket, o.CloudStorage.Path)
			if err != nil {
				return err
			}
			p.collectors = append(p.collectors, storageCollector)
		}
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

	for _, c := range p.collectors {
		err = c.RegisterResult(evalResults)

		if err != nil {
			p.out.ErrorPrint("failed to register evaluation results", err)
			log.Errorf("could not register evaluation results: %s", err)
			return err
		}
	}

	for _, c := range p.collectors {
		err = c.Close()
		if err != nil {
			p.out.ErrorPrint("failed to close results registration", err)
			log.Errorf("could not finalize registering evaluation results: %s", err)
			return err
		}
		log.Infof("Collector %s processing closed", reflect.TypeOf(c).Name())

	}
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
	config.Outputs = append(config.Outputs, ConfigOutput{
		FileName: cliConfig.OutputFile,
	})

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
