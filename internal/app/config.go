//Copyright 2022 Google LLC
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package app

import (
	"gopkg.in/yaml.v2"
)

const (
	DefaultGitRepository = "https://github.com/mikouaj/gke-review"
	DefaultGitBranch     = "main"
	DefaultGitPolicyDir  = "gke-policies"
)

type ReadFileFn func(string) ([]byte, error)

type ConfigNg struct {
	SilentMode      bool            `yaml:"silent"`
	CredentialsFile string          `yaml:"credentialsFile"`
	Clusters        []ConfigCluster `yaml:"clusters"`
	Policies        []ConfigPolicy  `yaml:"policies"`
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

func ReadConfig(path string, readFn ReadFileFn) (*ConfigNg, error) {
	data, err := readFn(path)
	if err != nil {
		return nil, err
	}
	config := &ConfigNg{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}
