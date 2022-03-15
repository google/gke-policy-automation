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
