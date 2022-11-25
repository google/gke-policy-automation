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
	cfg "github.com/google/gke-policy-automation/internal/config"
	cli "github.com/urfave/cli/v2"
)

type CliConfig struct {
	ConfigFile          string
	SilentMode          bool
	JsonOutput          bool
	K8SCheck            bool
	CredentialsFile     string
	DumpFile            string
	ClusterName         string
	ClusterLocation     string
	ProjectName         string
	GitRepository       string
	GitBranch           string
	GitDirectory        string
	LocalDirectory      string
	OutputFile          string
	DocumentationOutput string
	DiscoveryEnabled    bool
	SccOrgNumber        string
}

func NewPolicyAutomationCli(p PolicyAutomation) *cli.App {
	app := &cli.App{
		Name:  "gke-policy",
		Usage: "Manage GKE policies",
		Commands: []*cli.Command{
			createCheckCommand(p),
			createDumpCommand(p),
			createConfigureCommand(p),
			createVersionCommand(p),
			createGenerateCommand(p),
		},
	}
	return app
}

func createCheckCommand(p PolicyAutomation) *cli.Command {
	config := &CliConfig{}
	return &cli.Command{
		Name:  "check",
		Usage: "Check GKE clusters against best practices",
		Flags: getCheckFlags(config),
		Action: func(c *cli.Context) error {
			defer p.Close()
			if err := p.LoadCliConfig(config, cfg.SetCheckConfigDefaults, cfg.ValidateClusterCheckConfig); err != nil {
				cli.ShowSubcommandHelp(c)
				return err
			}
			return p.Check()
		},
		Subcommands: []*cli.Command{
			{
				Name:  "best-practices",
				Usage: "Check GKE clusters against best practices",
				Flags: getCheckFlags(config),
				Action: func(c *cli.Context) error {
					defer p.Close()
					if err := p.LoadCliConfig(config, cfg.SetCheckConfigDefaults, cfg.ValidateClusterCheckConfig); err != nil {
						cli.ShowSubcommandHelp(c)
						return err
					}
					return p.CheckBestPractices()
				},
			},
			{
				Name:  "scalability",
				Usage: "Check GKE clusters against scalability limits",
				Flags: getCheckFlags(config),
				Action: func(c *cli.Context) error {
					defer p.Close()
					config.K8SCheck = true
					if err := p.LoadCliConfig(config, cfg.SetScalabilityConfigDefaults, cfg.ValidateScalabilityCheckConfig); err != nil {
						cli.ShowSubcommandHelp(c)
						return err
					}
					return p.CheckScalability()
				},
			},
			{
				Name:  "policies",
				Usage: "Validates policy files from the defined source",
				Flags: getCheckFlags(config),
				Action: func(c *cli.Context) error {
					defer p.Close()
					if err := p.LoadCliConfig(config, cfg.SetPolicyConfigDefaults, cfg.ValidatePolicyCheckConfig); err != nil {
						cli.ShowSubcommandHelp(c)
						return err
					}
					return p.PolicyCheck()
				},
			},
		},
	}
}

func createDumpCommand(p PolicyAutomation) *cli.Command {
	config := &CliConfig{}
	return &cli.Command{
		Name:  "dump",
		Usage: "Download and dump data",
		Subcommands: []*cli.Command{
			{
				Name:  "cluster",
				Usage: "Download and dump GKE cluster configuration",
				Flags: getDumpFlags(config),
				Action: func(c *cli.Context) error {
					defer p.Close()
					if err := p.LoadCliConfig(config, cfg.SetCheckConfigDefaults, cfg.ValidateClusterDumpConfig); err != nil {
						cli.ShowSubcommandHelp(c)
						return err
					}
					return p.ClusterJSONData()
				},
			},
		},
	}
}

func createConfigureCommand(p PolicyAutomation) *cli.Command {
	config := &CliConfig{}
	return &cli.Command{
		Name:  "configure",
		Usage: "Configure GKE Policy Automation environment",
		Subcommands: []*cli.Command{
			{
				Name:  "scc",
				Usage: "Configure GKE Policy Automation in Security Command Center",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "organization",
						Aliases:     []string{"o"},
						Usage:       "Organization number",
						Destination: &config.SccOrgNumber,
					},
				},
				Action: func(c *cli.Context) error {
					defer p.Close()
					if err := p.LoadCliConfig(config, nil, nil); err != nil {
						cli.ShowSubcommandHelp(c)
						return err
					}
					return p.ConfigureSCC(config.SccOrgNumber)
				},
			},
		},
	}
}

