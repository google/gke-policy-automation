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
	"os"
	"time"

	cfg "github.com/google/gke-policy-automation/internal/config"
	"github.com/google/gke-policy-automation/internal/inputs"
	"github.com/google/gke-policy-automation/internal/inputs/clients"
	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/outputs"
	pbc "github.com/google/gke-policy-automation/internal/outputs/pubsub"
	"github.com/google/gke-policy-automation/internal/outputs/storage"
)

type validateConfig func(config cfg.Config) error
type setConfigDefaults func(config *cfg.Config)

func (p *PolicyAutomationApp) LoadCliConfig(cliConfig *CliConfig, defaultsFn setConfigDefaults, validateFn validateConfig) error {
	var config *cfg.Config
	var err error
	if cliConfig.ConfigFile != "" {
		config, err = newConfigFromFile(cliConfig.ConfigFile)
	} else {
		config = newConfigFromCli(cliConfig)
	}
	if err != nil {
		return err
	}
	if defaultsFn != nil {
		defaultsFn(config)
	}
	if validateFn != nil {
		if err := validateFn(*config); err != nil {
			return err
		}
	}
	return p.LoadConfig(config)
}

func (p *PolicyAutomationApp) LoadConfig(config *cfg.Config) error {
	p.config = config
	if p.config.JSONOutput {
		p.collectors = []outputs.ValidationResultCollector{outputs.NewConsoleJSONResultCollector(outputs.NewStdOutOutput())}
	} else {
		if !p.config.SilentMode {
			p.out = outputs.NewStdOutOutput()
			p.collectors = []outputs.ValidationResultCollector{outputs.NewConsoleResultCollector(p.out)}
			p.clusterDumpCollectors = append(p.clusterDumpCollectors, outputs.NewOutputClusterDumpCollector(p.out))
		}
	}
	if err := p.loadInputsConfig(config); err != nil {
		return err
	}
	if err := p.loadOutputsConfig(config); err != nil {
		return err
	}
	return nil
}

func (p *PolicyAutomationApp) loadInputsConfig(config *cfg.Config) error {
	if err := p.loadGKEApiInputConfig(config.Inputs.GKEApi, config.CredentialsFile); err != nil {
		return err
	}
	if err := p.loadGKELocalInputConfig(config.Inputs.GKELocalInput); err != nil {
		return err
	}
	if err := p.loadK8SApiInputConfig(config.Inputs.K8sAPI, config.CredentialsFile); err != nil {
		return err
	}
	if err := p.loadMetricsAPIInputConfig(config.Inputs.MetricsAPI, config.CredentialsFile); err != nil {
		return err
	}
	return nil
}

func (p *PolicyAutomationApp) loadGKEApiInputConfig(config *cfg.GKEApiInput, credentialsFile string) error {
	if config == nil || !config.Enabled {
		return nil
	}
	var input inputs.Input
	var err error
	if credentialsFile != "" {
		input, err = inputs.NewGKEApiInputWithCredentials(p.ctx, p.config.CredentialsFile)
	} else {
		input, err = inputs.NewGKEApiInput(p.ctx)
	}
	if err != nil {
		return err
	}
	p.inputs = append(p.inputs, input)
	return nil
}

func (p *PolicyAutomationApp) loadGKELocalInputConfig(config *cfg.GKELocalInput) error {
	if config != nil && config.Enabled {
		p.inputs = append(p.inputs, inputs.NewGKELocalInput(config.DumpFile))
	}
	return nil
}

func (p *PolicyAutomationApp) loadK8SApiInputConfig(config *cfg.K8SAPIInput, credentialsFile string) error {
	if config == nil || !config.Enabled {
		return nil
	}
	k8InputBuilder := inputs.NewK8sAPIInputBuilder(p.ctx, config.APIVersions).
		WithCredentialsFile(p.config.CredentialsFile)
	k8Input, err := k8InputBuilder.Build()
	if err != nil {
		return err
	}
	p.inputs = append(p.inputs, k8Input)
	return nil
}

func (p *PolicyAutomationApp) loadMetricsAPIInputConfig(config *cfg.MetricsAPIInput, credentialsFile string) error {
	if config == nil || !config.Enabled {
		return nil
	}
	var metricQueries []clients.MetricQuery
	for _, m := range config.Metrics {
		metricQueries = append(metricQueries, clients.MetricQuery{Name: m.MetricName, Query: m.Query})
	}

	metricInputBuilder := inputs.NewMetricsInputBuilder(p.ctx, metricQueries).
		WithCredentialsFile(p.config.CredentialsFile)

	metricInput, err := metricInputBuilder.Build()
	if err != nil {
		return err
	}
	p.inputs = append(p.inputs, metricInput)
	return nil
}

