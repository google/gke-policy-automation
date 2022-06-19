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
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

	cfg "github.com/google/gke-policy-automation/internal/config"
	"github.com/google/gke-policy-automation/internal/gke"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/outputs"
	pbc "github.com/google/gke-policy-automation/internal/outputs/pubsub"
	"github.com/google/gke-policy-automation/internal/outputs/storage"
	"github.com/google/gke-policy-automation/internal/policy"
	"github.com/google/gke-policy-automation/internal/version"
)

var errNoPolicies = errors.New("no policies to check against")

type PolicyAutomation interface {
	LoadCliConfig(cliConfig *CliConfig, validateFn cfg.ValidateConfig) error
	Close() error
	ClusterReview() error
	ClusterOfflineReview() error
	ClusterJSONData() error
	Version() error
	PolicyCheck() error
}

type PolicyAutomationApp struct {
	ctx        context.Context
	config     *cfg.Config
	out        *outputs.Output
	collectors []outputs.ValidationResultCollector
	gke        *gke.GKEClient
	gkeLocal   *gke.GKELocalClient
	discovery  gke.DiscoveryClient
}

func NewPolicyAutomationApp() PolicyAutomation {
	out := outputs.NewSilentOutput()
	return &PolicyAutomationApp{
		ctx:        context.Background(),
		config:     &cfg.Config{},
		out:        out,
		collectors: []outputs.ValidationResultCollector{outputs.NewConsoleResultCollector(out)},
	}
}

func (p *PolicyAutomationApp) LoadCliConfig(cliConfig *CliConfig, validateFn cfg.ValidateConfig) error {
	var config *cfg.Config
	var err error
	if cliConfig.ConfigFile != "" {
		if config, err = newConfigFromFile(cliConfig.ConfigFile); err != nil {
			return err
		}
	} else {
		config = newConfigFromCli(cliConfig)
	}
	cfg.SetConfigDefaults(config)
	if validateFn != nil {
		if err := validateFn(*config); err != nil {
			return err
		}
	}
	return p.LoadConfig(config)
}

func (p *PolicyAutomationApp) LoadConfig(config *cfg.Config) (err error) {
	p.config = config
	if !p.config.SilentMode {
		p.out = outputs.NewStdOutOutput()
		p.collectors = []outputs.ValidationResultCollector{outputs.NewConsoleResultCollector(p.out)}
	}
	if p.config.DumpFile != "" {
		// read file and instantiate the configuration
		p.gkeLocal, err = gke.NewGKELocalClient(p.ctx, p.config.DumpFile)
		return
	}
	if p.config.CredentialsFile != "" {
		p.gke, err = gke.NewClientWithCredentialsFile(p.ctx, p.config.K8SCheck, p.config.CredentialsFile)
	} else {
		p.gke, err = gke.NewClient(p.ctx, p.config.K8SCheck)
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

			var storagePath = o.CloudStorage.Path

			if !o.CloudStorage.SkipDatePrefix {
				storagePath = addDatetimePrefix(storagePath, time.Now())
			}

			storageCollector, err := outputs.NewCloudStorageResultCollector(storageClient, outputs.MapEvaluationResultsToJsonWithTime, o.CloudStorage.Bucket, storagePath)
			if err != nil {
				return err
			}
			p.collectors = append(p.collectors, storageCollector)
		}
		if len(o.PubSub.Topic) > 0 {
			var client outputs.PubSubClient
			if p.config.CredentialsFile != "" {
				client, err = pbc.NewPubSubClientWithCredentialsFile(p.ctx, o.PubSub.Project, p.config.CredentialsFile)
			} else {
				client, err = pbc.NewPubSubClient(p.ctx, o.PubSub.Project)
			}
			if err != nil {
				return err
			}
			p.collectors = append(p.collectors, outputs.NewPubSubResultCollector(client, o.PubSub.Project, o.PubSub.Topic))
		}
	}
	return
}

