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

package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/gke-policy-automation/internal/log"
	"gopkg.in/yaml.v2"
)

var (
	DefaultK8SApiVersions = []string{"v1", "autoscaling/v1"}
)

const (
	DefaultGitRepository = "https://github.com/google/gke-policy-automation"
	DefaultGitBranch     = "main"
	DefaultGitPolicyDir  = "gke-policies"
	DefaultK8SClientQPS  = 50
)

type ReadFileFn func(string) ([]byte, error)
type ValidateConfig func(config Config) error

type Config struct {
	SilentMode       bool                   `yaml:"silent"`
	JsonOutput       bool                   `yaml:"jsonOutput"`
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
	GKEApi        GKEApiInput     `yaml:"gkeAPI"`
	GKELocalInput GKELocalInput   `yaml:"gkeLocal"`
	K8sApi        K8SApiInput     `yaml:"k8sAPI"`
	MetricsApi    MetricsApiInput `yaml:"metricsAPI"`
	Rest          RestInput       `yaml:"rest"`
}

type GKEApiInput struct {
	Enabled bool `yaml:"enabled"`
}

type GKELocalInput struct {
	Enabled  bool   `yaml:"enabled"`
	DumpFile string `yaml:"file"`
}

type K8SApiInput struct {
	Enabled     bool     `yaml:"enabled"`
	ApiVersions []string `yaml:"resourceAPIVersions"`
	MaxQPS      int      `yaml:"clientMaxQPS"`
}

type MetricsApiInput struct {
	Enabled   bool           `yaml:"enabled"`
	ProjectId string         `yaml:"project"`
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
	ApiVersions    []string `yaml:"resourceAPIVersions"`
	MaxQPS         int      `yaml:"clientMaxQPS"`
	TimeoutSeconds int      `yaml:"clientTimeoutSeconds"`
}

func ReadConfig(path string, readFn ReadFileFn) (*Config, error) {
	data, err := readFn(path)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func ValidateClusterJSONDataConfig(config Config) error {
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

func ValidateClusterOfflineReviewConfig(config Config) error {
	var errors = make([]error, 0)
	if config.DumpFile == "" {
		errors = append(errors, fmt.Errorf("cluster dump file is not set"))
	}
	for i, cluster := range config.Clusters {
		if cluster.ID == "" {
			if cluster.Name == "" {
				errors = append(errors, fmt.Errorf("cluster [%v]: name is not set", i))
			}
		}
	}
	errors = append(errors, validatePolicySourceConfig(config.Policies)...)
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
	if err := ValidateClusterCheckConfig(config); err != nil {
		return nil
	}
	if !config.K8SApiConfig.Enabled {
		return errors.New("kubernetes API client is disabled")
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

// setConfigDefaults checks passed config and sets default values if needed
func SetConfigDefaults(config *Config) {
	if len(config.Policies) < 1 {
		log.Debugf("no policies defined, using default GIT policy source: repo %s, branch %s, directory %s",
			DefaultGitRepository, DefaultGitBranch, DefaultGitPolicyDir)
		config.Policies = append(config.Policies, ConfigPolicy{
			GitRepository: DefaultGitRepository,
			GitBranch:     DefaultGitBranch,
			GitDirectory:  DefaultGitPolicyDir})
	}
	if config.K8SApiConfig.MaxQPS == 0 {
		config.K8SApiConfig.MaxQPS = DefaultK8SClientQPS
	}
	if len(config.K8SApiConfig.ApiVersions) == 0 {
		config.K8SApiConfig.ApiVersions = DefaultK8SApiVersions
	}
}