func (p *PolicyAutomationApp) loadOutputsConfig(config *cfg.Config) error {
	for _, out := range config.Outputs {
		if err := p.loadFileOutputConfig(out.FileName); err != nil {
			return nil
		}
		if err := p.loadCloudStorageOutputConfig(out.CloudStorage, config.CredentialsFile); err != nil {
			return nil
		}
		if err := p.loadPubSubOutputConfig(out.PubSub, config.CredentialsFile); err != nil {
			return nil
		}
		if err := p.loadSccOutputConfig(out.SecurityCommandCenter, config.CredentialsFile); err != nil {
			return nil
		}
	}
	return nil
}

func (p *PolicyAutomationApp) loadFileOutputConfig(fileName string) error {
	if fileName != "" {
		log.Infof("Loading File output")
		p.collectors = append(p.collectors, outputs.NewJSONResultToFileCollector(fileName))
		p.clusterDumpCollectors = append(p.clusterDumpCollectors, outputs.NewFileClusterDumpCollector(fileName))
		p.policyDocsFile = fileName
	}
	return nil
}

func (p *PolicyAutomationApp) loadCloudStorageOutputConfig(config cfg.CloudStorageOutput, credentialsFile string) error {
	if config.Bucket == "" || config.Path == "" {
		return nil
	}
	log.Infof("Loading Cloud Storage output")
	var client *storage.CloudStorageClient
	var err error
	if credentialsFile != "" {
		client, err = storage.NewCloudStorageClientWithCredentialsFile(p.ctx, credentialsFile)
	} else {
		client, err = storage.NewCloudStorageClient(p.ctx)
	}
	if err != nil {
		return err
	}

	storagePath := config.Path
	if !config.SkipDatePrefix {
		storagePath = addDateTimePrefix(storagePath, time.Now())
	}
	storageCollector, err := outputs.NewCloudStorageResultCollector(client, config.Bucket, storagePath)
	if err != nil {
		return err
	}
	p.collectors = append(p.collectors, storageCollector)
	return nil
}

func (p *PolicyAutomationApp) loadPubSubOutputConfig(config cfg.PubSubOutput, credentialsFile string) error {
	if config.Topic == "" {
		return nil
	}
	log.Infof("Loading PubSub output")
	var client outputs.PubSubClient
	var err error
	if credentialsFile != "" {
		client, err = pbc.NewPubSubClientWithCredentialsFile(p.ctx, config.Project, credentialsFile)
	} else {
		client, err = pbc.NewPubSubClient(p.ctx, config.Project)
	}
	if err != nil {
		return err
	}
	p.collectors = append(p.collectors, outputs.NewPubSubResultCollector(client, config.Project, config.Topic))
	return nil
}

func (p *PolicyAutomationApp) loadSccOutputConfig(config cfg.SecurityCommandCenterOutput, credsFile string) error {
	if config.OrganizationNumber == "" {
		return nil
	}
	log.Infof("Loading Security Command Center output")
	collector, err := outputs.NewSccCollector(p.ctx, config.OrganizationNumber, config.ProvisionSource, credsFile)
	if err != nil {
		return err
	}
	p.collectors = append(p.collectors, collector)
	return nil
}

func newConfigFromFile(path string) (*cfg.Config, error) {
	return cfg.ReadConfig(path, os.ReadFile)
}

func addDateTimePrefix(value string, time time.Time) string {
	return fmt.Sprintf("%s_%s", time.Format("20060102_1504"), value)
}

func newConfigFromCli(cliConfig *CliConfig) *cfg.Config {
	config := &cfg.Config{}
	config.SilentMode = cliConfig.SilentMode
	config.JSONOutput = cliConfig.JSONOutput
	config.CredentialsFile = cliConfig.CredentialsFile
	config.DumpFile = cliConfig.DumpFile
	if cliConfig.DiscoveryEnabled {
		config.ClusterDiscovery.Enabled = true
		if cliConfig.ProjectName != "" {
			config.ClusterDiscovery.Projects = []string{cliConfig.ProjectName}
		}
	} else {
		if cliConfig.ClusterName != "" || cliConfig.ClusterLocation != "" || cliConfig.ProjectName != "" {
			config.Clusters = []cfg.ConfigCluster{
				{
					Name:     cliConfig.ClusterName,
					Location: cliConfig.ClusterLocation,
					Project:  cliConfig.ProjectName,
				},
			}
		}
	}
	if cliConfig.OutputFile != "" {
		config.Outputs = append(config.Outputs, cfg.ConfigOutput{
			FileName: cliConfig.OutputFile,
		})
	}
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
