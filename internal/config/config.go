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

// Package config implements application configuration related features
package config

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/google/gke-policy-automation/internal/log"
	"gopkg.in/yaml.v3"
)

var (
	DefaultK8SApiVersions = []string{"v1", "autoscaling/v1"}
)

const (
	DefaultGitRepository = "https://github.com/google/gke-policy-automation"
	DefaultGitBranch     = "main"
	DefaultGitPolicyDir  = "gke-policies-v2"
	DefaultK8SClientQPS  = 50
)

type ReadFileFn func(string) ([]byte, error)

type Config struct {
	SilentMode       bool                   `yaml:"silent"`
	JSONOutput       bool                   `yaml:"jsonOutput"`
	DumpFile         string                 `yaml:"dumpFile"`
	CredentialsFile  string                 `yaml:"credentialsFile"`
	Clusters         []ConfigCluster        `yaml:"clusters"`
	Policies         []ConfigPolicy         `yaml:"policies"`
	Inputs           ConfigInput            `yaml:"inputs"`
	Outputs          []ConfigOutput         `yaml:"outputs"`
	ClusterDiscovery ClusterDiscovery       `yaml:"clusterDiscovery"`
	PolicyExclusions ConfigPolicyExclusions `yaml:"policyExclusions"`
	Metrics          []ConfigMetric         `yaml:"metrics"`
	K8SApiConfig     K8SApiConfig           `yaml:"kubernetesAPIClient"`
}

type ConfigPolicy struct {
	LocalDirectory string `yaml:"local"`
	GitRepository  string `yaml:"repository"`
	GitBranch      string `yaml:"branch"`
	GitDirectory   string `yaml:"directory"`
}

type ConfigCluster struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Project  string `yaml:"project"`
	Location string `yaml:"location"`
}

type ConfigInput struct {
	GKEApi        *GKEApiInput     `yaml:"gkeAPI"`
	GKELocalInput *GKELocalInput   `yaml:"gkeLocal"`
	K8sAPI        *K8SAPIInput     `yaml:"k8sAPI"`
	MetricsAPI    *MetricsAPIInput `yaml:"metricsAPI"`
	Rest          *RestInput       `yaml:"rest"`
}

type GKEApiInput struct {
	Enabled bool `yaml:"enabled"`
}

type GKELocalInput struct {
	Enabled  bool   `yaml:"enabled"`
	DumpFile string `yaml:"file"`
}

type K8SAPIInput struct {
	Enabled     bool     `yaml:"enabled"`
	APIVersions []string `yaml:"resourceAPIVersions"`
	MaxQPS      int      `yaml:"clientMaxQPS"`
}

type MetricsAPIInput struct {
	Enabled   bool           `yaml:"enabled"`
	ProjectID string         `yaml:"project"`
	Address   string         `yaml:"address"`
	Username  string         `yaml:"username"`
	Password  string         `yaml:"password"`
	Metrics   []ConfigMetric `yaml:"metrics"`
}
type RestInput struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
}

type ConfigOutput struct {
	FileName              string                      `yaml:"file"`
	PubSub                PubSubOutput                `yaml:"pubsub"`
	CloudStorage          CloudStorageOutput          `yaml:"cloudStorage"`
	SecurityCommandCenter SecurityCommandCenterOutput `yaml:"securityCommandCenter"`
}

type ConfigMetric struct {
	MetricName string `yaml:"name"`
	Query      string `yaml:"query"`
}

type PubSubOutput struct {
	Project string `yaml:"project"`
	Topic   string `yaml:"topic"`
}

type CloudStorageOutput struct {
	Bucket         string `yaml:"bucket"`
	Path           string `yaml:"path"`
	SkipDatePrefix bool   `yaml:"skipDatePrefix"`
}

type SecurityCommandCenterOutput struct {
	OrganizationNumber string `yaml:"organization"`
	ProvisionSource    bool   `yaml:"provisionSource"`
}

type ClusterDiscovery struct {
	Enabled      bool     `yaml:"enabled"`
	Organization string   `yaml:"organization"`
	Folders      []string `yaml:"folders"`
	Projects     []string `yaml:"projects"`
}

type ConfigPolicyExclusions struct {
	Policies     []string `yaml:"policies"`
	PolicyGroups []string `yaml:"policyGroups"`
}

type K8SApiConfig struct {
	Enabled        bool     `yaml:"enabled"`
	APIVersions    []string `yaml:"resourceAPIVersions"`
	MaxQPS         int      `yaml:"clientMaxQPS"`
	TimeoutSeconds int      `yaml:"clientTimeoutSeconds"`
}

func ReadConfig(path string, readFn ReadFileFn) (*Config, error) {
	data, err := readFn(path)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	config := &Config{}
	if err := decoder.Decode(&config); err != nil && err != io.EOF {
		return nil, err
	}
	return config, nil
}