func (p *PolicyAutomationApp) Close() error {
	errors := make([]error, 0)
	if p.gke != nil {
		if err := p.gke.Close(); err != nil {
			log.Warnf("error when closing GKE client: %s", err)
			errors = append(errors, err)
		}
	}
	if p.discovery != nil {
		if err := p.discovery.Close(); err != nil {
			log.Warnf("error when closing discovery client: %s", err)
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func (p *PolicyAutomationApp) ClusterReview() error {
	log.Info("Cluster review starting")
	files, err := p.loadPolicyFiles()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		p.out.ColorPrintf("[yellow][bold]No policies to check against\n")
		log.Errorf("No policies to check against")
		return errNoPolicies
	}
	// create a PolicyAgent client instance
	pa := policy.NewPolicyAgent(p.ctx)
	p.out.ColorPrintf("[light_gray][bold]Parsing REGO policies...\n")
	log.Info("Parsing rego policies")
	// parsing policies before running checks
	if err := pa.WithFiles(files, p.config.PolicyExclusions); err != nil {
		p.out.ErrorPrint("could not parse policy files", err)
		log.Errorf("could not parse policy files: %s", err)
		return err
	}

	clusterIds, err := p.getClusters()
	if err != nil {
		p.out.ErrorPrint("could not get clusters", err)
		log.Errorf("could not get clusters: %s", err)
		return nil
	}
	evalResults := make([]*policy.PolicyEvaluationResult, 0)
	for _, clusterId := range clusterIds {
		log.Infof("Fetching GKE cluster %s", clusterId)
		p.out.ColorPrintf("[light_gray][bold]Fetching GKE cluster details... [%s]\n", clusterId)
		// getting Cluster information, from Cluster APIs and from Kubernetes APIs
		cluster, err := p.gke.GetCluster(clusterId, p.config.K8SCheck, cfg.APIVERSIONS)
		if err != nil {
			p.out.ErrorPrint("could not fetch the cluster details", err)
			log.Errorf("could not fetch cluster details: %s", err)
			return err
		}
		p.out.ColorPrintf("[light_gray][bold]Evaluating policies against GKE cluster... [%s]\n",
			cluster.Id)
		log.Infof("Evaluating policies against GKE cluster %s", clusterId)
		// evaluating Rego Policies
		evalResult, err := pa.Evaluate(cluster)
		if err != nil {
			p.out.ErrorPrint("failed to evaluate policies", err)
			log.Errorf("could not evaluate rego policies on cluster %s: %s", cluster.Id, err)
			return err
		}
		evalResult.ClusterName = clusterId
		evalResults = append(evalResults, evalResult)
	}

	for _, c := range p.collectors {
		log.Debugf("Collector %s registering the results", reflect.TypeOf(c).String())
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
		log.Debugf("Collector %s processing closed", reflect.TypeOf(c).String())
	}
	log.Info("Cluster review finished")
	return nil
}

func (p *PolicyAutomationApp) ClusterOfflineReview() error {
	// Loading the policies
	files, err := p.loadPolicyFiles()
	if err != nil {
		return err
	}
	// Creating the Policy Agent instance and check the policies
	pa := policy.NewPolicyAgent(p.ctx)
	p.out.ColorPrintf("[light_gray][bold]Parsing REGO policies...\n")
	log.Info("Parsing rego policies")
	if err := pa.WithFiles(files, p.config.PolicyExclusions); err != nil {
		p.out.ErrorPrint("could not parse policy files", err)
		log.Errorf("could not parse policy files: %s", err)
		return err
	}

	clusterName, err := p.gkeLocal.GetClusterName()
	if err != nil {
		p.out.ErrorPrint("could not create cluster path", err)
		log.Errorf("could not create cluster path: %s", err)
		return err
	}
	p.out.ColorPrintf("[light_gray][bold]Fetching GKE cluster details... [Name: %s]\n",
		clusterName)

	evalResults := make([]*policy.PolicyEvaluationResult, 0)
	cluster, err := p.gkeLocal.GetCluster()
	if err != nil {
		p.out.ErrorPrint("could not fetch the cluster details", err)
		log.Errorf("could not fetch cluster details: %s", err)
		return err
	}
	p.out.ColorPrintf("[light_gray][bold]Evaluating policies against GKE cluster... [ID: %s]\n",
		cluster.Id)
	evalResult, err := pa.Evaluate(cluster)
	if err != nil {
		p.out.ErrorPrint("failed to evalute policies", err)
		return err
	}
	evalResult.ClusterName = clusterName
	evalResults = append(evalResults, evalResult)

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
	clusterIds, err := p.getClusters()
	if err != nil {
		p.out.ErrorPrint("could not get clusters", err)
		log.Errorf("could not get clusters: %s", err)
	}
	for _, clusterId := range clusterIds {
		cluster, err := p.gke.GetCluster(clusterId, p.config.K8SCheck, cfg.APIVERSIONS)
		if err != nil {
			p.out.ErrorPrint("could not fetch the cluster details", err)
			log.Errorf("could not fetch cluster details: %s", err)
			return err
		}
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
	p.out.Printf("%s\n", version.Version)
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
	if err := pa.WithFiles(files, p.config.PolicyExclusions); err != nil {
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
		log.Infof("Discovering clusters for organization %s", p.config.ClusterDiscovery.Organization)
		p.out.ColorPrintf("[light_gray][bold]Discovering clusters in for an organization... [%s]\n", p.config.ClusterDiscovery.Organization)
		return p.discovery.GetClustersInOrg(p.config.ClusterDiscovery.Organization)
	}
	clusters := make([]string, 0)
	for _, folder := range p.config.ClusterDiscovery.Folders {
		log.Infof("Discovering clusters in a folder %s", folder)
		p.out.ColorPrintf("[light_gray][bold]Discovering clusters in a folder... [%s]\n", folder)
		results, err := p.discovery.GetClustersInFolder(folder)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, results...)
	}
	for _, project := range p.config.ClusterDiscovery.Projects {
		log.Infof("Discovering clusters in a project %s", project)
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

func addDatetimePrefix(value string, time time.Time) string {
	return fmt.Sprintf("%s_%s", time.Format("20060102_1504"), value)
}

func newConfigFromFile(path string) (*cfg.Config, error) {
	return cfg.ReadConfig(path, os.ReadFile)
}

func newConfigFromCli(cliConfig *CliConfig) *cfg.Config {
	config := &cfg.Config{}
	config.SilentMode = cliConfig.SilentMode
	config.K8SCheck = cliConfig.K8SCheck
	config.CredentialsFile = cliConfig.CredentialsFile
	config.DumpFile = cliConfig.DumpFile
	if cliConfig.ClusterName != "" || cliConfig.ClusterLocation != "" || cliConfig.ProjectName != "" {
		config.Clusters = []cfg.ConfigCluster{
			{
				Name:     cliConfig.ClusterName,
				Location: cliConfig.ClusterLocation,
				Project:  cliConfig.ProjectName,
			},
		}
	}
	config.Outputs = append(config.Outputs, cfg.ConfigOutput{
		FileName: cliConfig.OutputFile,
	})

	if cliConfig.LocalDirectory != "" || cliConfig.GitRepository != "" {
		config.Policies = append(config.Policies, cfg.ConfigPolicy{
			LocalDirectory: cliConfig.LocalDirectory,
			GitRepository:  cliConfig.GitRepository,
			GitBranch:      cliConfig.GitBranch,
			GitDirectory:   cliConfig.GitDirectory,
		})
	}
	return config
}

func getClusterName(c cfg.ConfigCluster) (string, error) {
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