func createGenerateCommand(p PolicyAutomation) *cli.Command {
	config := &CliConfig{}
	return &cli.Command{
		Name:  "generate",
		Usage: "Generate GKE Policy outputs.",
		Subcommands: []*cli.Command{
			{
				Name:  "policy-docs",
				Usage: "Generate documentation for policy files",
				Flags: (getPolicyDocumentationFlags(config)),
				Action: func(c *cli.Context) error {
					defer p.Close()
					if err := p.LoadCliConfig(config, cfg.SetPolicyConfigDefaults, cfg.ValidateGeneratePolicyDocsConfig); err != nil {
						cli.ShowSubcommandHelp(c)
						return err
					}
					return p.PolicyGenerateDocumentation()
				},
			},
		},
	}
}

func createVersionCommand(p PolicyAutomation) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Shows application version",
		Action: func(c *cli.Context) error {
			defer p.Close()
			if err := p.LoadCliConfig(&CliConfig{}, nil, nil); err != nil {
				cli.ShowSubcommandHelp(c)
				return err
			}
			return p.Version()
		},
	}
}

func getCommonFlags(config *CliConfig) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "Path to the configuration file",
			Destination: &config.ConfigFile,
		},
		&cli.BoolFlag{
			Name:        "silent",
			Aliases:     []string{"s"},
			Usage:       "Disables standard console output",
			Destination: &config.SilentMode,
		},
		&cli.StringFlag{
			Name:        "creds",
			Usage:       "Path to GCP JSON credentials file",
			Destination: &config.CredentialsFile,
		},
	}
}

func getClusterSourceFlags(config *CliConfig) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "discovery",
			Usage:       "Enables cluster discovery on a given project",
			Destination: &config.DiscoveryEnabled,
		},
		&cli.StringFlag{
			Name:        "dump",
			Aliases:     []string{"d"},
			Usage:       "Path to the JSON file with cluster data dump for local checks",
			Destination: &config.DumpFile,
		},
		&cli.StringFlag{
			Name:        "project",
			Aliases:     []string{"p"},
			Usage:       "Name of a GCP project with a GKE cluster to check",
			Destination: &config.ProjectName,
		},
		&cli.StringFlag{
			Name:        "name",
			Aliases:     []string{"n"},
			Usage:       "Name of a GKE cluster to check",
			Destination: &config.ClusterName,
		},
		&cli.StringFlag{
			Name:        "location",
			Aliases:     []string{"l"},
			Usage:       "Location (region or zone) of a GKE cluster to check",
			Destination: &config.ClusterLocation,
		},
	}
}

func getOutputFlags(config *CliConfig) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "out-file",
			Aliases:     []string{"f"},
			Usage:       "Path to the file for storing results",
			Destination: &config.OutputFile,
		},
		&cli.BoolFlag{
			Name:        "json",
			Usage:       "Outputs results to standard console in JSON format",
			Destination: &config.JsonOutput,
		},
	}
}

func getPolicySourceFlags(config *CliConfig) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "local-policy-dir",
			Usage:       "Local directory with GKE policies",
			Destination: &config.LocalDirectory,
		},
		&cli.StringFlag{
			Name:        "git-policy-repo",
			Usage:       "GIT repository with GKE policies",
			Destination: &config.GitRepository,
		},
		&cli.StringFlag{
			Name:        "git-policy-branch",
			Usage:       "Branch name for policies GIT repository",
			Destination: &config.GitBranch,
		},
		&cli.StringFlag{
			Name:        "git-policy-dir",
			Usage:       "Directory name for policies from GIT repository",
			Destination: &config.GitDirectory,
		},
	}
}

func getPolicyDocumentationFlags(config *CliConfig) []cli.Flag {
	flags := getCommonFlags(config)
	flags = append(flags, getPolicySourceFlags(config)...)
	flags = append(flags, getOutputFlags(config)...)
	return flags
}

func getCheckFlags(config *CliConfig) []cli.Flag {
	flags := getCommonFlags(config)
	flags = append(flags, getClusterSourceFlags(config)...)
	flags = append(flags, getPolicySourceFlags(config)...)
	flags = append(flags, getOutputFlags(config)...)
	return flags
}

func getDumpFlags(config *CliConfig) []cli.Flag {
	flags := getCommonFlags(config)
	flags = append(flags, getClusterSourceFlags(config)...)
	flags = append(flags, getOutputFlags(config)...)
	return flags
}