func ValidateClusterDumpConfig(config Config) error {
	var errors = make([]error, 0)
	errors = append(errors, validateClustersConfig(config)...)
	if len(errors) > 0 {
		for _, err := range errors {
			log.Warnf("configuration validation error: %s", err)
		}
		return errors[0]
	}
	return nil
}

func ValidateClusterCheckConfig(config Config) error {
	var errors = make([]error, 0)
	errors = append(errors, validateClustersConfig(config)...)
	errors = append(errors, validatePolicySourceConfig(config.Policies)...)
	errors = append(errors, validateOutputConfig(config.Outputs)...)
	if config.Inputs.GKEApi == nil && config.Inputs.GKELocalInput == nil {
		errors = append(errors, fmt.Errorf("either gkeAPI input or gkeLocalInput has to be declared"))
	}
	if config.Inputs.GKEApi != nil && !config.Inputs.GKEApi.Enabled {
		if config.Inputs.GKELocalInput == nil || !config.Inputs.GKELocalInput.Enabled {
			errors = append(errors, fmt.Errorf("either gkeAPI input or gkeLocalInput has to be enabled"))
		}
	}
	if config.Inputs.GKELocalInput != nil && !config.Inputs.GKELocalInput.Enabled {
		if config.Inputs.GKEApi == nil || !config.Inputs.GKEApi.Enabled {
			errors = append(errors, fmt.Errorf("either gkeAPI input or gkeLocalInput has to be enabled"))
		}
	}
	if len(errors) > 0 {
		for _, err := range errors {
			log.Warnf("configuration validation error: %s", err)
		}
		return errors[0]
	}
	return nil
}

func ValidatePolicyCheckConfig(config Config) error {
	errors := validatePolicySourceConfig(config.Policies)
	if len(errors) > 0 {
		for _, err := range errors {
			log.Warnf("configuration validation error: %s", err)
		}
		return errors[0]
	}
	return nil
}

func ValidateGeneratePolicyDocsConfig(config Config) error {
	var errors = make([]error, 0)
	errors = append(errors, validatePolicySourceConfig(config.Policies)...)
	if len(config.Outputs) != 1 {
		errors = append(errors, fmt.Errorf("specify output file"))
	}
	if len(errors) > 0 {
		for _, err := range errors {
			log.Warnf("configuration validation error: %s", err)
		}
		return errors[0]
	}
	return nil
}

func ValidateScalabilityCheckConfig(config Config) error {
	var errors = make([]error, 0)
	errors = append(errors, validateClustersConfig(config)...)
	errors = append(errors, validatePolicySourceConfig(config.Policies)...)
	errors = append(errors, validateOutputConfig(config.Outputs)...)
	if config.Inputs.MetricsAPI == nil || !config.Inputs.MetricsAPI.Enabled {
		errors = append(errors, fmt.Errorf("metricsAPI input has to be enabled"))
	}
	if config.Inputs.GKEApi == nil || !config.Inputs.GKEApi.Enabled {
		errors = append(errors, fmt.Errorf("gkeAPI input has to be enabled"))
	}
	if len(errors) > 0 {
		for _, err := range errors {
			log.Warnf("configuration validation error: %s", err)
		}
		return errors[0]
	}
	if config.Inputs.MetricsAPI.Enabled {
		if (config.Inputs.MetricsAPI.Username != "" && config.Inputs.MetricsAPI.Password == "") ||
			(config.Inputs.MetricsAPI.Password != "" && config.Inputs.MetricsAPI.Username == "") {
			return fmt.Errorf("can't set username without password or password without the username")
		}
		if config.Inputs.MetricsAPI.ProjectID != "" && config.Inputs.MetricsAPI.Address != "" {
			return fmt.Errorf("projectID should be not set when custom Prometheus address is specified")
		}
	}
	return nil
}

func validateClustersConfig(config Config) []error {
	if config.ClusterDiscovery.Enabled {
		discovery := config.ClusterDiscovery
		if config.DumpFile != "" {
			return []error{fmt.Errorf("cluster discovery is enabled along with a dump file")}
		}
		if len(config.Clusters) > 0 {
			return []error{fmt.Errorf("cluster discovery is enabled along with a defined cluster list")}
		}
		if len(discovery.Folders) < 1 && len(discovery.Projects) < 1 && discovery.Organization == "" {
			return []error{fmt.Errorf("cluster discovery is enabled but none of organization, folder list or project list are defined")}
		}
	} else {
		if config.DumpFile == "" && len(config.Clusters) < 1 {
			return []error{fmt.Errorf("cluster discovery is disabled and there are no clusters defined")}
		}
		var errors = make([]error, 0)
		for i, cluster := range config.Clusters {
			if cluster.ID == "" {
				if cluster.Name == "" {
					errors = append(errors, fmt.Errorf("cluster [%v]: name is not set", i))
				}
				if cluster.Location == "" {
					errors = append(errors, fmt.Errorf("cluster [%v]: location is not set", i))
				}
				if cluster.Project == "" {
					errors = append(errors, fmt.Errorf("cluster [%v]: project is not set", i))
				}
			} else {
				if cluster.Name != "" || cluster.Location != "" || cluster.Project != "" {
					errors = append(errors, fmt.Errorf("cluster [%v]: ID is set along with name or location or project", i))
				}
			}
		}
		return errors
	}
	return nil
}

