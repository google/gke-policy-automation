package app

import (
	"context"
	"fmt"

	"github.com/mikouaj/gke-review/internal/policy"
	cli "github.com/urfave/cli/v2"
)

type Review func(c *Config)

func CreateReviewApp(review Review) *cli.App {
	config := &Config{}
	app := &cli.App{
		Name:  "gke-review",
		Usage: "Review GKE cluster against set of policies",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "project",
				Aliases:     []string{"p"},
				Usage:       "Name of a GCP project",
				Required:    true,
				Destination: &config.ProjectName,
			},
			&cli.StringFlag{
				Name:        "cluster",
				Aliases:     []string{"c"},
				Usage:       "Name of a GKE cluster to review",
				Required:    true,
				Destination: &config.ClusterName,
			},
			&cli.StringFlag{
				Name:        "location",
				Aliases:     []string{"l"},
				Usage:       "GKE cluster location (region or zone)",
				Required:    true,
				Destination: &config.ClusterLocation,
			},
			&cli.BoolFlag{
				Name:        "silent",
				Aliases:     []string{"s"},
				Usage:       "Silent mode",
				Required:    false,
				Destination: &config.SilentMode,
			},
			&cli.StringFlag{
				Name:        "creds",
				Usage:       "Path to GCP JSON credentials file",
				Required:    false,
				Destination: &config.CredentialsFile,
			},
			&cli.StringFlag{
				Name:        "local-policy-dir",
				Usage:       "Local directory with GKE policies",
				Required:    false,
				Destination: &config.LocalDirectory,
			},
			&cli.StringFlag{
				Name:        "git-policy-repo",
				Usage:       "GIT repository with GKE policies",
				Value:       DefaultGitRepository,
				Required:    false,
				Destination: &config.GitRepository,
			},
			&cli.StringFlag{
				Name:        "git-policy-branch",
				Usage:       "Branch name for policies GIT repository",
				Value:       DefaultGitBranch,
				Required:    false,
				DefaultText: DefaultGitBranch,
				Destination: &config.GitBranch,
			},
			&cli.StringFlag{
				Name:        "git-policy-dir",
				Usage:       "Directory name for policies from GIT repository",
				Value:       DefaultGitPolicyDir,
				Required:    false,
				DefaultText: DefaultGitPolicyDir,
				Destination: &config.GitDirectory,
			},
		},
		Action: func(c *cli.Context) error {
			review(config)
			return nil
		},
	}
	return app
}

func GkeReview(c *Config) {
	if err := c.Load(context.Background()); err != nil {
		fmt.Printf("error when loading config: %s", err)
		return
	}
	defer c.Close()
	c.out.Printf(c.out.Color("[white][bold]Fetching GKE cluster details... [projects/%s/locations/%s/clusters/%s]\n"),
		c.ProjectName,
		c.ClusterLocation,
		c.ClusterName)
	cluster, err := c.gke.GetCluster(c.ProjectName, c.ClusterLocation, c.ClusterName)
	if err != nil {
		c.out.ErrorPrint("could not fetch the cluster details", err)
		return
	}
	policySrc := getPolicySource(c)
	if _, ok := policySrc.(*policy.GitPolicySource); ok {
		c.out.Printf(c.out.Color("[white][bold]Reading policy files from GIT repository... [%s branch=%q directory=%q]\n"),
			c.GitRepository,
			c.GitBranch,
			c.GitDirectory)
	}
	files, err := policySrc.GetPolicyFiles()
	if err != nil {
		c.out.ErrorPrint("could not read policy files", err)
		return
	}
	c.out.Printf(c.out.Color("[white][bold]Evaluating REGO policies...\n"))
	pa := policy.NewPolicyAgent(c.ctx, files)
	results, err := pa.EvaluatePolicies(cluster)
	if err != nil {
		c.out.ErrorPrint("could not parse policies", err)
		return
	}
	if len(results.Errored()) > 0 {
		c.out.Printf(c.out.Color("\n[white][bold]Policy parsing errors:\n\n"))
		for _, errored := range results.Errored() {
			c.out.Printf(c.out.Color("[light_yellow][bold]- %s: [reset][yellow]%s\n"), errored.Name, errored.ProcessingErrors[0])
		}
	}

	for _, group := range results.Groups() {
		c.out.Printf(c.out.Color("\n[white][bold]Group %q:\n\n"), group)
		for _, policy := range results.Policies(group) {
			if policy.Valid {
				c.out.Printf(c.out.Color("[bold][green][\u2713] %s: [reset][green]%s\n"), policy.FullName, policy.Description)
			}
			if !policy.Valid {
				c.out.Printf(c.out.Color("[bold][red][x] %s: [reset][red]%s. [bold]Violations:[reset][red] %s\n"), policy.FullName, policy.Description, policy.Violations[0])
			}
		}
	}

	c.out.Printf(c.out.Color("\n[bold][green]Review complete! Policies: %d valid, %d violated, %d errored.\n"),
		results.ValidCount(),
		results.ViolatedCount(),
		results.ErroredCount())
}

func getPolicySource(c *Config) policy.PolicySource {
	return policy.NewGitPolicySource(c.GitRepository,
		c.GitBranch,
		c.GitDirectory)
}
