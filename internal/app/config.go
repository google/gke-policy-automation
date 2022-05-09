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

	"github.com/google/gke-policy-automation/internal/log"
	"gopkg.in/yaml.v2"
)

const (
	DefaultGitRepository = "https://github.com/google/gke-policy-automation"
	DefaultGitBranch     = "main"
	DefaultGitPolicyDir  = "gke-policies"
)

type ReadFileFn func(string) ([]byte, error)
type ValidateConfig func(config Config) error

type Config struct {
	SilentMode       bool             `yaml:"silent"`
	CredentialsFile  string           `yaml:"credentialsFile"`
	Clusters         []ConfigCluster  `yaml:"clusters"`
	Policies         []ConfigPolicy   `yaml:"policies"`
	ClusterDiscovery ClusterDiscovery `yaml:"clusterDiscovery"`
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

type ClusterDiscovery struct {
	Enabled      bool     `yaml:"enabled"`
	Organization string   `yaml:"organization"`
	Folders      []string `yaml:"folders"`
	Projects     []string `yaml:"projects"`
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

func ValidateClusterReviewConfig(config Config) error {
	var errors = make([]error, 0)
	errors = append(errors, validateClustersConfig(config)...)
	errors = append(errors, validatePolicySourceConfig(config.Policies)...)
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

func validateClustersConfig(config Config) []error {
	if config.ClusterDiscovery.Enabled {
		discovery := config.ClusterDiscovery
		if len(config.Clusters) > 0 {
			return []error{fmt.Errorf("cluster discovery is enabled along with a defined cluster list")}
		}
		if len(discovery.Folders) < 1 && len(discovery.Projects) < 1 && discovery.Organization == "" {
			return []error{fmt.Errorf("cluster discovery is enabled but none of organization, folder list or project list are defined")}
		}
	} else {
		if len(config.Clusters) < 1 {
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

//setConfigDefaults checks passed config and sets default values if needed
func setConfigDefaults(config *Config) {
	if len(config.Policies) < 1 {
		log.Debugf("no policies defined, using default GIT policy source: repo %s, branch %s, directory %s",
			DefaultGitRepository, DefaultGitBranch, DefaultGitPolicyDir)
		config.Policies = append(config.Policies, ConfigPolicy{
			GitRepository: DefaultGitRepository,
			GitBranch:     DefaultGitBranch,
			GitDirectory:  DefaultGitPolicyDir})
	}
}