func validatePolicySourceConfig(policies []ConfigPolicy) []error {
	if len(policies) < 1 {
		return []error{fmt.Errorf("there are no policy sources defined")}
	}
	var errors = make([]error, 0)
	for i, policy := range policies {
		if policy.LocalDirectory == "" {
			if policy.GitRepository == "" {
				errors = append(errors, fmt.Errorf("policy source [%v]: repository URL is not set", i))
			}
			if policy.GitBranch == "" {
				errors = append(errors, fmt.Errorf("policy source [%v]: repository branch is not set", i))
			}
			if policy.GitDirectory == "" {
				errors = append(errors, fmt.Errorf("policy source [%v]: repository directory is not set", i))
			}
		} else {
			if policy.GitRepository != "" || policy.GitBranch != "" || policy.GitDirectory != "" {
				errors = append(errors, fmt.Errorf("policy source [%v]: local directory is set along with GIT parameters", i))
			}
		}
	}
	return errors
}

func validateOutputConfig(outputs []ConfigOutput) []error {
	var errors = make([]error, 0)
	for _, output := range outputs {
		if output.FileName != "" && !strings.HasSuffix(output.FileName, ".json") {
			errors = append(errors, fmt.Errorf("invalid output - filename should end with .json"))
		}
		if output.CloudStorage.Bucket == "" && output.CloudStorage.Path != "" {
			errors = append(errors, fmt.Errorf("invalid output - bucket empty for path: %s", output.CloudStorage.Path))
		}
		if output.CloudStorage.Bucket != "" && output.CloudStorage.Path == "" {
			errors = append(errors, fmt.Errorf("invalid output - path empty for bucket: %s", output.CloudStorage.Bucket))
		}
		errors = append(errors, validatePubSubConfig(output.PubSub)...)
	}
	return errors
}

func validatePubSubConfig(pubsub PubSubOutput) []error {
	var errors = make([]error, 0)
	if pubsub.Project != "" && pubsub.Topic == "" {
		errors = append(errors, fmt.Errorf("PubSub Topic is not set for the project [%s]", pubsub.Project))
	}
	if pubsub.Topic != "" && pubsub.Project == "" {
		errors = append(errors, fmt.Errorf("PubSub Project name is not set for the topic [%s]", pubsub.Topic))
	}
	return errors
}

func SetCheckConfigDefaults(config *Config) {
	SetPolicyConfigDefaults(config)
	if config.Inputs.GKEApi == nil {
		log.Debugf("Configuring GKEApi input defaults")
		config.Inputs.GKEApi = &GKEApiInput{
			Enabled: true,
		}
	}
}

func SetScalabilityConfigDefaults(config *Config) {
	SetPolicyConfigDefaults(config)
	if config.Inputs.MetricsAPI == nil {
		log.Debugf("configuring MetricsApi input defaults")
		config.Inputs.MetricsAPI = &MetricsAPIInput{
			Enabled: true,
			Metrics: getScalabilityMetricsDefaults(),
		}
	} else {
		config.Inputs.MetricsAPI.Metrics = append(config.Inputs.MetricsAPI.Metrics, getScalabilityMetricsDefaults()...)
	}
	if config.Inputs.GKEApi == nil {
		log.Debugf("configuring GKEApi input defaults")
		config.Inputs.GKEApi = &GKEApiInput{
			Enabled: true,
		}
	}
	if config.Inputs.K8sAPI != nil {
		log.Debugf("Configuring K8SApiConfig input defaults")
		if config.Inputs.K8sAPI.MaxQPS == 0 {
			config.Inputs.K8sAPI.MaxQPS = DefaultK8SClientQPS
		}
		if len(config.Inputs.K8sAPI.APIVersions) == 0 {
			config.Inputs.K8sAPI.APIVersions = DefaultK8SApiVersions
		}
	}
}

func SetPolicyConfigDefaults(config *Config) {
	if len(config.Policies) < 1 {
		log.Debugf("no policies defined, using default GIT policy source: repo %s, branch %s, directory %s",
			DefaultGitRepository, DefaultGitBranch, DefaultGitPolicyDir)
		config.Policies = append(config.Policies, ConfigPolicy{
			GitRepository: DefaultGitRepository,
			GitBranch:     DefaultGitBranch,
			GitDirectory:  DefaultGitPolicyDir})
	}
}
