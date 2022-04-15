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

import cli "github.com/urfave/cli/v2"

type CliConfig struct {
	ConfigFile      string
	SilentMode      bool
	CredentialsFile string
	ClusterName     string
	ClusterLocation string
	ProjectName     string
	GitRepository   string
	GitBranch       string
	GitDirectory    string
	LocalDirectory  string
}

func NewPolicyAutomationCli(p PolicyAutomation) *cli.App {
	app := &cli.App{
		Name:  "gke-policy",
		Usage: "Manage GKE policies",
		Commands: []*cli.Command{
			CreateClusterCommand(p),
			CreateVersionCommand(p),
			CreatePolicyCheckCommand(p),
		},
	}
	return app
}

func CreateClusterCommand(p PolicyAutomation) *cli.Command {
	config := &CliConfig{}
	return &cli.Command{
		Name:  "cluster",
		Usage: "Manage policies against GKE clusters",
		Subcommands: []*cli.Command{
			{
				Name:  "review",
				Usage: "Evaluate policies against given GKE cluster",
				Flags: append([]cli.Flag{
					&cli.BoolFlag{
						Name:        "silent",
						Aliases:     []string{"s"},
						Usage:       "",
						Destination: &config.SilentMode,
					},
					&cli.StringFlag{
						Name:        "config",
						Aliases:     []string{"c"},
						Usage:       "Path to the configuration file",
						Destination: &config.ConfigFile,
					},
					&cli.StringFlag{
						Name:        "creds",
						Usage:       "Path to GCP JSON credentials file",
						Destination: &config.CredentialsFile,
					},
					&cli.StringFlag{
						Name:        "project",
						Aliases:     []string{"p"},
						Usage:       "Name of a GCP project",
						Destination: &config.ProjectName,
					},
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "Name of a GKE cluster to review",
						Destination: &config.ClusterName,
					},
					&cli.StringFlag{
						Name:        "location",
						Aliases:     []string{"l"},
						Usage:       "GKE cluster location (region or zone)",
						Destination: &config.ClusterLocation,
					},
				}, getPolicySourceFlags(config)...),
				Action: func(c *cli.Context) error {
					defer p.Close()
					if err := p.LoadCliConfig(config, ValidateClusterReviewConfig); err != nil {
						cli.ShowSubcommandHelp(c)
						return err
					}
					p.ClusterReview()
					return nil
				},
			},
		},
	}
}

func CreateVersionCommand(p PolicyAutomation) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Shows application version",
		Action: func(c *cli.Context) error {
			defer p.Close()
			if err := p.LoadCliConfig(&CliConfig{}, nil); err != nil {
				cli.ShowSubcommandHelp(c)
				return err
			}
			p.Version()
			return nil
		},
	}
}

func CreatePolicyCheckCommand(p PolicyAutomation) *cli.Command {
	config := &CliConfig{}
	return &cli.Command{
		Name:  "policy",
		Usage: "Manages policy files",
		Subcommands: []*cli.Command{
			{
				Name:  "check",
				Usage: "Validates policy files from defined source",
				Flags: (getPolicySourceFlags(config)),
				Action: func(c *cli.Context) error {
					defer p.Close()
					if err := p.LoadCliConfig(config, ValidatePolicyCheckConfig); err != nil {
						cli.ShowSubcommandHelp(c)
						return err
					}
					p.PolicyCheck()
					return nil
				},
			},
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
